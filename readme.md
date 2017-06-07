# Yahho Finance Query in golang

## Historical quotes

in Q1/2017 Yahoo changed the endpoint for quering historical quotes

new query url:

https://query1.finance.yahoo.com/v7/finance/download/BAS.DE?period1=946854000&period2=1496008800&interval=1d&events=history&crumb=fCv.rYUxTML

translates to:

https://query1.finance.yahoo.com/v7/finance/download/{symbol}?period1={startDate}&period2={endDate}&interval={interval}&events={type}&crumb={crumb}

symbol = the symbol of the stock
startDate = start date as unix time stamp
endDate = end date as unix time stamp
interval = what time interval (1d = daily, 1wk = weekly, 1mo = monthly)
event = type of the query (history, div, split)
crumb = a cookie parameter from the main stock page to verify, that you come from the actual yahoo page

to retrieve the crumb, first a call to the main stock page is necassary:
https://finance.yahoo.com/quote/BAS.DE/history?p=BAS.DE

translates to:
https://finance.yahoo.com/quote/{symbol}/history?p={symbol}

the yahoo finance page is a reactjs page, which is rendered from a json string. We have to search for a `<script>` tag with `root.App.main`:
 
```
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