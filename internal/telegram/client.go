package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Handler processes a Telegram update.
type Handler interface {
	HandleUpdate(ctx context.Context, update Update) error
}

// Client polls Telegram updates using long polling.
type Client struct {
	token        string
	baseURL      string
	httpClient   *http.Client
	pollTimeout  time.Duration
	pollInterval time.Duration
}

// NewClient constructs a Telegram client.
func NewClient(token string, pollInterval, pollTimeout time.Duration) *Client {
	return &Client{
		token:       token,
		baseURL:     fmt.Sprintf("https://api.telegram.org/bot%s", token),
		pollTimeout: pollTimeout,
		httpClient: &http.Client{
			Timeout: pollTimeout + 5*time.Second,
		},
		pollInterval: pollInterval,
	}
}

// Poll starts long polling for updates and dispatches them to the handler.
func (c *Client) Poll(ctx context.Context, handler Handler) error {
	var offset int64

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		updates, err := c.getUpdates(ctx, offset)
		if err != nil {
			log.Printf("telegram polling error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		for _, u := range updates {
			offset = u.UpdateID + 1
			if err := handler.HandleUpdate(ctx, u); err != nil {
				log.Printf("telegram handler error: %v", err)
			}
		}

		// Cooldown between polling attempts.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(c.pollInterval):
		}
	}
}

func (c *Client) getUpdates(ctx context.Context, offset int64) ([]Update, error) {
	reqCtx, cancel := context.WithTimeout(ctx, c.httpClient.Timeout)
	defer cancel()

	url := fmt.Sprintf("%s/getUpdates?timeout=%d&offset=%d", c.baseURL, int(c.pollTimeout.Seconds()), offset)
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getUpdates status: %s", resp.Status)
	}

	var envelope struct {
		OK          bool     `json:"ok"`
		Result      []Update `json:"result"`
		Description string   `json:"description,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	if !envelope.OK {
		return nil, fmt.Errorf("telegram API error: %s", envelope.Description)
	}

	return envelope.Result, nil
}
