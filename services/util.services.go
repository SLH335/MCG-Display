package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
)

func RenderComponent(component templ.Component) string {
	buf := new(bytes.Buffer)
	component.Render(context.Background(), buf)

	return buf.String()
}

func GetCredentials(n int) (username, password string, err error) {
	err = godotenv.Load()
	if err != nil {
		return "", "", err
	}

	suffix := fmt.Sprintf("_%d", n)

	username = os.Getenv("WEBUNTIS_USERNAME" + suffix)
	password = os.Getenv("WEBUNTIS_PASSWORD" + suffix)

	if username == "" && password == "" {
		return "", "", errors.New("error: no credentials found in .env")
	}

	return username, password, nil
}

func ParseDateRange(start, end, days string) (startTime, endTime time.Time, err error) {
	const layout string = "2006-01-02"
	const defaultDays int = 7

	if start == "" {
		// if no start date is given, use today
		now := time.Now()
		startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		// if start date is given, try to parse time
		startTime, err = time.Parse(layout, start)
		if err != nil {
			return startTime, endTime, errors.New("error: start date is invalid")
		}
	}

	if days == "" && end == "" {
		// if neither end date nor day amount is given, use range of one week
		endTime = startTime.Add(time.Duration(defaultDays-1) * 24 * time.Hour)
	} else if days == "" {
		// if end date is given, try to parse time
		endTime, err = time.Parse(layout, end)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("error: end date is invalid")
		}
	} else if end == "" {
		daysInt, err := strconv.Atoi(days)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("error: day amount is not a valid integer")
		}
		endTime = startTime.Add(time.Duration(daysInt-1) * 24 * time.Hour)
	} else {
		// error if both end date and day amount are given
		return time.Time{}, time.Time{}, errors.New("error: end date and day amount cannot both be given")
	}

	// error if end date is not after start date
	if !endTime.After(startTime) {
		return time.Time{}, time.Time{}, errors.New("error: end date must be after start date")
	}

	return startTime, endTime, nil
}
