package main

import "time"

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
