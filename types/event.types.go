package types

import "time"

type Event struct {
	Title       string
	Description string
	Category    string
	Date        string
	FullDay     bool
	Start       time.Time
	End         time.Time
	Location    string
}
