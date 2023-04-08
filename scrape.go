package main

import (
	"github.com/gocolly/colly"
)

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
