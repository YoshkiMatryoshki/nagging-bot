package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SetBotCommands registers bot commands for Telegram clients.
func SetBotCommands(ctx context.Context, token string) error {
	payload := map[string]any{
		"commands": []map[string]string{
			{"command": "start", "description": "Register"},
			{"command": "reminder", "description": "Create reminder"},
			{"command": "list", "description": "List reminders"},
			{"command": "delete", "description": "Delete reminder"},
			{"command": "test", "description": "Demo reminder (restricted)"},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("setMyCommands status %s", resp.Status)
	}

	return nil
}
