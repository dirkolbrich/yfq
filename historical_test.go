package yahoofinancequery

import (
	"errors"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// test html generated from finance.yahoo.com
var testHTML = `
    <!DOCTYPE html>
    <html>
    <head">
        <meta charset="utf-8">
        <title>title</title>
        <meta name="keywords" content="test">
        <meta name="description" lang="en-US" content="yahoo finance test html">
    </head>

    <body>
        <div id="app">
            <p>test paragraph</p>
        </div>
        <script>
            (function (root) {
                /* -- Data -- */
                root.App || (root.App = {});
                root.App.main = {
                  "context": {
                    "dispatcher": {
                      "stores": {
                        "FlyoutStore": {},
                        "CrumbStore": {
                          "crumb": "fCv.rYUxTML"
                        },
                        "HeaderStore": {}
                      }
                    },
                    "options": {}
                  },
                  "plugins": {}
                };
            }(this));
        </script>
    </body>
    </html>
`

// nowUnix represents the string for the now Unix time
var unixNow = strconv.Itoa(int(time.Now().Unix()))

// 2016-12-31 - 1483142400 GMT
// 2017-06-01 - 1496275200 GMT

// dateParseTests is a table for testing parsing Dates
var parseDateTests = []struct {
	s      string // start input
	e      string // end input
	expS   string // expected start output
	expE   string // expected end output
	expErr error  // expected error output
}{
	{"", "", "0", unixNow, nil},                                   // no input
	{"2017-06-01", "", "1496275200", unixNow, nil},                // only start date
	{"", "2017-06-01", "0", "1496275200", nil},                    // only end date
	{"2016-12-31", "2017-06-01", "1483142400", "1496275200", nil}, // both same date
	{"16-12-31", "", "16-12-31", "", errors.New("error")},         // start invalid date
	{"", "17-06-01", "", "17-06-01", errors.New("error")},         // start invalid date
}

func TestParseDates(t *testing.T) {
	for n, tt := range parseDateTests {
		s, e, err := parseDates(tt.s, tt.e)
		if (s != tt.expS) || (e != tt.expE) || (reflect.TypeOf(err) != reflect.TypeOf(tt.expErr)) {
			t.Errorf("parseDates(%d, %d): expected %d %d %v, actual %d %d %v", tt.s, tt.e, tt.expS, tt.expE, tt.expErr, s, e, err)
		}
	}
}

// dateTests is a table for testing ordering Dates
var dateTests = []struct {
	s    string // start input
	e    string // end input
	expS string // expected start output
	expE string // expected end output
}{
	{"", "", "", ""},                                         // no input
	{"2017-06-07", "", "2017-06-07", ""},                     // only start date
	{"", "2017-06-07", "", "2017-06-07"},                     // only end date
	{"2017-06-07", "2017-06-07", "2017-06-07", "2017-06-07"}, // both same date
	{"2017-06-07", "2017-07-07", "2017-06-07", "2017-07-07"}, // default s earlier then e
	{"2017-07-07", "2017-06-07", "2017-06-07", "2017-07-07"}, // s later than e
}

func TestOrderDates(t *testing.T) {
	for _, tt := range dateTests {
		s, e := orderDates(tt.s, tt.e)
		if s != tt.expS && e != tt.expE {
			t.Errorf("orderDates(%d, %d): expected %d %d, actual %d %d", tt.s, tt.e, tt.expS, tt.expE, s, e)
		}
	}
}
