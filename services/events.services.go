package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/mcg-dallgow/mcg-display/services/webuntis"
	. "github.com/mcg-dallgow/mcg-display/types"
)

func GetEvents(start, end time.Time) (events map[string][]Event, err error) {
	eventList := []Event{}

	username, password, err := GetCredentials()
	if err != nil {
		return events, err
	}
	session, err := webuntis.LoginPassword(username, password)
	if err != nil {
		return events, err
	}
	defer session.Logout()

	exams, err := getExams(session, start, end)
	if err != nil {
		return events, err
	}
	calendarEvents, err := getCalendarEvents(session, start, end)
	if err != nil {
		return events, err
	}
	timetableEvents, err := getTimetableEvents(session, start, end)
	if err != nil {
		return events, err
	}

	eventList = append(eventList, exams...)
	eventList = append(eventList, calendarEvents...)
	eventList = append(eventList, timetableEvents...)

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

func getExams(session webuntis.Session, start, end time.Time) (events []Event, err error) {
	cache := Cache{"exams", start, end}
	events, err = getCachedEvents(cache)
	if err == nil && len(events) > 0 {
		return events, nil
	}

	exams, err := session.GetExams(start, end, false)
	if err != nil {
		return events, err
	}

	for _, exam := range exams {
		events = append(events, Event{
			Title:       generateExamTitle(exam),
			Description: generateExamDescription(exam),
			Category:    ExamEvent,
			Date:        exam.Start.Format("2006-01-02"),
			FullDay:     false,
			Start:       exam.Start.Time,
			End:         exam.End.Time,
			Location:    formatLocation(exam.Rooms[0].ShortName),
		})
	}

	eventsJson, err := json.Marshal(events)
	cache.Write(eventsJson)

	return events, nil
}

func getCalendarEvents(session webuntis.Session, start, end time.Time) (events []Event, err error) {
	cache := Cache{"calendar", start, end}
	events, err = getCachedEvents(cache)
	if err == nil && len(events) > 0 {
		return events, nil
	}

	calendarEvents, err := session.GetCalendarEvents(start, end)
	if err != nil {
		return events, err
	}

	for _, calendarEvent := range calendarEvents {
		events = append(events, Event{
			Title:       calendarEvent.Name,
			Description: calendarEvent.Notes,
			Category:    getCalendarEventCategory(calendarEvent),
			Date:        calendarEvent.Date,
			FullDay:     calendarEvent.FullDay,
			Start:       calendarEvent.Start,
			End:         calendarEvent.End,
			Location:    formatLocation(calendarEvent.Location),
		})
	}

	eventsJson, err := json.Marshal(events)
	cache.Write(eventsJson)

	return events, nil
}

func getTimetableEvents(session webuntis.Session, start, end time.Time) (events []Event, err error) {
	cache := Cache{"timetable", start, end}
	events, err = getCachedEvents(cache)
	if err == nil && len(events) > 0 {
		return events, nil
	}

	timetableEvents, err := session.GetTimetableEvents(start, end)
	if err != nil {
		return events, err
	}

	for _, timetableEvent := range timetableEvents {
		title := fmt.Sprintf("%s %s", timetableEvent.Title, strings.Join(timetableEvent.Classes, ", "))

		events = append(events, Event{
			Title:    title,
			Category: StudentEvent,
			Date:     timetableEvent.Start.Format("2006-01-02"),
			FullDay:  false,
			Start:    timetableEvent.Start,
			End:      timetableEvent.End,
		})
	}

	eventsJson, err := json.Marshal(events)
	cache.Write(eventsJson)

	return events, err
}

func getCachedEvents(cache Cache) (events []Event, err error) {
	if !cache.IsValid() {
		return events, errors.New("error: event cache is not valid")
	}
	eventsJson, err := cache.Load()
	if err != nil {
		return events, err
	}
	err = json.Unmarshal(eventsJson, &events)

	return events, err
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

func generateExamTitle(exam webuntis.Exam) string {
	// exam type
	examType := ""
	switch exam.Type.ShortName {
	case "LEK-Test":
		if AnyContain([]string{exam.Name, exam.Text}, "Test", true) {
			examType = "Test"
		} else {
			examType = "LEK"
		}
	default:
		examType = exam.Type.ShortName
	}

	// exam class or grade level
	examClass := ""
	examClasses := []string{}
	for _, class := range exam.Classes {
		examClasses = append(examClasses, strings.Replace(class.ShortName, "Jhg", "Jg", 1))
	}
	slices.Sort(examClasses)
	examClasses = slices.Compact(examClasses)
	if len(examClasses) > 2 {
		gradeLevels := []string{}
		re := regexp.MustCompile("[0-9]+")
		for _, class := range examClasses {
			gradeLevels = append(gradeLevels, re.FindString(class))
		}
		if len(slices.Compact(gradeLevels)) == 1 {
			examClass = "Jg" + gradeLevels[0]
		}
	}
	if examClass == "" {
		examClass = strings.Join(examClasses, ", ")
	}

	// exam subject
	examSubject := getExamSubject(exam).String()

	// exam course type
	examCourseType := ""
	if examClass == "Jg11" || examClass == "Jg12" {
		for _, courseType := range []string{"GK", "LK"} {
			subject := ""
			if len(exam.Subject.ShortName) >= 2 {
				subject = exam.Subject.ShortName[:2]
			}
			if AnyContain([]string{exam.Name, exam.Text, subject}, courseType, false) {
				examCourseType = courseType
				break
			}
		}
	}

	// exam teacher
	examTeachers := []string{}
	for _, teacher := range exam.Teachers {
		teacherName := teacher.LongName

		switch teacher.ShortName {
		case "UrSoF":
			teacherName = "Urschel"
		}

		examTeachers = append(examTeachers, teacherName)
	}
	examTeacher := strings.Join(examTeachers, ", ")

	// combine elements into title
	title := strings.Join([]string{examType, examClass, examSubject, examCourseType, examTeacher}, " ")
	title = strings.ReplaceAll(strings.TrimSpace(title), "  ", " ")
	return title
}

func generateExamDescription(exam webuntis.Exam) string {
	var usedWords []string
	title := generateExamTitle(exam)

	if Contains(title, "GK", false) {
		usedWords = append(usedWords, []string{"Grund", "Grundkurs"}...)
	} else if Contains(title, "LK", false) {
		usedWords = append(usedWords, []string{"Leistungs", "Leistungskurs"}...)
	}
	if Contains(title, "KA", false) {
		usedWords = append(usedWords, "Klassenarbeit")
	}
	usedWords = append(usedWords, getExamSubject(exam).Variants()...)
	usedWords = append(usedWords, strings.Split(title, " ")...)

	if isUseful(exam.Name, usedWords) && isUseful(exam.Text, usedWords) && len(exam.Name)+len(exam.Text) < 75 {
		return exam.Name + " - " + exam.Text
	}

	if isUseful(exam.Text, usedWords) {
		return exam.Text
	}

	if isUseful(exam.Name, usedWords) {
		return exam.Name
	}

	return ""
}

func isUseful(text string, usedWords []string) bool {
	if len(text) == 0 {
		return false
	}
	usefulText := text
	for _, word := range usedWords {
		usefulText = strings.ReplaceAll(usefulText, word, "")
	}
	if float32(len(strings.Trim(usefulText, " .,-/&0123456789")))/float32(len(text)) < 0.4 {
		return false
	}
	return true
}

func getExamSubject(exam webuntis.Exam) Subject {
	for i := 0; i < int(EmptySubject)-1; i++ {
		subject := Subject(i)
		if exam.Subject.ShortName != "" {
			shortName := exam.Subject.ShortName[:2]
			if subject == Seminarkurs && len(exam.Subject.ShortName) >= 4 {
				shortName = exam.Subject.ShortName[2:4]
			}
			if subject.Short() == shortName {
				return subject
			}
		} else {
			if AnyContainAny([]string{exam.Name, exam.Text}, subject.Variants(), false) {
				return subject
			}
		}
	}
	return EmptySubject
}

func getCalendarEventCategory(event webuntis.CalendarEvent) EventCategory {
	switch event.Calendar {
	case "Termine Jahrgang 7-9":
		return SekIEvent
	case "Termine Jahrgang 10 und Oberstufe":
		return SekIIEvent
	case "Lernende":
		return StudentEvent
	case "Lehrkräfte":
		return TeacherEvent
	case "Öffentlich":
		if Contains(event.Name, "AG", false) {
			return AGEvent
		} else {
			return PublicEvent
		}
	default:
		return PublicEvent
	}
}

func formatLocation(room string) string {
	switch room {
	case "Turnhalle":
		return "TH"
	case "SHA":
		return "TH (A)"
	case "SHB":
		return "TH (B)"
	case "SHC":
		return "TH (C)"
	}
	return room
}
