package scraper

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"

	"pse_pp_scraper/lib/config"
)

//ScrapeResult holds the data collected from the entire scrape
type ScrapeResult struct {
	SiteCount    int
	PageCount    int
	ElapsedTime  float64
	ListingCount int
	PriceCount   int
	AveragePrice int
	Prices       []int
}

//PageResult holds the data collected from a single scraped page of listings
type PageResult struct {
	ListingCount int
	PriceCount   int
	Prices       []int
}

//pre-compiled regular expressions
var priceRegEx *regexp.Regexp
var numbersRegEx *regexp.Regexp

//getPageURL Prepare a URL for scraping
func getPageURL(siteConfig config.ListingSiteConfig, suburb string, state string, postCode string, pageNo int) string {
	url := siteConfig.SearchURL
	url = strings.Replace(url, "{{suburb}}", strings.Replace(suburb, " ", "-", -1), -1)
	url = strings.Replace(url, "{{state}}", state, -1)
	url = strings.Replace(url, "{{postCode}}", postCode, -1)
	url = strings.Replace(url, "{{pageNo}}", strconv.Itoa(pageNo), -1)
	return url
}

//FetchHTMLPage handles the generics associated with fetching and preparing an HTML page for scapint
func FetchHTMLPage(url string) (*goquery.Document, error) {
	//Debug
	//fmt.Println(url)

	//We have to do this the hard way so we can modify the request and change the user agent
	//to get the results from some of the target listing sites.
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf(`Failed to get target url ("%s"): %s`, url, err)
	}

	//Modify the User Agent (important)
	req.Header.Set("User-Agent", "Golang_Spider_Bot/3.0")
	//ensure the connection closes gracefully
	req.Header.Add("Connection", "close")

	//Perform the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf(`Failed to get target url ("%s"): %s`, url, err)
	}

	//Ensure we close the response after we're finished
	defer resp.Body.Close()

	//Did we get the right status code?
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(`Response status code error while fetching (%s): %d %s`, url, resp.StatusCode, resp.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(`Failed to parse (%s): %s`, url, err)
	}

	return doc, nil
}

//CalculateAveragePrice takes one or more scrapeResult structs and calculates the combined average of all of the scraped prices
func CalculateAveragePrice(scrapeResults ...ScrapeResult) int {

	//TODO: this basic way of calculating the average price should be replaced with a nicer stats
	//package someday as this basic system could overflow if a large number of results maxed out the uint64 ceiling
	var totalPrice uint64
	var priceCount uint64

	for _, sr := range scrapeResults {
		priceCount += uint64(sr.PriceCount)
		for _, price := range sr.Prices {
			totalPrice += uint64(price)
		}
	}

	if priceCount > 0 {
		return int(totalPrice / priceCount)
	} else {
		return 0
	}
}

//scrapeListingPage scrapes a single page for all of the listings and prices.
//If a page has already been loaded and parsed with qoquery, it can be passed into the existingPage parameter
func scrapeListingPage(c chan PageResult, wg *sync.WaitGroup, doc *goquery.Document, siteConfig config.ListingSiteConfig, suburb string, state string, postcode string, pageNumber int) {
	//Make sure we decrement the wait group stack
	defer wg.Done()

	result := PageResult{Prices: make([]int, 0)}

	//Do we need to fetch the HTML page?
	var err error
	if doc == nil {
		doc, err = FetchHTMLPage(getPageURL(siteConfig, suburb, state, postcode, pageNumber))
		if err != nil {
			//Send the empty result to the channel and bail
			c <- result
			return
		}
	}

	//Debug
	//fmt.Println("Page", pageNumber)

	//Find all of the listing cards in the doc
	doc.Find(siteConfig.ListingSelector).Each(func(index int, item *goquery.Selection) {
		result.ListingCount++
		priceTag := item.Find(siteConfig.PriceSelector)
		priceText := priceTag.Text()
		match := priceRegEx.FindStringSubmatch(priceText)
		var price int
		if len(match) > 0 {
			price, _ = strconv.Atoi(numbersRegEx.ReplaceAllString(match[0], ""))
			//Ignore wild prices
			if price >= 10000 && price <= 999999999 {
				result.Prices = append(result.Prices, price)
				result.PriceCount++
			}
		}

		//Debug
		//fmt.Println(index, priceText, price)
	})

	//Send the result to the channel
	c <- result
}

//scrapeSite is responsible for scraping a single site
func scrapeSite(c chan ScrapeResult, siteWg *sync.WaitGroup, siteConfig config.ListingSiteConfig, maxPages int, suburb string, state string, postcode string) {
	defer siteWg.Done()
	result := ScrapeResult{Prices: make([]int, 0)}

	//Fetch the first page of results
	doc, err := FetchHTMLPage(getPageURL(siteConfig, suburb, state, postcode, 1))
	if err != nil {
		c <- result
		fmt.Println("Error fetching the HTML Page: ", err)
		return
	}

	//Find the pager and determine the number of pages found for this query
	//TODO: This doesn't work for all sites. We'll need varying methods
	doc.Find(siteConfig.PageLinkSelector).Each(func(index int, item *goquery.Selection) {
		pageNum, _ := strconv.Atoi(item.Text())
		if result.PageCount < pageNum {
			result.PageCount = pageNum
		}
		if result.PageCount > maxPages {
			result.PageCount = maxPages
		}
	})

	//Debug
	fmt.Printf("%d pages found for %s\n", result.PageCount, siteConfig.Name)

	//Are there any listings / pages?
	if result.PageCount > 0 {
		//Waitgroup for all page scrape operations
		var pageWg sync.WaitGroup

		//Instantiate the channel with the number of pages we're going to scrape
		queue := make(chan PageResult, result.PageCount)

		//We've already loaded and parsed the first page - scrape this page without loading it again
		pageWg.Add(1)
		go scrapeListingPage(queue, &pageWg, doc, siteConfig, suburb, state, postcode, 1)

		//If there is more than one page, cycle through the remaining pages and scrape each one
		for pageNum := 2; pageNum <= result.PageCount; pageNum++ {
			pageWg.Add(1)
			go scrapeListingPage(queue, &pageWg, nil, siteConfig, suburb, state, postcode, pageNum)
		}

		//Wait for each page to scrape
		pageWg.Wait()
		close(queue)

		//Consolidate the results
		for pageResult := range queue {
			result.ListingCount += pageResult.ListingCount
			result.PriceCount += pageResult.PriceCount
			result.Prices = append(result.Prices, pageResult.Prices...)
		}
	}

	//Debug
	//fmt.Println(result)
	result.AveragePrice = CalculateAveragePrice(result)
	c <- result
}

//ScrapeListings is the entry point for the application scraping listings from one of the property sites
func ScrapeListings(c chan ScrapeResult, sites []config.ListingSiteConfig, maxPages int, suburb string, state string, postcode string) {
	operationStart := time.Now()
	result := ScrapeResult{Prices: make([]int, 0)}

	fmt.Printf("Search received for %s, %s %s\n", suburb, state, postcode)

	//Waitgroup for all site scrape operations
	var siteWg sync.WaitGroup

	//Instantiate the channel with the number of sites we're going to scrape
	siteQueue := make(chan ScrapeResult, len(sites))

	//Iterate over the sites to scrape
	for _, site := range sites {
		if site.Enabled {
			siteWg.Add(1)
			go scrapeSite(siteQueue, &siteWg, site, maxPages, suburb, state, postcode)
		}
	}

	//Wait for each site to scrape
	//TODO: switch to select and add a timeout to ensure we avoid deadlocking
	siteWg.Wait()
	close(siteQueue)

	//Consolidate the results
	for siteResult := range siteQueue {
		result.SiteCount++
		result.ListingCount += siteResult.ListingCount
		result.PriceCount += siteResult.PriceCount
		result.PageCount += siteResult.PageCount
		result.Prices = append(result.Prices, siteResult.Prices...)
	}

	result.ElapsedTime = time.Since(operationStart).Seconds()
	result.AveragePrice = CalculateAveragePrice(result)

	c <- result
}

func init() {
	//Compile the regular expressions required for this package
	//very simple match for the moment - needs to be improved to handle other forms of digits like 1.2M and 600K)
	priceRegEx = regexp.MustCompile(`\$*\d{1,3}\,*\d{1,3}\,*\d{1,3}`)
	//simple regex to strip non-numerics
	numbersRegEx = regexp.MustCompile("[^a-zA-Z0-9]+")
}
