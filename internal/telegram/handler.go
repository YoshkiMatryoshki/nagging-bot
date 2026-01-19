package telegram

import (
	"context"
	"log"
	"strings"
	"time"

	"naggingbot/internal/domain"
)

// StartHandler handles /start messages: upserts the user, seeds a demo reminder, and demo occurrences.
type StartHandler struct {
	users       domain.UserStore
	reminders   domain.ReminderStore
	occurrences domain.OccurrenceStore
}

// NewStartHandler constructs a StartHandler with required stores.
func NewStartHandler(users domain.UserStore, reminders domain.ReminderStore, occurrences domain.OccurrenceStore) *StartHandler {
	return &StartHandler{
		users:       users,
		reminders:   reminders,
		occurrences: occurrences,
	}
}

func (h *StartHandler) HandleUpdate(ctx context.Context, update Update) error {
	if update.Message == nil {
		return nil
	}

	msg := update.Message
	if !strings.HasPrefix(msg.Text, "/start") {
		return nil
	}

	user := msg.From
	if user == nil {
		log.Printf("telegram: /start received but user is nil (chat_id=%d)", msg.Chat.ID)
		return nil
	}

	domainUser, err := h.ensureUser(ctx, user)
	if err != nil {
		log.Printf("telegram: failed to upsert user %d: %v", user.ID, err)
		return nil
	}

	if err := h.createDemoReminder(ctx, domainUser); err != nil {
		log.Printf("telegram: failed to create demo reminder for user %d: %v", user.ID, err)
		return nil
	}

	log.Printf("telegram: /start handled for user_id=%d username=%s first=%s last=%s lang=%s",
		user.ID, user.Username, user.FirstName, user.LastName, user.LanguageCode)
	return nil
}

func (h *StartHandler) ensureUser(ctx context.Context, u *User) (*domain.User, error) {
	domainUser := &domain.User{
		TelegramID: u.ID,
		Username:   u.Username,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Language:   u.LanguageCode,
	}
	if err := h.users.Upsert(ctx, domainUser); err != nil {
		return nil, err
	}
	return domainUser, nil
}

func (h *StartHandler) createDemoReminder(ctx context.Context, u *domain.User) error {

	now := time.Now().UTC()
	start := now.Add(1 * time.Minute)
	end := now.Add(10 * time.Minute)

	rem := &domain.Reminder{
		UserID:      u.ID,
		Name:        "Demo reminder",
		Description: "Demo occurrences every minute",
		StartDate:   start,
		EndDate:     end,
		TimesOfDay:  nil,
		TimeZone:    "UTC",
		IsActive:    true,
	}
	if err := h.reminders.Create(ctx, rem); err != nil {
		return err
	}

	for t := start; !t.After(end); t = t.Add(time.Second * 30) {
		occ := &domain.Occurrence{
			ReminderID: rem.ID,
			FireAtUtc:  t,
			Status:     domain.OccurrenceCreated,
		}
		if err := h.occurrences.Create(ctx, occ); err != nil {
			log.Printf("telegram: failed to create occurrence at %s: %v", t, err)
		}
	}

	return nil
}
