package main

import (
	"context"
	"errors"
	"log"
	"time"

	"naggingbot/internal/domain"
	"naggingbot/internal/scheduler"
	"naggingbot/internal/storage/memory"
)

func main() {
	log.Println("NaggingBot starting up (scheduler demo)")

	now := time.Now().UTC()
	occurrenceStore := memory.NewInMemoryOccurrenceStore()
	_ = occurrenceStore.Create(context.Background(), &domain.Occurrence{
		ReminderID: 1,
		FireAtUtc:  now,
		Status:     domain.OccurrenceCreated,
	})
	_ = occurrenceStore.Create(context.Background(), &domain.Occurrence{
		ReminderID: 1,
		FireAtUtc:  now.Add(5 * time.Second),
		Status:     domain.OccurrenceCreated,
	})
	_ = occurrenceStore.Create(context.Background(), &domain.Occurrence{
		ReminderID: 1,
		FireAtUtc:  now.Add(9 * time.Second),
		Status:     domain.OccurrenceCreated,
	})
	_ = occurrenceStore.Create(context.Background(), &domain.Occurrence{
		ReminderID: 1,
		FireAtUtc:  now.Add(10 * time.Second),
		Status:     domain.OccurrenceCreated,
	})
	_ = occurrenceStore.Create(context.Background(), &domain.Occurrence{
		ReminderID: 1,
		FireAtUtc:  now.Add(10 * time.Second),
		Status:     domain.OccurrenceCreated,
	})

	notifier := &scheduler.LoggingNotifier{}
	sched := scheduler.New(occurrenceStore, notifier, time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		time.Sleep(13 * time.Second)
		cancel()
	}()

	if err := sched.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Printf("scheduler stopped with error: %v", err)
	}
}
