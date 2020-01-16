package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly"
)

type Showtime struct {
	Title string
	Date  string
	Time  string
}

func handler(request events.APIGatewayProxyRequest) error {
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit("https://hollywoodtheatre.org/m/calendar/")
	return nil
}

func main() {
	lambda.Start(handler)
}
