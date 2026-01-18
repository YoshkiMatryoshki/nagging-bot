package main

import (
	"context"
	"errors"
	"log"
	"time"

	"naggingbot/internal/app"
	"naggingbot/internal/domain"
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

	// Seed demo user.
	demoUser := &domain.User{
		TelegramID: 123456789,
		Username:   "demo_user",
		FirstName:  "Demo",
		LastName:   "User",
		Language:   "en",
	}
	_ = userStore.Upsert(ctx, demoUser)

	// Seed demo reminder for the user.
	demoReminder := &domain.Reminder{
		UserID:      demoUser.ID,
		Name:        "Vitamin D",
		Description: "Take 2 pills",
		StartDate:   time.Now().UTC(),
		EndDate:     time.Now().UTC().Add(24 * time.Hour),
		TimesOfDay: []domain.TimeOfDay{
			{Hour: 9, Minute: 0},
			{Hour: 13, Minute: 0},
			{Hour: 19, Minute: 0},
		},
		TimeZone: "UTC",
		IsActive: true,
	}
	_ = reminderStore.Create(ctx, demoReminder)

	now := time.Now().UTC()
	_ = occurrenceStore.Create(ctx, &domain.Occurrence{
		ReminderID: demoReminder.ID,
		FireAtUtc:  now,
		Status:     domain.OccurrenceCreated,
	})
	_ = occurrenceStore.Create(ctx, &domain.Occurrence{
		ReminderID: demoReminder.ID,
		FireAtUtc:  now.Add(5 * time.Second),
		Status:     domain.OccurrenceCreated,
	})
	_ = occurrenceStore.Create(ctx, &domain.Occurrence{
		ReminderID: demoReminder.ID,
		FireAtUtc:  now.Add(9 * time.Second),
		Status:     domain.OccurrenceCreated,
	})
	_ = occurrenceStore.Create(ctx, &domain.Occurrence{
		ReminderID: demoReminder.ID,
		FireAtUtc:  now.Add(10 * time.Second),
		Status:     domain.OccurrenceCreated,
	})
	_ = occurrenceStore.Create(ctx, &domain.Occurrence{
		ReminderID: demoReminder.ID,
		FireAtUtc:  now.Add(10 * time.Second),
		Status:     domain.OccurrenceCreated,
	})

	notifier := &scheduler.LoggingNotifier{}
	sched := scheduler.New(occurrenceStore, notifier, cfg.SchedulerInterval)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start Telegram polling in a separate goroutine.
	tgClient := telegram.NewClient(cfg.BotToken, cfg.PollInterval, cfg.PollTimeout)
	tgHandler := &telegram.StartHandler{}
	go func() {
		if err := tgClient.Poll(ctx, tgHandler); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("telegram poller stopped: %v", err)
		}
	}()

	if err := sched.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("scheduler stopped with error: %v", err)
	}
}
