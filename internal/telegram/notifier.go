package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"naggingbot/internal/domain"
	"naggingbot/internal/scheduler"
)

// Notifier sends messages to Telegram chats.
type Notifier struct {
	token     string
	users     domain.UserStore
	httpClient *http.Client
}

// NewNotifier constructs a Telegram notifier.
func NewNotifier(token string, users domain.UserStore) *Notifier {
	return &Notifier{
		token: token,
		users: users,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (n *Notifier) Send(ctx context.Context, occ scheduler.OccurrenceWithReminder) error {
	if occ.Reminder == nil {
		return fmt.Errorf("telegram notifier: missing reminder for occurrence %d", occ.Occurrence.ID)
	}

	user, err := n.users.GetByID(ctx, occ.Reminder.UserID)
	if err != nil {
		return fmt.Errorf("telegram notifier: get user %d: %w", occ.Reminder.UserID, err)
	}
	if user == nil || user.TelegramID == 0 {
		return fmt.Errorf("telegram notifier: no telegram id for user %d", occ.Reminder.UserID)
	}

	text := fmt.Sprintf("Reminder: %s\n%s\nOccurrence #%d at %s",
		occ.Reminder.Name, occ.Reminder.Description, occ.Occurrence.ID, occ.Occurrence.FireAtUtc.Format(time.RFC3339))

	// Inline keyboard with Done / Ignore.
	replyMarkup := BuildInitialMarkup(occ.Occurrence.ID)

	payload := map[string]any{
		"chat_id": user.TelegramID,
		"text":    text,
		"reply_markup": replyMarkup,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram notifier: sendMessage status %s", resp.Status)
	}
	return nil
}
