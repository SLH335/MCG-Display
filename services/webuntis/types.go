package webuntis

import (
	"strings"
	"time"
)

type Time struct {
	time.Time
}

func (timestamp *Time) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")[:19]
	if s == "null" {
		timestamp.Time = time.Time{}
		return
	}
	timestamp.Time, err = time.Parse("2006-01-02T15:04:05", s)
	return
}

type Exam struct {
	Id           int               `json:"examId"`
	Type         UntisValue        `json:"examType"`
	Name         string            `json:"examName"`
	Text         string            `json:"examText"`
	Start        Time              `json:"examStart"`
	End          Time              `json:"examEnd"`
	Duration     int               `json:"examDuration"`
	NumStudents  int               `json:"numStudents"`
	Subject      UntisValue        `json:"subject"`
	Classes      []UntisValue      `json:"classes"`
	Students     []AssignedStudent `json:"students"`
	Teachers     []UntisValue      `json:"teachers"`
	Invigilators []Invigilator     `json:"invigilators"`
	Rooms        []UntisValue      `json:"rooms"`
}

type CalendarEvent struct {
	Id       int64
	Name     string
	Notes    string
	Date     string
	Start    time.Time
	End      time.Time
	FullDay  bool
	Location string
	Calendar string
	Color    string
}

type UntisValue struct {
	Id          int    `json:"id"`
	ShortName   string `json:"shortName"`
	LongName    string `json:"longName"`
	DisplayName string `json:"displayName"`
}

type AssignedStudent struct {
	Id                       int    `json:"id"`
	ShortName                string `json:"shortName"`
	LongName                 string `json:"longName"`
	DisplayName              string `json:"displayName"`
	Gender                   string `json:"gender"`
	ImageUrl                 string `json:"imageUrl"`
	GradeProtection          bool   `json:"gradeProtection"`
	DisadvantageCompensation bool   `json:"disadvantageCompensation"`
}

type Invigilator struct {
	Start    Time         `json:"start"`
	End      Time         `json:"end"`
	Teachers []UntisValue `json:"teachers"`
}
