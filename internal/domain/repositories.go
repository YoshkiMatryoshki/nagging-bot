package domain

import (
	"context"
	"time"
)

// UserStore defines the minimal operations needed for user data.
type UserStore interface {
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	Upsert(ctx context.Context, user *User) error
}

// ReminderStore defines the minimal operations needed for reminders.
type ReminderStore interface {
	GetByID(ctx context.Context, id int64) (*Reminder, error)
	ListByUser(ctx context.Context, userID int64) ([]*Reminder, error)
	Create(ctx context.Context, reminder *Reminder) error
	Update(ctx context.Context, reminder *Reminder) error
	DeleteByID(ctx context.Context, id int64) error
}

// OccurrenceStore defines the minimal operations needed for occurrences.
type OccurrenceStore interface {
	GetByID(ctx context.Context, id int64) (*Occurrence, error)
	ListByReminder(ctx context.Context, reminderID int64) ([]*Occurrence, error)
	ListPendingInRange(ctx context.Context, startUTC, endUTC time.Time) ([]*Occurrence, error)
	Create(ctx context.Context, occurrence *Occurrence) error
	UpdateStatus(ctx context.Context, id int64, status OccurrenceStatus) error
	DeleteByReminder(ctx context.Context, reminderID int64) error
}
