package main

import (
	"fmt"

	"github.com/dirkolbrich/yfc"
)

func main() {
	// create a new Historical
	historical := yfc.NewHistorical()

	// start a query
	quotes, err := historical.Query("AAPL")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("received quotes for AAPL: %v\n", len(quotes))
	if len(quotes) > 0 {
		fmt.Printf("first: %+v\n", quotes[0])
		fmt.Printf("last:  %+v\n", quotes[len(quotes)-1])
	}

	// set a range for the query
	historical.StartDate = "2017-09-01"
	historical.EndDate = "2010-09-30"

	// retry the query with another symbol
	quotes, err = historical.Query("bas.de")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("received quotes for BAS.DE: %v\n", len(quotes))
	if len(quotes) > 0 {
		fmt.Printf("first: %+v\n", quotes[0])
		fmt.Printf("last:  %+v\n", quotes[len(quotes)-1])
	}
}
