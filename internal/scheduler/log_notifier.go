package scheduler

import (
	"context"
	"log"

	"naggingbot/internal/domain"
)

// LoggingNotifier is a stub notifier that logs instead of calling Telegram.
type LoggingNotifier struct{}

func (n *LoggingNotifier) Send(ctx context.Context, occ *domain.Occurrence) error {
	log.Printf("notifier: occurrence %d sent", occ.ID)
	return nil
}
