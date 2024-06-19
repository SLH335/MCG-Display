package services

import (
	"regexp"
	"slices"
	"strings"
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
			Title:       generateExamTitle(exam),
			Description: generateExamDescription(exam),
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
	case "KA":
		examType = "Klassenarbeit"
	default:
		examType = exam.Type.DisplayName
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
	examSubject := ""
	for i := 0; i < int(WAT); i++ {
		subject := Subject(i)
		if exam.Subject.ShortName != "" {
			if subject.Short() == exam.Subject.ShortName[:2] {
				examSubject = subject.String()
				break
			}
		} else {
			if AnyContain([]string{exam.Name, exam.Text}, subject.String(), true) ||
				AnyContain([]string{exam.Name, exam.Text}, subject.Short(), false) {

				examSubject = subject.String()
				break
			}
		}
	}

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
	description := exam.Text
	// use original exam name as description, if real description is empty
	if description == "" {
		description = exam.Name
	}

	// do not display description if every word is already present in generated exam title
	title := generateExamTitle(exam)
	usefulDescription := description
	for _, word := range strings.Split(title, " ") {
		usefulDescription = strings.ReplaceAll(usefulDescription, word, "")
	}
	if strings.TrimSpace(usefulDescription) == "" {
		return ""
	}
	return description
}
