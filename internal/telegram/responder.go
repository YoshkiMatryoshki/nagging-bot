package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Responder can edit messages (e.g., update inline keyboards).
type Responder interface {
	EditMessageReplyMarkup(ctx context.Context, chatID int64, messageID int64, markup any) error
	EditMessageText(ctx context.Context, chatID int64, messageID int64, text string, markup any) error
}

type httpResponder struct {
	token      string
	httpClient *http.Client
}

// NewHTTPResponder constructs a responder using Telegram Bot API.
func NewHTTPResponder(token string) Responder {
	return &httpResponder{
		token: token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *httpResponder) EditMessageReplyMarkup(ctx context.Context, chatID int64, messageID int64, markup any) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageReplyMarkup", r.token)
	payload := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"reply_markup": markup,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("editMessageReplyMarkup status %s", resp.Status)
	}
	return nil
}

func (r *httpResponder) EditMessageText(ctx context.Context, chatID int64, messageID int64, text string, markup any) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", r.token)
	payload := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
	}
	if markup != nil {
		payload["reply_markup"] = markup
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("editMessageText status %s", resp.Status)
	}
	return nil
}
