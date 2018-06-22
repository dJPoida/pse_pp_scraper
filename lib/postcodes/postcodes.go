//package postcodes handles all of the suburb state and postcode lookup routines
package postcodes

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/argusdusty/Ferret"
)

//SSPStruct contains the suburb state postcode values ready for search
type SSPStruct struct {
	Suburb   string
	State    string
	PostCode string
}

var searchEngine *ferret.InvertedSuffix
var converter = ferret.UnicodeToLowerASCII
var correction = func(b []byte) [][]byte { return ferret.ErrorCorrect(b, ferret.LowercaseLetters) }
var lengthSorter = func(s string, v interface{}, l int, i int) float64 {
	val, ok := v.(uint64)
	sortWeight := -int64(val)

	//fmt.Println(s, sortWeight)

	if ok {
		return float64(sortWeight)
	}
	return -9999
}

//LoadPostcodes opens the postcode datafile specified in the config.json
func LoadPostcodes(fileName string) error {
	t := time.Now()
	Data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}

	Words := make([]string, 0)
	Values := make([]interface{}, 0)

	for _, Vals := range bytes.Split(Data, []byte("\n")) {
		Vals = bytes.TrimSpace(Vals)

		if len(Vals) > 0 {
			// Split on tab.
			currentRow := strings.Split(strings.ToLower(string(Vals)), "\t")
			postCode := currentRow[1]
			suburb := currentRow[2]
			state := currentRow[4]

			//Debug Display all elements.
			//fmt.Println(postCode, suburb, state)

			//Concatenate the three strings to form a single "word"
			suburbStatePostcode := fmt.Sprintf("%s\t%s\t%s", suburb, state, postCode)

			//Use the length of the suburb to sort
			numericPostcode, err := strconv.ParseUint(postCode, 10, 64)
			if err != nil {
				numericPostcode = 0
			}
			sortWeight := (uint64(len(suburb)) * 1000) + numericPostcode

			Words = append(Words, suburbStatePostcode)
			Values = append(Values, sortWeight)
		}
	}

	fmt.Printf("Loaded postcodes in: %f %s\n", time.Since(t).Seconds(), "seconds")

	searchEngine = ferret.New(Words, Words, Values, converter)

	return nil
}

//FindSuburbStatePostcode takes a simple search term and converts it into a suburb state postcode combo
func FindSuburbStatePostcode(searchTerm string) SSPStruct {
	result := SSPStruct{}

	var searchResult []string
	searchResult, _, _ = searchEngine.SortedErrorCorrectingQuery(searchTerm, 1, correction, lengthSorter)

	if len(searchResult) > 0 {
		// Split on Tab.
		splitResult := strings.Split(searchResult[0], "\t")
		result.Suburb = splitResult[0]
		result.State = splitResult[1]
		result.PostCode = splitResult[2]
	}

	return result
}
