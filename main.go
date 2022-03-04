package main

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	"github.com/artnoi43/fngobot/fetch"
	"github.com/artnoi43/fngobot/fetch/bitkub"
	"github.com/artnoi43/fngobot/fetch/satang"
	"github.com/pkg/errors"
)

type exchange string
type answer map[string]map[exchange]fetch.Quoter

var (
	bk       exchange = "Bitkub"
	st       exchange = "Satang"
	fetchMap          = map[exchange]fetch.FetchFunc{
		bk: bitkub.Get,
		st: satang.Get,
	}
	printer = message.NewPrinter(language.English)
)

func init() {
	if len(os.Args) < 1 {
		panic("missing ticker(s)")
	}
}

func main() {
	tickers := os.Args[1:]
	var wg sync.WaitGroup
	for _, ticker := range tickers {
		ticker = strings.ToUpper(ticker)
		wg.Add(1)
		go func(t string) {
			defer wg.Done()
			quotes := getQuotes(t)
			printQuotes(quotes)
		}(ticker)
	}
	wg.Wait()
}

func printQuotes(a answer) {
	var errs []string
	appendErrs := func(err error, e exchange, t, what string) {
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("error getting %s for %s (%s)", what, t, e))
			errs = append(errs, err.Error())
		}
	}
	for t, m := range a {
		for e, q := range m {
			last, err := q.Last()
			appendErrs(err, e, t, "last")
			bid, err := q.Bid()
			appendErrs(err, e, t, "bid")
			ask, err := q.Ask()
			appendErrs(err, e, t, "ask")
			printer.Printf("%s: %s - bid %v, ask %v, last %v\n", e, t, bid, ask, last)
		}
	}
	printer.Println("errors encountered:", errs)
}

func getQuotes(t string) answer {
	return answer{
		t: {
			bk: fetchQuotes(t, bk),
			st: fetchQuotes(t, st),
		},
	}
}

func fetchQuotes(t string, e exchange) fetch.Quoter {
	q, err := fetchMap[e](t)
	if err != nil {
		panic(fmt.Sprintln("failed to get quote for", e))
	}
	return q
}
