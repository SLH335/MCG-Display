package types

import "time"

type Event struct {
	Title       string
	Description string
	Category    EventCategory
	Date        string
	FullDay     bool
	Start       time.Time
	End         time.Time
	Location    string
}

type EventCategory int

const (
	PublicEvent EventCategory = iota
	AGEvent
	ExamEvent
	StudentEvent
	SekIEvent
	SekIIEvent
	TeacherEvent
)

func (c EventCategory) String() string {
	return []string{
		"Öffentlich",
		"AG",
		"Prüfung",
		"Lernende",
		"Sek I",
		"Sek II",
		"Lehrkräfte",
	}[c]
}

func (c EventCategory) Color() string {
	return []string{
		"emerald-400", // Öffentlich
		"emerald-400", // AG
		"rose-400",    // Prüfung
		"amber-400",   // Lernende
		"amber-400",   // Sek I
		"amber-400",   // Sek II
		"sky-400",     // Lehrkräfte
	}[c]
}

func (c EventCategory) BackgroundColor() string {
	return []string{
		"[#D6E4E1]", // Öffentlich
		"[#D6E4E1]", // AG
		"[#E7D8DD]", // Prüfung
		"[#E8E2DB]", // Lernende
		"[#E8E2DB]", // Sek I
		"[#E8E2DB]", // Sek II
		"[#D6E1ED]", // Lehrkräfte
	}[c]
}
