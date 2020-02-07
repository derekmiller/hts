package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
	"google.golang.org/appengine/log"
)

const hollywoodTheatreURL = "https://hollywoodtheatre.org/m/calendar/"

type scrapedShowtime struct {
	Series string
	Title  string
	Date   string
	Time   string
}

type showtime struct {
	Series        string
	Title         string
	StartDateTime time.Time
}

func handler(ctx context.Context) error {
	scrapedShowtimes := scrapeShowtimes(hollywoodTheatreURL)

	for _, s := range scrapedShowtimes {
		if s.Title == "" {
			log.Errorf(ctx, "no title found for scraped entry %v", s)
			continue
		}
		if s.Date == "" || s.Time == "" {
			log.Errorf(ctx, "no date and/or time for %v", s)
			continue
		}
		parsedDateTime, err := parseDateTime(s.Date, s.Time)
		if err != nil {
			log.Errorf(ctx, "unable to parse datetime for %v on %v at %v: %v", s.Title, s.Date, s.Time, err)
			continue
		}
		showtime := showtime{
			Series:        s.Series,
			Title:         s.Title,
			StartDateTime: parsedDateTime,
		}
		fmt.Printf("Showtime: %v\n", showtime)
	}
	return nil
}

func scrapeShowtimes(url string) []scrapedShowtime {
	var scrapedShowtimes []scrapedShowtime
	c := colly.NewCollector()

	c.OnHTML(".calendar__events__day", func(e *colly.HTMLElement) {
		date := e.Attr("data-calendar-date")
		e.ForEach(".calendar__events__day__event", func(_ int, e *colly.HTMLElement) {
			series := e.ChildText(".calendar__events__day__event__series")
			title := e.ChildText(".calendar__events__day__event__title")
			e.ForEach(".showtime-square", func(_ int, e *colly.HTMLElement) {
				time := e.ChildText(":first-child")
				scrapedShowtime := scrapedShowtime{
					Series: series,
					Title:  title,
					Date:   date,
					Time:   time,
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
	format := "2006-01-023:04 PM"
	location, err := time.LoadLocation("America/Los_Angeles")
	parsedDateTime, err := time.ParseInLocation(format, dateTime, location)
	return parsedDateTime, err
}

func main() {
	lambda.Start(handler)
}
