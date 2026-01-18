package main

import (
	"context"
	"errors"
	"log"

	"naggingbot/internal/app"
	"naggingbot/internal/scheduler"
	"naggingbot/internal/storage/memory"
	"naggingbot/internal/telegram"
)

func main() {
	log.Println("NaggingBot starting up (scheduler demo)")

	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	log.Printf("config loaded: poll=%s scheduler=%s db=%s", cfg.PollInterval, cfg.SchedulerInterval, cfg.DBPath)

	ctx := context.Background()
	occurrenceStore := memory.NewInMemoryOccurrenceStore()
	userStore := memory.NewInMemoryUserStore()
	reminderStore := memory.NewInMemoryReminderStore()

	notifier := &scheduler.LoggingNotifier{}
	sched := scheduler.New(occurrenceStore, reminderStore, notifier, cfg.SchedulerInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Telegram polling in a separate goroutine.
	tgClient := telegram.NewClient(cfg.BotToken, cfg.PollInterval, cfg.PollTimeout)
	tgHandler := telegram.NewStartHandler(userStore, reminderStore, occurrenceStore)
	go func() {
		if err := tgClient.Poll(ctx, tgHandler); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("telegram poller stopped: %v", err)
		}
	}()

	if err := sched.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("scheduler stopped with error: %v", err)
	}
}
