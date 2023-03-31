package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
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
	err := scrape()
	if err != nil {
		log.Panicf("error: %v", err)
	}
}

func scrape() error {
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
	return nil
}

func scrapeShowtimes() []scrapedShowtime {
	var scrapedShowtimes []scrapedShowtime
	c := colly.NewCollector()

	c.OnHTML(".calendar__events__day", func(e *colly.HTMLElement) {
		date := e.Attr("data-calendar-date")
		e.ForEach(".calendar__events__day__event", func(_ int, e *colly.HTMLElement) {
			series := e.ChildText(".calendar__events__day__event__series")
			title := e.ChildText(".calendar__events__day__event__title")
			e.ForEach(".showtime-square", func(_ int, e *colly.HTMLElement) {
				time := e.ChildText(":first-child")
				urlPath := e.ChildAttr(":first-child", "href")
				scrapedShowtime := scrapedShowtime{
					Series:  series,
					Title:   title,
					Date:    date,
					Time:    time,
					URLPath: urlPath,
				}
				scrapedShowtimes = append(scrapedShowtimes, scrapedShowtime)
			})
		})
	})

	c.Visit(hollywoodTheatreURL)
	return scrapedShowtimes
}

func parseDateTime(date, t string) (time.Time, error) {
	dateTime := date + t
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}
	parsedDateTime, err := time.ParseInLocation(dateTimeFormat, dateTime, location)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDateTime, nil
}

func main() {
	lambda.Start(handler)
}
