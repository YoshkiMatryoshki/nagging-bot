package telegram

import (
	"context"
	"log"
	"strings"
)

// CommandHandler processes bot commands (e.g., "/start").
type CommandHandler interface {
	HandleCommand(ctx context.Context, msg *Message) error
}

// CallbackHandler processes callback queries (e.g., inline buttons).
type CallbackHandler interface {
	HandleCallback(ctx context.Context, cb *CallbackQuery) error
}

// Dispatcher routes updates to command or callback handlers.
type Dispatcher struct {
	commands map[string]CommandHandler
	callback CallbackHandler
}

// NewDispatcher constructs a dispatcher with optional handlers.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		commands: make(map[string]CommandHandler),
	}
}

// HandleUpdate allows Dispatcher to satisfy the Client Handler interface.
func (d *Dispatcher) HandleUpdate(ctx context.Context, update Update) error {
	d.Dispatch(ctx, update)
	return nil
}

// RegisterCommand registers a handler for a given command (e.g., "/start").
func (d *Dispatcher) RegisterCommand(cmd string, h CommandHandler) {
	d.commands[cmd] = h
}

// RegisterCallback sets the handler for callback queries.
func (d *Dispatcher) RegisterCallback(h CallbackHandler) {
	d.callback = h
}

// Dispatch routes the update to the appropriate handler.
func (d *Dispatcher) Dispatch(ctx context.Context, update Update) {
	// Callback query has priority.
	if update.CallbackQuery != nil && d.callback != nil {
		if err := d.callback.HandleCallback(ctx, update.CallbackQuery); err != nil {
			log.Printf("telegram callback handler error: %v", err)
		}
		return
	}

	// Commands in messages.
	if update.Message != nil {
		text := strings.TrimSpace(update.Message.Text)
		if strings.HasPrefix(text, "/") {
			cmd := firstToken(text)
			if h, ok := d.commands[cmd]; ok {
				if err := h.HandleCommand(ctx, update.Message); err != nil {
					log.Printf("telegram command handler error (%s): %v", cmd, err)
				}
			}
		}
	}
}

func firstToken(s string) string {
	if idx := strings.IndexAny(s, " \t\r\n"); idx >= 0 {
		return s[:idx]
	}
	return s
}
