package telegram

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"naggingbot/internal/domain"
)

// ListHandler handles /list to show user reminders (limited to 20).
type ListHandler struct {
	users     domain.UserStore
	reminders domain.ReminderStore
	responder Responder
}

func NewListHandler(users domain.UserStore, reminders domain.ReminderStore, responder Responder) *ListHandler {
	return &ListHandler{users: users, reminders: reminders, responder: responder}
}

func (h *ListHandler) HandleCommand(ctx context.Context, msg *Message) error {
	user := msg.From
	if user == nil {
		return nil
	}

	domainUser, err := h.users.GetByTelegramID(ctx, user.ID)
	if err != nil {
		log.Printf("telegram: list fetch user failed: %v", err)
		return nil
	}
	if domainUser == nil {
		h.reply(ctx, user.ID, "No reminders found.")
		return nil
	}

	rems, err := h.reminders.ListByUser(ctx, domainUser.ID)
	if err != nil {
		log.Printf("telegram: list reminders failed: %v", err)
		return nil
	}

	if len(rems) == 0 {
		h.reply(ctx, user.ID, "No reminders found.")
		return nil
	}

	// Sort by ID desc.
	sort.Slice(rems, func(i, j int) bool {
		return rems[i].ID > rems[j].ID
	})
	if len(rems) > 20 {
		rems = rems[:20]
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Your reminders (latest up to 20):\n")
	for _, r := range rems {
		fmt.Fprintf(&b, "#%d: %s | %s | %s to %s | TZ=%s | Times=%s\n",
			r.ID, r.Name, r.Description, r.StartDate.Format("02.01.2006"), r.EndDate.Format("02.01.2006"), r.TimeZone, formatTimes(r.TimesOfDay))
	}

	h.reply(ctx, user.ID, b.String())
	return nil
}

func (h *ListHandler) reply(ctx context.Context, chatID int64, text string) {
	if h.responder == nil {
		return
	}
	if err := h.responder.SendMessage(ctx, chatID, text); err != nil {
		log.Printf("telegram: failed to send list reply: %v", err)
	}
}

func formatTimes(t []domain.TimeOfDay) string {
	if len(t) == 0 {
		return "n/a"
	}
	parts := make([]string, 0, len(t))
	for _, v := range t {
		parts = append(parts, fmt.Sprintf("%02d:%02d", v.Hour, v.Minute))
	}
	return strings.Join(parts, ";")
}
