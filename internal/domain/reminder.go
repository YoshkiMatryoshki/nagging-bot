package domain

import "time"

// Reminder defines a repeating rule created by the user.
type Reminder struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	TimesOfDay  []TimeOfDay
	// TimeZone stores the IANA time zone (e.g., "Europe/Moscow") used to compute occurrences.
	// TODO: support re-computing future occurrences if the user changes their preferred time zone.
	TimeZone string
	IsActive    bool
}

// TimeOfDay stores a wall-clock time without a date.
type TimeOfDay struct {
	Hour   int
	Minute int
}
