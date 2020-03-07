package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/gocolly/colly"
)

const (
	localDynamoDbURL    = "http://docker.for.mac.localhost:8000/"
	dynamoDbTableName   = "HollywoodTheatreShowtimes"
	hollywoodTheatreURL = "https://hollywoodtheatre.org/m/calendar/"
	dateTimeFormat      = "2006-01-023:04 PM"
	timezone            = "America/Los_Angeles"
)

type scrapedShowtime struct {
	Series string
	Title  string
	Date   string
	Time   string
}

type showtime struct {
	Series   string
	Title    string
	DateTime time.Time
}

func handler(ctx context.Context) error {
	svc := getDynamoDBService()
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
			st := showtime{
				Series:   strings.ReplaceAll(s.Series, ":", ""),
				Title:    s.Title,
				DateTime: parsedDateTime,
			}
			fmt.Printf("Showtime: %v\n", st)
			av, err := dynamodbattribute.MarshalMap(st)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error marshalling new showtime item %v: %v", av, err)
				return
			}
			input := &dynamodb.PutItemInput{
				Item:      av,
				TableName: aws.String(dynamoDbTableName),
			}
			_, err = svc.PutItem(input)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error calling PutItem %v: %v", input, err)
				return
			}
			return
		}(s)
	}
	wg.Wait()
	return nil
}

func getDynamoDBService() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
	if os.Getenv("ENVIRONMENT") == "development" {
		return dynamodb.New(
			sess,
			&aws.Config{
				Endpoint: aws.String(localDynamoDbURL),
			})
	}
	return dynamodb.New(sess)
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
