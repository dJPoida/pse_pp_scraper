# Demo Golang Property Price Scraper
**Author:** Peter Eldred, June 2018

This project is a quick demo application pulled together by Peter Eldred 06-2018 which scrapes Australian real estate websites for average property prices.

![Screenshot](examples/screenshot_1.png)

# What does it do?
The application serves a single page on **localhost:8000**. The user is prompted to enter a search term of either a suburb or postcode. Upon submission of the search term, the application uses [Ferret](github.com/argusdusty/Ferret) to look up the most appropriate suburb, state and postcode match for the search criteria.

* The search is submitted to all of the real estate listing sites specified in the config.json file.
* Goroutines are used to scrape each of the listing sites in the shortest possible time
* Multiple searches can be asynchronously executed

# Project Dependancies
Be sure your go installation has these dependencies  before running:
* net/http
* github.com/gorilla/mux
* github.com/PuerkitoBio/goquery
* github.com/argusdusty/Ferret

# How to run
* **Install:** `go get github.com/dJPoida/Demo-Golang-Property-Price-Scraper`
* **Update:** `go get -u github.com/dJPoida/Demo-Golang-Property-Price-Scraper`
* **Run:** `go run main.go` from within the project path
