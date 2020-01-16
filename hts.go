package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
)

const hollywoodTheatreURL = "https://hollywoodtheatre.org/m/calendar/"

type Showtime struct {
	Series string
	Title  string
	Date   string
	Time   string
}

func handler() error {
	c := colly.NewCollector()

	c.OnHTML(".calendar__events__day", func(e *colly.HTMLElement) {
		date := e.Attr("data-calendar-date")
		e.ForEach(".calendar__events__day__event", func(_ int, e *colly.HTMLElement) {
			series := e.ChildText(".calendar__events__day__event__series")
			title := e.ChildText(".calendar__events__day__event__title")
			e.ForEach(".showtime-square", func(_ int, e *colly.HTMLElement) {
				time := e.ChildText(":first-child")
				showtime := Showtime{
					Series: series,
					Title:  title,
					Date:   date,
					Time:   time,
				}
				fmt.Printf("Showtime: %v\n", showtime)
			})
		})
	})

	c.Visit(hollywoodTheatreURL)
	return nil
}

func main() {
	lambda.Start(handler)
}
