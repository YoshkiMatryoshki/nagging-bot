package telegram

import (
	"context"
	"log"
	"strings"
)

// StartHandler handles /start messages, logs user info, and ignores other messages.
type StartHandler struct{}

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

	log.Printf("telegram: /start from user_id=%d username=%s first=%s last=%s lang=%s",
		user.ID, user.Username, user.FirstName, user.LastName, user.LanguageCode)
	return nil
}
