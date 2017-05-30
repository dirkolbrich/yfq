package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// scrap the necassary json from the yahoo finance page
func getJson(symbol string) (crumb string) {
	if symbol == "" {
		log.Println("No symbol provided")
		return ""
	}

	// basic url
	var url = strings.Replace("https://finance.yahoo.com/quote/symbol/history?p=symbol", "symbol", symbol, -1)

	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	var js string
	// Tokenize html
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			// End of the document, we're done
			return
		// find a start token
		case html.StartTagToken:
			// get tag name of token and find <script> token
			tag, _ := z.TagName()
			if string(tag) == "script" {
				// forward to next token and check if it is a text token
				token := z.Next()
				if token == html.TextToken {
					// select the token
					data := z.Token()
					// match data for reactjs "root.App.main"
					matched, _ := regexp.MatchString("root.App.main", data.Data)
					if matched {
						re := regexp.MustCompile(`root.App.main\s+=\s+(?P<json>\{.*\})`)
						matches := re.FindStringSubmatch(data.Data)
						// map named groups
						groups := make(map[string]string)
						for i, name := range re.SubexpNames() {
							if i != 0 {
								groups[name] = matches[i]
							}
						}
						js = groups["json"]
						return js
					}
				}
			}
		}
	}
	return js
}

// get the crumb cookie from
func getCrumb(raw string) (crumb string) {
	var f interface{}
	err := json.Unmarshal([]byte(raw), &f)
	if err != nil {
		log.Println(err)
	}

	// walk through json map - what a pain in the ar**
	m := f.(map[string]interface{})
	context := m["context"].(map[string]interface{})
	dispatcher := context["dispatcher"].(map[string]interface{})
	stores := dispatcher["stores"].(map[string]interface{})
	crumbStore := stores["CrumbStore"].(map[string]interface{})

	// fmt.Printf("crumb: %v\n", crumbStore["crumb"])
	// for k, v := range crumbStore {
	// 	fmt.Printf("key: %v calue: %v\n", k, v)
	// }

	return crumbStore["crumb"].(string)
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

// var startDate = flag.String("start", "", "start date in format yyyy-mm-dd")
// var endDate = flag.String("end", "", "start date in format yyyy-mm-dd")
// var interval = flag.String("interval", "1d", "param for query type, e.g. d for daily quotes")

func main() {
	flag.Parse()
	fmt.Println(*symbol)

	js := getJson(*symbol)
	crumb := getCrumb(js)
	fmt.Printf("%T %v %s\n", crumb, len(crumb), crumb)

	// s, e := orderDates(*startDate, *endDate)
	// println(s, e)

	// url := Historical(*symbol, *startDate, *endDate, *param)
	// fmt.Println(url)

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
