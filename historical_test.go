package yfc

import (
	"errors"
	// "net/http"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestBuildCrumbURL(t *testing.T) {
	// test url generated from a given symbol string and the base crumb url
	var testCases = []struct {
		m      string // test message
		symbol string // symbol to build crumb from
		expURL string // expected crumb url string
		expErr error  // expected error output
	}{
		{"testing standard symbol string", "AAPL", "https://finance.yahoo.com/quote/AAPL/history?p=AAPL", nil},
		{"testing empty symbol string", "", "", errors.New("error")},
	}

	for _, tc := range testCases {
		url, err := buildCrumbURL(tc.symbol)
		if url != tc.expURL || (reflect.TypeOf(err) != reflect.TypeOf(tc.expErr)) {
			t.Errorf("\n%v: buildCrumbURL(%v): \nexpected %v %v, \n  actual %v %v",
				tc.m, tc.symbol, tc.expURL, tc.expErr, url, err)
		}
	}
}

func TestGetCrumb(t *testing.T) {
	// test html generated from finance.yahoo.com
	var testStringLong = `<!DOCTYPE html><html><head><title>title</title></head><body><div id="app"><p>test paragraph</p></div><script>(function (root) {/* -- Data -- */root.App || (root.App = {}); root.App.main = {"context": {"dispatcher": {"stores": {"FlyoutStore": {}, "CrumbStore": {"crumb": "crumb.Test"}, "HeaderStore": {}}}, "options": {}}, "plugins": {}};}(this)); </script></body></html>`
	var testStringShort = `"CrumbStore": {"crumb": "crumb.Test"}`
	var testStringFailure = `<!DOCTYPE html><html><head></head><body></body></html>`

	var testCases = []struct {
		m        string // test message
		url      string // rssponse body to parse
		expCrumb string // expected crumb string
		expErr   error  // expected error output
	}{
		{"testing long string", testStringLong, "crumb.Test", nil},
		{"testing short string", testStringShort, "crumb.Test", nil},
		{"testing failure string", testStringFailure, "", errors.New("error")},
		{"testing empty string", "", "", errors.New("error")},
	}

	for _, tc := range testCases {
		crumb, err := parseCrumb(tc.url)
		if crumb != tc.expCrumb || (reflect.TypeOf(err) != reflect.TypeOf(tc.expErr)) {
			t.Errorf("\n%v: getCrumb(%v): \nexpected %v %v, \n  actual %v %v",
				tc.m, tc.url, tc.expCrumb, tc.expErr, crumb, err)
		}
	}
}

func TestOrderDates(t *testing.T) {
	var testCases = []struct {
		m      string // test message
		s      string // start input
		e      string // end input
		expS   string // expected start output
		expE   string // expected end output
		expErr error  // expected error output
	}{
		{"testing no input", "", "", "", "", nil},
		{"testing only start date", "2017-06-07", "", "2017-06-07", "", nil},
		{"testing only end date", "", "2017-06-07", "", "2017-06-07", nil},
		{"testing equal dates",
			"2017-06-07", "2017-06-07", "2017-06-07", "2017-06-07", nil},
		{"testing default start before end date",
			"2017-06-07", "2017-07-07", "2017-06-07", "2017-07-07", nil},
		{"testing start date after end date",
			"2017-07-07", "2017-06-07", "2017-06-07", "2017-07-07", nil},
	}

	for _, tc := range testCases {
		s, e, err := orderDates(tc.s, tc.e)
		if s != tc.expS && e != tc.expE {
			t.Errorf("\n%v: orderDates(%#v, %#v): \nexpected %#v %#v %#v, \n  actual %#v %#v %#v",
				tc.m, tc.s, tc.e, tc.expS, tc.expE, tc.expErr, s, e, err)
		}
	}
}

func TestParseDates(t *testing.T) {
	// unixNow represents the string for the now Unix time
	// 2016-12-31 - 1483142400 GMT
	// 2017-06-01 - 1496275200 GMT
	var unixNow = strconv.Itoa(int(time.Now().Unix()))

	var testCases = []struct {
		m      string // test message
		s      string // start input
		e      string // end input
		expS   string // expected start output
		expE   string // expected end output
		expErr error  // expected error output
	}{
		{"testing valid start and end date",
			"2016-12-31", "2017-06-01", "1483142400", "1496275200", nil},
		{"testing both start and end date with empty input",
			"", "", "0", unixNow, nil},
		{"testing both start and end date with invalid input",
			"1234", "abcd", "0", unixNow, nil},
		{"testing valid start date with empty end date",
			"2017-06-01", "", "1496275200", unixNow, nil},
		{"testing valid start date with invalid end date",
			"2017-06-01", "1234", "1496275200", unixNow, nil},
		{"testing empty start date with valid end date",
			"", "2017-06-01", "0", "1496275200", nil},
		{"testing invalid start date with valid end date",
			"abcd", "2017-06-01", "0", "1496275200", nil},
		{"testing both start and end date with same value",
			"2017-06-01", "2017-06-01", "1496275200", "1496275200", nil},
	}

	for _, tc := range testCases {
		s, e, err := parseDates(tc.s, tc.e)
		if (s != tc.expS) || (e != tc.expE) || (reflect.TypeOf(err) != reflect.TypeOf(tc.expErr)) {
			t.Errorf("\n%v: parseDates(%#v, %#v): \nexpected %#v %#v %#v, \n  actual %#v %#v %#v",
				tc.m, tc.s, tc.e, tc.expS, tc.expE, tc.expErr, s, e, err)
		}
	}
}

func TestParseDateStringToUnix(t *testing.T) {
	// unixNow represents the string for the now Unix time
	// 2006-01-02 - 1483142400 GMT
	// 2017-06-01 - 1496275200 GMT
	// 2016-02-29 - 1456704000 GMT

	var testCases = []struct {
		m       string // test message
		s       string // string input
		expUnix string // expected unix output
		expErr  error  // expected error output
	}{
		{"testing no input",
			"", "", errors.New("error")},
		{"testing standard date",
			"2006-01-02", "1136160000", nil},
		{"testing valid date",
			"2017-06-01", "1496275200", nil},
		{"testing leap year date",
			"2016-02-29", "1456704000", nil},
		{"testing invalid date",
			"2017-13-01", "2017-13-01", errors.New("error")},
		{"testing invalid date",
			"2017-07-32", "2017-07-32", errors.New("error")},
		{"testing invalid date with string",
			"abcd", "abcd", errors.New("error")},
		{"testing invalid date with integer",
			"1234", "1234", errors.New("error")},
	}

	for _, tc := range testCases {
		unix, err := parseDateStringToUnix(tc.s)
		if (unix != tc.expUnix) || (reflect.TypeOf(err) != reflect.TypeOf(tc.expErr)) {
			t.Errorf("\n%v: parseDateStringToUnix(%#v): \nexpected %#v %#v, \n  actual %#v %#v",
				tc.m, tc.s, tc.expUnix, tc.expErr, unix, err)
		}
	}
}
