package scheduler

import (
	"context"
	"log"
)

// Notifier sends reminder messages to the user.
type Notifier interface {
	Send(ctx context.Context, occ OccurrenceWithReminder) error
}

// LoggingNotifier logs outgoing notifications.
type LoggingNotifier struct{}

func (n *LoggingNotifier) Send(ctx context.Context, occ OccurrenceWithReminder) error {
	if occ.Reminder != nil {
		log.Printf("notifier: occurrence %d sent | reminder=%q desc=%q",
			occ.Occurrence.ID, occ.Reminder.Name, occ.Reminder.Description)
	} else {
		log.Printf("notifier: occurrence %d sent", occ.Occurrence.ID)
	}
	return nil
}

// MultiNotifier dispatches to multiple notifiers.
type MultiNotifier struct {
	inner []Notifier
}

func NewMultiNotifier(notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{inner: notifiers}
}

func (m *MultiNotifier) Send(ctx context.Context, occ OccurrenceWithReminder) error {
	for _, n := range m.inner {
		if err := n.Send(ctx, occ); err != nil {
			// Log and continue fan-out.
			log.Printf("notifier error: %v", err)
		}
	}
	return nil
}
