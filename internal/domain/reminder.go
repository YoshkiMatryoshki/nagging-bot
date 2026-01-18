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
	TimeZone    string
	IsActive    bool
}

// TimeOfDay stores a wall-clock time without a date.
type TimeOfDay struct {
	Hour   int
	Minute int
}
