# yfc - Yahoo Finance Query in golang

A web scrapper to query the Yahoo finance page for stock quotes.

## Basic example

```go
ackage main

import (
    "fmt"

    "github.com/dirkolbrich/yfc"
)

func main() {
    // create a new Historical
    historical := yfc.NewHistorical()

    // start a query
    quotes, _ := historical.Query("AAPL")

    // print the result
    fmt.Printf("received quotes for AAPL: %v\n", len(quotes))
    if len(quotes) > 0 {
        fmt.Printf("first: %+v\n", quotes[0])
        fmt.Printf("last:  %+v\n", quotes[len(quotes)-1])
    }
}
```

`quotes` is a slice of `[]Quotes` which contains all available historical quotes in ascending order.

```go
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
```

Start and End Date of the query can be set individually, e.g. for retrieving just the quotes from the last year. Start and end date must be a date formated according to ISO 8601: yyyy-mm-dd or an empty string.

- If both start and end date are valid, the date range will be parsed as specified.
- If both start and end date are empty or invalid, the date range is set to start date at unix = 0 and end date to unix = now().
- If start date is empty/invalid and end date is valid, date range is set to start date at unix = 0 and end date to the specified date.
- If start date is valid and end date is empty/invalid, date range is set to the specified start date and end date to unix = now().

```go
// set a range for the query
historical.StartDate = "2017-01-01"
historical.EndDate = "2017-12-31"
```

## How does it work

In Q1/2017 Yahoo changed the endpoint for quering historical quotes.

The new query URL is:

`https://query1.finance.yahoo.com/v7/finance/download/BAS.DE?period1=946854000&period2=1496008800&interval=1d&events=history&crumb=fCv.rYUxTML`

which translates to:

`https://query1.finance.yahoo.com/v7/finance/download/{symbol}?period1={startDate}&period2={endDate}&interval={interval}&events={type}&crumb={crumb}`

- `symbol` = the symbol of the stock
- `startDate` = start date as unix time stamp
- `endDate` = end date as unix time stamp
- `interval` = what time interval (1d = daily, 1wk = weekly, 1mo = monthly) for the quotes
- `event` = type of the query (history, div, split)
- `crumb` = a cookie parameter from the main stock page to verify, that you come from the actual yahoo page

To retrieve the crumb *and* the session cookie, first a call to the main stock page is necessary:

`https://finance.yahoo.com/quote/BAS.DE/history?p=BAS.DE`

which translates to:

`https://finance.yahoo.com/quote/{symbol}/history?p={symbol}`

The Yahoo finance page is a Reactjs page, which is rendered from a JSON string. We have to search for a `<script>` tag within `root.App.main`:

```json
<script>
// ...
root.App.main = {
    "context": {
        "dispatcher": {
            "stores": {
                // ...
                "CrumbStore": {
                    "crumb": "fCv.rYUxTML" // this is the actual crumb
                },
                // ...
            }
        }
    }
};
// ...
</script>
```

For the actual query to download the historical quotes in cvs format, the `crumb` *and* the cookie have to be included in the main query.
