package scheduler

import (
	"context"
	"log"
)

// LoggingNotifier is a stub notifier that logs instead of calling Telegram.
// It expects reminder info to be passed in the payload.
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
