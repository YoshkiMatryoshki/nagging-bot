package domain

import "time"

// Occurrence represents a single scheduled reminder execution.
type Occurrence struct {
	ID         int64
	ReminderID int64
	FireAtUtc  time.Time
	Status     OccurrenceStatus
}

// OccurrenceStatus is the lifecycle state of an occurrence.
type OccurrenceStatus int

const (
	OccurrenceCreated OccurrenceStatus = iota
	OccurrenceSent
	OccurrenceDone
	OccurrenceIgnored
)
