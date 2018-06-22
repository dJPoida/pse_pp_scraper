//package main
//This is a quick demo application pulled together by Peter Eldred 06-2018
//
//It exposes a single API method which receives a search criteria to scrape
//several Australian realestate websites. It then compiles a list of average
//prices for the search criteria and returns a consolidated result.
//
//This program demonstrates some basic strengths that golang has to offer,
//mainly around concurrency.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"pse_pp_scraper/lib/config"
	"pse_pp_scraper/lib/postcodes"
	"pse_pp_scraper/lib/scraper"
)

//Global variable for holding the application configuration
var appConfig config.Config

//SearchFormData is the struct received from the search form
type SearchFormData struct {
	SearchID   int    `json:"searchId"`
	SearchText string `json:"searchText"`
}

//SearchResponseData is the struct sent back in response to a search
type SearchResponseData struct {
	Request      SearchFormData `json:"request"`
	Suburb       string         `json:"suburb"`
	State        string         `json:"state"`
	PostCode     string         `json:"postCode"`
	AvgPrice     int            `json:"avgPrice"`
	PriceCount   int            `json:"priceCount"`
	ListingCount int            `json:"listingCount"`
	SiteCount    int            `json:"siteCount"`
	PageCount    int            `json:"pageCount"`
	ElapsedTime  float64        `json:"elapsedTime"`
}

//scrapeHandler takes post ajax requests from the client javascript and executes the scrape task
func searchHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var formData SearchFormData
	if err := json.NewDecoder(r.Body).Decode(&formData); err != nil {
		fmt.Println("ERROR:", err)
		http.Error(w, "Bad Request", http.StatusTeapot)
		return
	}

	//Parse the search text and determine the best suburb / state / postcode combo
	sspStruct := postcodes.FindSuburbStatePostcode(formData.SearchText)

	//TODO: Scrape multiple listing sites
	scrapeResultChan := make(chan scraper.ScrapeResult)
	go scraper.ScrapeListings(scrapeResultChan, appConfig.ScraperConfig.Sites, appConfig.ScraperConfig.MaxPages, sspStruct.Suburb, sspStruct.State, sspStruct.PostCode)
	scrapeResult := <-scrapeResultChan

	response := SearchResponseData{Request: formData,
		AvgPrice:     scrapeResult.AveragePrice,
		PriceCount:   scrapeResult.PriceCount,
		ListingCount: scrapeResult.ListingCount,
		PageCount:    scrapeResult.PageCount,
		SiteCount:    scrapeResult.SiteCount,
		ElapsedTime:  scrapeResult.ElapsedTime,
		Suburb:       sspStruct.Suburb,
		State:        sspStruct.State,
		PostCode:     sspStruct.PostCode}

	json.NewEncoder(w).Encode(response)
}

func init() {
	//Load the Configuration
	var err error
	appConfig, err = config.LoadConfiguration("config.json")
	if err != nil {
		panic(fmt.Errorf("Failed to load config. %s", err))
	}

	//Load the Postcodes
	err = postcodes.LoadPostcodes(appConfig.PostCodeFileName)
	if err != nil {
		panic(fmt.Errorf("Failed to load post code database. %s", err))
	}
}

//main is the main function in the application. When this function exits, the program ends.
func main() {
	//Define the Routes
	router := mux.NewRouter().StrictSlash(true)

	//API routes
	router.HandleFunc("/search", searchHandler).Methods("POST")

	//web and asset routes
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	http.Handle("/", router)

	//Listen for incoming HTTP requests
	fmt.Println("Listening on localhost:", appConfig.ServerPort)
	serverPort := fmt.Sprintf(":%d", appConfig.ServerPort)
	httpError := http.ListenAndServe(serverPort, router)
	if httpError != nil {
		panic(fmt.Errorf("Failed to serve index page: %s", httpError))
	}
}
