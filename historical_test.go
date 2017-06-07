package yahoofinancequery

import (
	"testing"
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
