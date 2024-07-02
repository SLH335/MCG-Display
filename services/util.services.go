package services

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/joho/godotenv"
)

func RenderComponent(component templ.Component) string {
	buf := new(bytes.Buffer)
	component.Render(context.Background(), buf)

	return buf.String()
}

func GetCredentials() (username, password string, err error) {
	err = godotenv.Load()
	if err != nil {
		return "", "", err
	}

	username = os.Getenv("WEBUNTIS_USERNAME")
	password = os.Getenv("WEBUNTIS_PASSWORD")

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

	// error if end date is before start date
	if endTime.Before(startTime) {
		return time.Time{}, time.Time{}, errors.New("error: end date cannot be before start date")
	}

	return startTime, endTime, nil
}

func Contains(str, sub string, ignoreCase bool) bool {
	if ignoreCase {
		return strings.Contains(strings.ToLower(str), strings.ToLower(sub))
	} else {
		return strings.Contains(str, sub)
	}
}

func ContainsAny(str string, subs []string, ignoreCase bool) bool {
	for _, sub := range subs {
		if Contains(str, sub, ignoreCase) {
			return true
		}
	}
	return false
}

func AnyContain(strs []string, sub string, ignoreCase bool) bool {
	for _, str := range strs {
		if Contains(str, sub, ignoreCase) {
			return true
		}
	}
	return false
}

func AnyContainAny(strs []string, subs []string, ignoreCase bool) bool {
	for _, str := range strs {
		for _, sub := range subs {
			if Contains(str, sub, ignoreCase) {
				return true
			}
		}
	}
	return false
}
