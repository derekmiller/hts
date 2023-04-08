package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

const (
	hollywoodTheatreURL = "https://hollywoodtheatre.org"
	dateTimeFormat      = "2006-01-023:04 PM"
	timezone            = "America/Los_Angeles"
)

type scrapedShowtime struct {
	Series  string
	Title   string
	Date    string
	Time    string
	URLPath string
}

type showtime struct {
	Series   string
	Title    string
	DateTime time.Time
	URL      string
}

func handler() {
	scrapedShowtimes := scrapeShowtimes()
	var wg sync.WaitGroup
	wg.Add(len(scrapedShowtimes))
	for _, s := range scrapedShowtimes {
		go func(s scrapedShowtime) {
			defer wg.Done()
			if s.Title == "" {
				fmt.Fprintf(os.Stderr, "no title found for scraped entry %v", s)
				return
			}
			if s.Date == "" || s.Time == "" {
				fmt.Fprintf(os.Stderr, "no date and/or time for scraped entry %v", s)
				return
			}
			parsedDateTime, err := parseDateTime(s.Date, s.Time)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to parse datetime for %v on %v at %v: %v", s.Title, s.Date, s.Time, err)
				return
			}
			url := hollywoodTheatreURL + s.URLPath
			st := showtime{
				Series:   strings.ReplaceAll(s.Series, ":", ""),
				Title:    s.Title,
				DateTime: parsedDateTime,
				URL:      url,
			}
			fmt.Printf("Showtime: %v\n", st)
			return
		}(s)
	}
	wg.Wait()
}

func main() {
	lambda.Start(handler)
}
