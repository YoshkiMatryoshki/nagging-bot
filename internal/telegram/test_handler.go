package telegram

import (
	"context"
	"fmt"
	"log"
	"time"

	"naggingbot/internal/domain"
)

// TestHandler creates a demo reminder for a specific allowed user.
type TestHandler struct {
	users       domain.UserStore
	reminders   domain.ReminderStore
	occurrences domain.OccurrenceStore
	responder   Responder
	allowedUser int64
}

func NewTestHandler(users domain.UserStore, reminders domain.ReminderStore, occurrences domain.OccurrenceStore, responder Responder, allowedUser int64) *TestHandler {
	return &TestHandler{
		users:       users,
		reminders:   reminders,
		occurrences: occurrences,
		responder:   responder,
		allowedUser: allowedUser,
	}
}

func (h *TestHandler) HandleCommand(ctx context.Context, msg *Message) error {
	user := msg.From
	if user == nil {
		return nil
	}
	if user.ID != h.allowedUser {
		h.reply(ctx, user.ID, "Unauthorized for /test")
		return nil
	}

	// Ensure user exists.
	domainUser := &domain.User{
		TelegramID: user.ID,
		Username:   user.Username,
		FirstName:  user.FirstName,
		LastName:   user.LastName,
		Language:   user.LanguageCode,
	}
	if err := h.users.Upsert(ctx, domainUser); err != nil {
		log.Printf("telegram: /test upsert user failed: %v", err)
		h.reply(ctx, user.ID, "Failed to upsert user")
		return nil
	}

	now := time.Now().UTC()
	start := now.Add(2 * time.Second)
	end := now.Add(40 * time.Second)

	rem := &domain.Reminder{
		UserID:      domainUser.ID,
		Name:        "Demo reminder",
		Description: "Demo occurrences every 10 seconds",
		StartDate:   start,
		EndDate:     end,
		TimeZone:    "UTC",
		IsActive:    true,
	}
	if err := h.reminders.Create(ctx, rem); err != nil {
		log.Printf("telegram: /test create reminder failed: %v", err)
		h.reply(ctx, user.ID, "Failed to create demo reminder")
		return nil
	}

	for t := start; !t.After(end); t = t.Add(10 * time.Second) {
		occ := &domain.Occurrence{
			ReminderID: rem.ID,
			FireAtUtc:  t,
			Status:     domain.OccurrenceCreated,
		}
		if err := h.occurrences.Create(ctx, occ); err != nil {
			log.Printf("telegram: failed to create occurrence at %s: %v", t, err)
		}
	}

	h.reply(ctx, user.ID, fmt.Sprintf("Demo reminder created with occurrences until %s", end.Format(time.RFC3339)))
	return nil
}

func (h *TestHandler) reply(ctx context.Context, chatID int64, text string) {
	if h.responder == nil {
		return
	}
	if err := h.responder.SendMessage(ctx, chatID, text); err != nil {
		log.Printf("telegram: failed to send reply: %v", err)
	}
}
