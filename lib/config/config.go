package config

import (
	"encoding/json"
	"os"
)

//Config stores the configuration as described in  config.json
type Config struct {

	//ServerPort is the port the server will listen on
	ServerPort uint32 `json:"serverPort"`

	//PostCodeFileName is the relative path to the datafile used for searching suburbs and postcodes
	PostCodeFileName string `json:"postCodeFileName"`

	//Scraper Config
	ScraperConfig struct {

		//MaxPages is the maximum number of pages we will allow the scraper to traverse through. Min = 1
		MaxPages int `json:"maxPages"`

		//Sites contains the specific configurations for scraping each listing site
		Sites []ListingSiteConfig `json:"sites"`
	} `json:"scraperConfig"`
}

//ListingSiteConfig holds the specific details for a listing site which are used by the generic scraper
type ListingSiteConfig struct {
	Name             string `json:"name"`
	Enabled          bool   `json:"enabled"`
	SearchURL        string `json:"searchUrl"`
	PageLinkSelector string `json:"pageLinkSelector"`
	ListingSelector  string `json:"listingSelector"`
	PriceSelector    string `json:"priceSelector"`
}

//LoadConfiguration retrieves the json object and parses it into the Config struct
func LoadConfiguration(filename string) (Config, error) {
	configFile, fileError := os.Open(filename)
	var parsedConfig Config

	defer configFile.Close()

	if fileError != nil {
		return parsedConfig, fileError
	}

	jsonParser := json.NewDecoder(configFile)
	parseError := jsonParser.Decode(&parsedConfig)
	return parsedConfig, parseError
}
