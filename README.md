# Demo Golang Property Price Scraper
This is a quick demo application pulled together by Peter Eldred 06-2018 which scrapes Australian real estate websites for average property prices.

# What does it do?
The appplication serves a single page on **localhost:8000**. The user is prompted to enter a search term of either a suburb or postcode. Upon submission of the search term, the application uses [Ferret](github.com/argusdusty/Ferret) to lookup the most appropriate suburb, state and postcode match for the search criteria.

* The search is submitted to all of the real estate listing sites specified in the config.json file.
* Goroutines are used to scrape each of the listing sites in the shortest possible time
* Multiple searches can be asynchronously executed

# Project Dependancies
* net/http
* github.com/gorilla/mux
* github.com/PuerkitoBio/goquery
* github.com/argusdusty/Ferret
