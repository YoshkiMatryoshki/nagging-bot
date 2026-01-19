package telegram

import (
	"context"
	"log"

	"naggingbot/internal/domain"
)

// StartHandler handles /start messages: upserts the user, seeds a demo reminder, and demo occurrences.
type StartHandler struct {
	users       domain.UserStore
	responder   Responder
}

// NewStartHandler constructs a StartHandler with required stores.
func NewStartHandler(users domain.UserStore, responder Responder) *StartHandler {
	return &StartHandler{
		users:     users,
		responder: responder,
	}
}

// HandleCommand processes the /start command.
func (h *StartHandler) HandleCommand(ctx context.Context, msg *Message) error {
	user := msg.From
	if user == nil {
		log.Printf("telegram: /start received but user is nil (chat_id=%d)", msg.Chat.ID)
		return nil
	}

	if _, err := h.ensureUser(ctx, user); err != nil {
		log.Printf("telegram: failed to upsert user %d: %v", user.ID, err)
		return nil
	}

	if h.responder != nil {
		msg := "You are registered.\n\nCommands:\n" +
			"/reminder <name>_<description>_<DD.MM.YYYY>_<DD.MM.YYYY>_<HH:MM;HH:MM>_<IANA timezone> - create reminder\n" +
			"/list - list latest reminders (up to 20)\n" +
			"/delete <id> - delete reminder and occurrences\n" +
			"/test - create demo reminder (restricted)\n\n" +
			"Example:\n/reminder Pill_VitC_19.01.2026_20.01.2026_08:00;13:00;19:00_Europe/Warsaw"
		if err := h.responder.SendMessage(ctx, user.ID, msg); err != nil {
			log.Printf("telegram: failed to send start ack: %v", err)
		}
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
