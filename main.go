package main

import (
	"encoding/csv"
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

var (
	httpClient *http.Client
)

// Historical fetches the latest historical quotes from finance.yahoo.com
func Historical(symbol, start, end, interval, event string) string {
	var crumbURL = "https://finance.yahoo.com/quote/{symbol}/history?p={symbol}"
	var historyURL = "https://query1.finance.yahoo.com/v7/finance/download/{symbol}?"
	var configURL = "period1={start}&period2={end}&interval={interval}&events={event}&crumb={crumb}"

	// validate symbol
	if symbol == "" {
		log.Println("no symbol provided")
		return ""
	}

	// query crumb url
	url := strings.Replace(crumbURL, "{symbol}", symbol, -1)
	crumb, _ := getCrumb(url)

	start, end = orderDates(start, end)
	start, end = parseDates(start, end)

	// modify configURL
	historyURL = strings.Replace(historyURL, "{symbol}", symbol, -1)
	configURL = strings.Replace(configURL, "{start}", start, -1)
	configURL = strings.Replace(configURL, "{end}", end, -1)
	configURL = strings.Replace(configURL, "{interval}", interval, -1)
	configURL = strings.Replace(configURL, "{event}", event, -1)
	configURL = strings.Replace(configURL, "{crumb}", crumb, -1)

	queryURL := historyURL + configURL
	data, _ := readCSVFromUrl(queryURL)
	fmt.Printf("%T %#v\n", data, data)

	for id, row := range data {
		if id == 10 {
			break
		}
		fmt.Printf("id:%v %T %q\n", id, row, row)
	}

	return historyURL + configURL
}

func readCSVFromUrl(url string) ([][]string, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	reader := csv.NewReader(r.Body)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// getCrumb scraps the neccessary json anf cookie from the yahoo finance page and returns the single crumb string
func getCrumb(url string) (crumb string, cookies []*http.Cookie) {
	// fetch yahoo finance for basic page
	r, err := http.Get(url)
	if err != nil {
		log.Println("Could not query page", err)
	}

	defer r.Body.Close()
	fmt.Printf("%T %v\n", r.Body, r.Body)
	for _, c := range r.Cookies() {
		fmt.Printf("%T %#v\n", c, c)
	}
	// collect the send cookies
	cookies = r.Cookies()

	// Tokenize html
	z := html.NewTokenizer(r.Body)

	var js string
L:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			// End of the document, we're done
			break
		// find a start token
		case html.StartTagToken:
			// get tag name of token and find <script> token
			tag, _ := z.TagName()
			if string(tag) == "script" {
				// script tag found, forward to next token and check if it is a text token
				token := z.Next()
				// token is text, not EndTagToken
				if token == html.TextToken {
					data := parseToken("root.App.main", z.Token())
					if len(data) > 0 {
						js = data
						break L
					}
				}
			}
		}
	}

	// search the data string for CrumbStore
	j := regexText(`(?P<CrumbStore>"CrumbStore"\s?:\s?(?P<crumb>{"crumb"\s?:\s?".*?"}))`, js)
	// select the CrumbStore group
	crumbStore := j["crumb"]

	// parsing json
	var f interface{}
	err = json.Unmarshal([]byte(crumbStore), &f)
	if err != nil {
		log.Println("invalid json", err)
	}

	// walk through json map
	m := f.(map[string]interface{})
	crumb = m["crumb"].(string)

	return crumb, cookies
}

// parseToken parses a html.Token if it contains a string
func parseToken(s string, t html.Token) string {
	var data string
	// match data for string
	matched, _ := regexp.MatchString(s, t.Data)
	if matched {
		// found string in Data
		data = t.Data
	}

	return data
}

// regexText searches the given string for the regular expression, returns a map
func regexText(regex string, s string) map[string]string {
	re := regexp.MustCompile(regex)
	matches := re.FindStringSubmatch(s)

	// rearrange submatches to map
	m := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 {
			m[name] = matches[i]
		}
	}

	return m
}

// parseDates parses start and end date to UNIX time string
func parseDates(start, end string) (s, e string) {
	start, end = orderDates(start, end)

	if len(start) == 0 {
		// set to min
		s = "0"
	} else {
		date, _ := time.Parse("2006-01-02", start)
		s = strconv.Itoa(int(date.Unix()))
	}

	if len(end) == 0 {
		// set to max
		e = "9999999999"
	} else {
		date, _ := time.Parse("2006-01-02", end)
		e = strconv.Itoa(int(date.Unix()))
	}

	return s, e
}

// orderDates validates the correct order of start to end - start must be earlier than end
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
var start = flag.String("start", "", "start date in format yyyy-mm-dd")
var end = flag.String("end", "", "start date in format yyyy-mm-dd")
var interval = flag.String("interval", "1d", "param for query interval, e.g. d for daily, weekly or monthly")
var event = flag.String("event", "history", "param for query type, e.g. d for history, div or split")

func main() {
	flag.Parse()
	fmt.Println(*symbol, *start, *end)

	url := Historical(*symbol, *start, *end, *interval, *event)
	fmt.Printf("%q", url)
}
