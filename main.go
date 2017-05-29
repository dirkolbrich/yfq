package main

import (
	"flag"
	"fmt"
	"strconv"
	"time"
)

var historicalParams = map[string]string{
	"d": "daily",
	"w": "weekly",
	"m": "monthly",
	"v": "dividends",
}

func Historical(symbol, start, end, param string) string {

	var baseURL = "http://ichart.finance.yahoo.com/table.csv?s="
	var configValue = "&ignore=.csv"

	start, end = orderDates(start, end)
	var dateValue = parseDates(start, end)

	paramValue := "&g=d"

	queryURL := baseURL + symbol + dateValue + paramValue + configValue

	return queryURL
}

// parse start and end date for query
func parseDates(start, end string) string {
	start, end = orderDates(start, end)

	// yahoo fincnae formats the query with
	// a: start month
	// b: start day
	// c: start year
	// d: end month
	// e: end day
	// f: end year
	var startDate, endDate string

	if start == "" {
		startDate = "&a=&b=&c="
	} else {
		date, _ := time.Parse("2006-01-02", start)
		// yahoo finance index for month starts with 0
		sMonth := "&a=" + strconv.Itoa(int(date.Month())-1)
		sDay := "&b=" + strconv.Itoa(date.Day())
		sYear := "&c=" + strconv.Itoa(date.Year())
		startDate = sMonth + sDay + sYear
		fmt.Println(startDate)
	}

	if end == "" {
		endDate = "&d=&e=&f="
	} else {
		date, _ := time.Parse("2006-01-02", end)
		// yahoo finance index for month starts with 0
		eMonth := "&d=" + strconv.Itoa(int(date.Month())-1)
		eDay := "&e=" + strconv.Itoa(date.Day())
		eYear := "&f=" + strconv.Itoa(date.Year())
		endDate = eMonth + eDay + eYear
		fmt.Println(endDate)
	}

	return startDate + endDate
}

// order dates - start must be earlier than end
func orderDates(start, end string) (string, string) {
	if (start > end) && (end != "") {
		tmp := start
		start = end
		end = tmp
	}

	return start, end
}

// define symbol flags
var symbol = flag.String("symbol", "", "stock symbol e.g. bas.de")
var startDate = flag.String("start", "", "start date in format yyyy-mm-dd")
var endDate = flag.String("end", "", "start date in format yyyy-mm-dd")
var param = flag.String("p", "d", "param for query type, e.g. d for daily quotes")

func main() {
	flag.Parse()
	fmt.Println(*symbol, *startDate, *endDate, *param)

	// s, e := orderDates(*startDate, *endDate)
	// println(s, e)

	url := Historical(*symbol, *startDate, *endDate, *param)
	fmt.Println(url)

	// url := "http://example.org/"
	// resp, err := http.Get(url)
	// if err != nil {
	// 	// handle error
	// }
	// defer resp.Body.Close()

	// body, err := ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	// handle error
	// }
	// fmt.Printf("%s", body)
}
