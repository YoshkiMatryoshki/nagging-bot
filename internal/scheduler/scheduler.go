package scheduler

import (
	"context"
	"log"
	"time"

	"naggingbot/internal/domain"
)

// Notifier sends reminder messages to the user.
type Notifier interface {
	Send(ctx context.Context, occ OccurrenceWithReminder) error
}

// Scheduler polls due occurrences and sends reminders.
type Scheduler struct {
	occurrenceStore domain.OccurrenceStore
	reminderStore   domain.ReminderStore
	notifier        Notifier
	interval        time.Duration
}

// New constructs a scheduler with a polling interval.
func New(occurrences domain.OccurrenceStore, reminders domain.ReminderStore, notifier Notifier, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = time.Minute
	}

	return &Scheduler{
		occurrenceStore: occurrences,
		reminderStore:   reminders,
		notifier:        notifier,
		interval:        interval,
	}
}

// Run starts the scheduler loop.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := s.tick(ctx); err != nil {
				log.Printf("scheduler tick error: %v", err)
			}
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) error {
	log.Println("scheduler tick")

	nowUTC := time.Now().UTC()
	due, err := s.occurrenceStore.ListPendingInRange(ctx, time.Time{}, nowUTC)
	if err != nil {
		return err
	}

	for _, occ := range due {
		payload := OccurrenceWithReminder{Occurrence: occ}
		if occ.ReminderID != 0 {
			if rem, err := s.reminderStore.GetByID(ctx, occ.ReminderID); err == nil {
				payload.Reminder = rem
			}
		}

		if err := s.notifier.Send(ctx, payload); err != nil {
			log.Printf("send occurrence %d failed: %v", occ.ID, err)
			continue
		}

		if err := s.occurrenceStore.UpdateStatus(ctx, occ.ID, domain.OccurrenceSent); err != nil {
			log.Printf("update occurrence %d status failed: %v", occ.ID, err)
		}
	}

	return nil
}

// OccurrenceWithReminder bundles occurrence and optional reminder for notifier.
type OccurrenceWithReminder struct {
	Occurrence *domain.Occurrence
	Reminder   *domain.Reminder
}
