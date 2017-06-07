package yahoofinancequery

import (
	"encoding/csv"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Quote is the basic struc representing a single quote
type Quote struct {
	Date     time.Time
	Symbol   string
	Open     float64
	High     float64
	Low      float64
	Close    float64
	AdjClose float64
	Volume   int64
}

// Historical fetches the latest historical quotes from finance.yahoo.com
func Historical(symbol, start, end string) (quotes []Quote, err error) {
	var crumbURL = "https://finance.yahoo.com/quote/{symbol}/history?p={symbol}"
	var historyURL = "https://query1.finance.yahoo.com/v7/finance/download/{symbol}?"
	var configURL = "period1={start}&period2={end}&interval=1d&events=history&crumb={crumb}"

	// validate symbol
	if len(symbol) == 0 {
		err := errors.New("No symbol provided")
		return quotes, err
	}

	// query crumb url
	url := strings.Replace(crumbURL, "{symbol}", symbol, -1)
	crumb, cookies := getCrumb(url)

	start, end = orderDates(start, end)
	start, end, err = parseDates(start, end)
	if err != nil {
		return quotes, err
	}

	// modify configURL
	historyURL = strings.Replace(historyURL, "{symbol}", symbol, -1)
	configURL = strings.Replace(configURL, "{start}", start, -1)
	configURL = strings.Replace(configURL, "{end}", end, -1)
	configURL = strings.Replace(configURL, "{crumb}", crumb, -1)

	queryURL := historyURL + configURL

	// query for csv
	data, err := readCSVFromURL(queryURL, cookies)
	if err != nil {
		err := errors.New("Could not establish new request: ")
		return quotes, err
	}

	quotes = parseHistoricalCSV(symbol, data)

	return quotes, nil
}

// readCSVFromURL fetches the  csv file from the provided url
// uses the provides cookies
func readCSVFromURL(url string, cookies []*http.Cookie) ([][]string, error) {
	// set client
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// define request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// issue rewuest
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read csv from response
	reader := csv.NewReader(resp.Body)
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}

// parseHistoricalCSV parse the returned csv into single Quotes and adds these to []Quote
func parseHistoricalCSV(symbol string, data [][]string) (quotes []Quote) {
	var csvHeader []string
	// parse csv to map
	for id, row := range data {
		if id == 0 {
			csvHeader = row
			continue
		}
		// rearange row to header
		quote := make(map[string]string)
		for i, v := range row {
			quote[csvHeader[i]] = v
		}

		// create new Quote and populate
		q := Quote{}
		q.Date, _ = time.Parse("2006-01-02", quote["Date"])
		q.Symbol = strings.ToUpper(symbol)
		q.Open, _ = strconv.ParseFloat(quote["Open"], 64)
		q.High, _ = strconv.ParseFloat(quote["High"], 64)
		q.Low, _ = strconv.ParseFloat(quote["Low"], 64)
		q.Close, _ = strconv.ParseFloat(quote["Close"], 64)
		q.AdjClose, _ = strconv.ParseFloat(quote["Adj Close"], 64)
		q.Volume, _ = strconv.ParseInt(quote["Volume"], 10, 64)

		quotes = append(quotes, q)
	}

	return
}

// getCrumb scraps the neccessary json and cookie from the yahoo finance page
// and returns the crumb string with the cookies
func getCrumb(url string) (crumb string, cookies []*http.Cookie) {
	// set client
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// define request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Could not establish new request: ", err)
	}
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Could not query page: ", err)
	}
	defer resp.Body.Close()

	// collect the send cookies
	cookies = resp.Cookies()

	// read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Could not read response body: ", err)
	}

	// search for crumb in body
	re := regexp.MustCompile(`(?P<CrumbStore>"CrumbStore"\s?:\s?{"crumb"\s?:\s?"(?P<crumb>.*?)"})`)
	matches := re.FindStringSubmatch(string(body))
	if matches == nil {
		log.Println("Could not find crumb")
	}

	// rearrange submatches to map
	matchMap := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 {
			matchMap[name] = matches[i]
		}
	}
	crumb = matchMap["crumb"]

	return crumb, cookies
}

// parseDates parses start and end date and converts to UNIX time string
func parseDates(start, end string) (s, e string, err error) {
	start, end = orderDates(start, end)

	if len(start) == 0 {
		// set to min
		s = "0"
	} else {
		date, err := time.Parse("2006-01-02", start)
		if err != nil {
			log.Fatal("Could not parse string to time: ", err)
		}
		s = strconv.Itoa(int(date.Unix()))
	}

	if len(end) == 0 {
		// set to max
		e = "9999999999"
	} else {
		date, err := time.Parse("2006-01-02", end)
		if err != nil {
			log.Fatal("Could not parse string to time: ", err)
		}
		e = strconv.Itoa(int(date.Unix()))
	}

	return s, e, nil
}

// orderDates validates the correct order of start to end
// start must be earlier than end
func orderDates(start, end string) (string, string) {
	if (start > end) && (end != "") {
		tmp := start
		start = end
		end = tmp
	}

	return start, end
}