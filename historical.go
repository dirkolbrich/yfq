package yfq

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// CRUMBURL is the URL to the Yahoo Finance webpage of the given symbol to scrap for the cookie and symbol
	CRUMBURL = "https://finance.yahoo.com/quote/{symbol}/history?p={symbol}"
	// HISTORYURL is the basic URL to download historic quotes
	HISTORYURL = "https://query1.finance.yahoo.com/v7/finance/download/{symbol}?"
	// CONFIGURL configurates the query to HISTORYURL
	CONFIGURL = "period1={start}&period2={end}&interval=1d&events=history&crumb={crumb}"
)

// Historical represents the basic Query for historical quotes
type Historical struct {
	StartDate string
	EndDate   string
	Quotes    []Quote

	crumbURL string
	crumb    string
	cookies  []*http.Cookie
}

// Quote represents a single quote
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

// NewHistorical return a Historical struct ready to use.
func NewHistorical() *Historical {
	return &Historical{}
}

// Query returns the parsed data string as a slice of []Quote
func (h *Historical) Query(symbol string) (quotes []Quote, err error) {
	data, err := h.baseQuery(symbol)

	quotes = parseHistoricalCSV(symbol, data)

	return quotes, err
}

// QueryRaw returns the raw csv string
func (h *Historical) QueryRaw(symbol string) (data [][]string, err error) {
	data, err = h.baseQuery(symbol)

	return data, nil
}

// ResetDates rests the given start end end dates to nil.
func (h *Historical) ResetDates() error {
	h.StartDate = ""
	h.EndDate = ""

	return nil
}

// RenewCrumb renews the crumb and cookie to be used in the main query.
func (h *Historical) RenewCrumb() error {
	crumb, cookies, err := getCrumb(h.crumbURL)
	if err != nil {
		return err
	}
	h.crumb = crumb
	h.cookies = cookies

	return nil
}

// baseQuery fetches the latest historical quotes from finance.yahoo.com
func (h *Historical) baseQuery(symbol string) (data [][]string, err error) {
	// validate symbol
	if len(symbol) == 0 {
		err := errors.New("No symbol provided")
		return data, err
	}

	// build or update the crumb
	err = h.buildCrumb(symbol)
	if err != nil {
		return data, err
	}

	// validate order of dates
	h.StartDate, h.EndDate, err = orderDates(h.StartDate, h.EndDate)
	if err != nil {
		return data, err
	}
	// parse dates
	start, end, err := parseDates(h.StartDate, h.EndDate)
	if err != nil {
		return data, err
	}

	// build url for the main historical query
	historyURL := strings.Replace(HISTORYURL, "{symbol}", symbol, -1)
	configURL := strings.Replace(CONFIGURL, "{start}", start, -1)
	configURL = strings.Replace(configURL, "{end}", end, -1)
	configURL = strings.Replace(configURL, "{crumb}", h.crumb, -1)

	queryURL := historyURL + configURL

	// query for csv
	data, err = readCSVFromURL(queryURL, h.cookies)
	if err != nil {
		err := fmt.Errorf("could not establish new csv request: %v", err)
		return data, err
	}

	// whats the csv like?
	// fmt.Println(len(data))
	// if len(data) > 0 {
	// 	fmt.Println(data[0])
	// 	fmt.Println(data[len(data)-1])
	// 	fmt.Printf("%#v\n", data[len(data)-1])
	// }

	return data, nil
}

// readCSVFromURL fetches the csv file from the provided URL.
// Is uses the provided cookies for the request.
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
	// add cookies to the request
	for _, c := range cookies {
		req.AddCookie(c)
	}

	// send request
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

// parseHistoricalCSV parses the returned CSV into single Quotes and adds these to []Quote.
// It also adds the symbol to each row.
func parseHistoricalCSV(symbol string, csv [][]string) (quotes []Quote) {
	var csvHeader []string
	// parse csv to map
	for id, row := range csv {
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

// buildCrumb build the necassary url and sets the crumb and cookies
func (h *Historical) buildCrumb(symbol string) error {
	// check for available crumb and cookies and query these first if needed
	// fmt.Printf("before\nh.crumb: %#v\nh.cookies: %#v\n", h.crumb, h.cookies)
	if (h.crumb == "") || (h.cookies == nil) {
		// build url for retrieving the crumb
		crumbURL, err := buildCrumbURL(symbol)
		if err != nil {
			return err
		}
		h.crumbURL = crumbURL

		crumb, cookies, err := getCrumb(crumbURL)
		if err != nil {
			return err
		}
		h.crumb = crumb
		h.cookies = cookies
	}
	// fmt.Printf("after\nh.crumb: %#v\nh.cookies: %#v\n", h.crumb, h.cookies)
	return nil
}

// buildCrumbURL builds the URL to request the crumb and cookies
func buildCrumbURL(symbol string) (url string, err error) {
	if symbol == "" {
		err := errors.New("could not build crumb URL, empty string given")
		return url, err
	}

	// build url for retrieving the crumb
	url = strings.Replace(CRUMBURL, "{symbol}", symbol, -1)

	return url, nil
}

// getCrumb scraps the neccessary json and cookie from the yahoo finance page
// and returns the crumb string with the cookies.
func getCrumb(url string) (crumb string, cookies []*http.Cookie, err error) {
	// set client with a timeout
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	// define request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err := fmt.Errorf("could not establish new crumb request: %v", err)
		return crumb, cookies, err
	}
	// fmt.Printf("url:%#v\n req: %#v\n", req.URL, req)

	// issue request
	resp, err := client.Do(req)
	if err != nil {
		err := fmt.Errorf("could not query page: %v", err)
		return crumb, cookies, err
	}
	defer resp.Body.Close()
	// fmt.Printf("resp: %#v\n", resp)

	// collect the received cookies
	cookies = resp.Cookies()

	// read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("could not read response body: %v", err)
		return crumb, cookies, err
	}

	// search for crumb in body
	crumb, err = parseCrumb(string(body))
	if err != nil {
		err := fmt.Errorf("could not parse response body for crumb: %v", err)
		return crumb, cookies, err
	}

	return crumb, cookies, nil
}

// parseCrumb searches for a crumb within a string.
func parseCrumb(s string) (crumb string, err error) {
	// search for crumb in body
	re := regexp.MustCompile(`(?P<CrumbStore>"CrumbStore"\s?:\s?{"crumb"\s?:\s?"(?P<crumb>.*?)"})`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		err := errors.New("could not find crumb")
		return crumb, err
	}

	// rearrange submatches to map
	matchMap := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 {
			matchMap[name] = matches[i]
		}
	}
	crumb = matchMap["crumb"]

	return crumb, nil
}

// orderDates validates the correct order of start to end date.
// Start must be earlier than end. If necassary, the method reorders the dates.
func orderDates(start, end string) (s, e string, err error) {
	s = start
	e = end

	if (start > end) && ((start != "") && (end != "")) {
		tmp := start
		s = end
		e = tmp
	}

	return s, e, err
}

// parseDates parses the start and end date and converts these to an UNIX time string.
// Start and end must be a date formated according to ISO 8601: yyyy-mm-dd or an empty string.
// There are three different cases for the start and end date:
// - start date empty, valid or invalid
// - end date empty, valid or invalid
// If both start and end date are valid, the date range will be parsed as specified.
// If both start and end are empty or invalid, the date range is set to max start (unix 0) and end to unix now.
// If start date is empty/invalid and end date is valid, date range is set to max start (unix 0) and end to specified date.
// If start date is valid and end date is empty/invalid, date range is set to start date and end date to unix now.
func parseDates(start, end string) (s, e string, err error) {
	if len(start) == 0 {
		// set to min
		s = "0"
	} else {
		s, err = parseDateStringToUnix(start)
		if err != nil {
			s = "0"
		}
	}

	if len(end) == 0 {
		// set to min
		e = strconv.Itoa(int(time.Now().Unix()))
	} else {
		e, err = parseDateStringToUnix(end)
		if err != nil {
			e = strconv.Itoa(int(time.Now().Unix()))
		}
	}

	return s, e, nil
}

// parseDate parses a single date string to an unix string.
func parseDateStringToUnix(s string) (unix string, err error) {
	date, err := time.Parse("2006-01-02", s)
	if err != nil {
		err = fmt.Errorf("Could not parse string \"%s\" to time: %v", s, err)
		return s, err
	}
	unix = strconv.Itoa(int(date.Unix()))

	return unix, nil
}
