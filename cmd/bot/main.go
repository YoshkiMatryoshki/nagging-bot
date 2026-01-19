package main

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"naggingbot/internal/app"
	"naggingbot/internal/scheduler"
	"naggingbot/internal/storage/sqlite"
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
	if err := sqlite.EnsureDB(ctx, cfg.DBPath); err != nil {
		log.Fatalf("failed to init database: %v", err)
	}

	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	occurrenceStore := sqlite.NewOccurrenceStore(db)
	userStore := sqlite.NewUserStore(db)
	reminderStore := sqlite.NewReminderStore(db)

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
