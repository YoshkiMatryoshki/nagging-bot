package telegram

import (
	"context"
	"log"
	"strconv"
	"strings"

	"naggingbot/internal/domain"
)

// DeleteHandler handles /delete <id> to remove reminder and occurrences.
type DeleteHandler struct {
	users       domain.UserStore
	reminders   domain.ReminderStore
	occurrences domain.OccurrenceStore
	responder   Responder
}

func NewDeleteHandler(users domain.UserStore, reminders domain.ReminderStore, occurrences domain.OccurrenceStore, responder Responder) *DeleteHandler {
	return &DeleteHandler{
		users:       users,
		reminders:   reminders,
		occurrences: occurrences,
		responder:   responder,
	}
}

func (h *DeleteHandler) HandleCommand(ctx context.Context, msg *Message) error {
	user := msg.From
	if user == nil {
		return nil
	}

	parts := strings.Split(strings.TrimSpace(msg.Text), " ")
	if len(parts) != 2 {
		h.reply(ctx, user.ID, "Usage: /delete <reminder_id>")
		return nil
	}

	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		h.reply(ctx, user.ID, "Invalid id")
		return nil
	}

	domainUser, err := h.users.GetByTelegramID(ctx, user.ID)
	if err != nil {
		log.Printf("telegram: delete fetch user failed: %v", err)
		h.reply(ctx, user.ID, "Failed to delete")
		return nil
	}
	if domainUser == nil {
		h.reply(ctx, user.ID, "Reminder not found")
		return nil
	}

	rem, err := h.reminders.GetByID(ctx, id)
	if err != nil {
		log.Printf("telegram: delete get reminder failed: %v", err)
		h.reply(ctx, user.ID, "Failed to delete")
		return nil
	}
	if rem == nil {
		h.reply(ctx, user.ID, "Reminder not found")
		return nil
	}
	if rem.UserID != domainUser.ID {
		h.reply(ctx, user.ID, "Cannot delete reminder of another user")
		return nil
	}

	if err := h.occurrences.DeleteByReminder(ctx, id); err != nil {
		log.Printf("telegram: delete occurrences failed: %v", err)
		h.reply(ctx, user.ID, "Failed to delete occurrences")
		return nil
	}
	if err := h.reminders.DeleteByID(ctx, id); err != nil {
		log.Printf("telegram: delete reminder failed: %v", err)
		h.reply(ctx, user.ID, "Failed to delete reminder")
		return nil
	}

	h.reply(ctx, user.ID, "Reminder deleted")
	return nil
}

func (h *DeleteHandler) reply(ctx context.Context, chatID int64, text string) {
	if h.responder == nil {
		return
	}
	if err := h.responder.SendMessage(ctx, chatID, text); err != nil {
		log.Printf("telegram: failed to send delete reply: %v", err)
	}
}
