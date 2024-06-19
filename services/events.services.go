package services

import (
	"slices"
	"time"

	"github.com/mcg-dallgow/mcg-display/services/webuntis"
	. "github.com/mcg-dallgow/mcg-display/types"
)

func GetEvents(start, end time.Time) (events map[string][]Event, err error) {
	eventList := []Event{}

	exams, err := getExams(start, end)
	if err != nil {
		return events, err
	}
	calendarEvents, err := getCalendarEvents(start, end)
	if err != nil {
		return events, err
	}

	eventList = append(eventList, exams...)
	eventList = append(eventList, calendarEvents...)

	sortEvents(eventList)

	events = make(map[string][]Event)
	currentTime := start
	for !currentTime.After(end) {
		date := currentTime.Format("2006-01-02")

		events[date] = make([]Event, 0)
		for _, event := range eventList {
			if event.Date == date {
				events[date] = append(events[date], event)
			}
		}
		currentTime = currentTime.Add(24 * time.Hour)
	}

	return events, nil
}

func getExams(start, end time.Time) (events []Event, err error) {
	username, password, err := GetCredentials(1)
	if err != nil {
		return events, err
	}
	session, err := webuntis.Login(username, password)
	if err != nil {
		return events, err
	}
	defer session.Logout()
	exams, err := session.GetExams(start, end, false)
	if err != nil {
		return events, err
	}

	for _, exam := range exams {
		events = append(events, Event{
			Title:       exam.Name,
			Description: exam.Text,
			Category:    "Prüfung",
			Date:        exam.Start.Format("2006-01-02"),
			FullDay:     false,
			Start:       exam.Start.Time,
			End:         exam.End.Time,
			Location:    exam.Rooms[0].ShortName,
		})
	}

	return events, nil
}

func getCalendarEvents(start, end time.Time) (events []Event, err error) {
	username, password, err := GetCredentials(2)
	if err != nil {
		return events, err
	}
	session, err := webuntis.Login(username, password)
	if err != nil {
		return events, err
	}
	defer session.Logout()
	calendarEvents, err := session.GetCalendarEvents(start, end)
	if err != nil {
		return events, err
	}

	for _, calendarEvent := range calendarEvents {
		events = append(events, Event{
			Title:       calendarEvent.Name,
			Description: calendarEvent.Notes,
			Category:    calendarEvent.Calendar,
			Date:        calendarEvent.Date,
			FullDay:     calendarEvent.FullDay,
			Start:       calendarEvent.Start,
			End:         calendarEvent.End,
			Location:    calendarEvent.Location,
		})
	}

	return events, nil
}

func sortEvents(events []Event) {
	slices.SortFunc(events, func(a, b Event) int {
		if a.Start.Before(b.Start) {
			return -1
		} else if b.Start.Before(a.Start) {
			return 1
		}
		if a.FullDay && !b.FullDay {
			return -1
		} else if b.FullDay && !a.FullDay {
			return 1
		}
		if a.End.Before(b.End) {
			return -1
		} else if b.End.Before(a.End) {
			return 1
		}
		if a.Title < b.Title {
			return -1
		}
		return 1
	})

}
