package telegram

import (
	"context"
	"log"

	"naggingbot/internal/domain"
)

// OccurrenceCallbackHandler handles Done/Ignore callbacks for occurrences.
type OccurrenceCallbackHandler struct {
	occurrences domain.OccurrenceStore
	responder   Responder
}

func NewOccurrenceCallbackHandler(occurrences domain.OccurrenceStore, responder Responder) *OccurrenceCallbackHandler {
	return &OccurrenceCallbackHandler{occurrences: occurrences, responder: responder}
}

func (h *OccurrenceCallbackHandler) HandleCallback(ctx context.Context, cb *CallbackQuery) error {
	if cb == nil || cb.Data == "" {
		return nil
	}

	// Ignore noop callbacks from already-handled messages.
	if cb.Data == "noop" {
		return nil
	}

	action, occID, err := ParseOccurrenceCallback(cb.Data)
	if err != nil {
		log.Printf("telegram: bad callback data %q: %v", cb.Data, err)
		return nil
	}

	var status domain.OccurrenceStatus
	switch action {
	case "done":
		status = domain.OccurrenceDone
	case "ignore":
		status = domain.OccurrenceIgnored
	default:
		return nil
	}

	if err := h.occurrences.UpdateStatus(ctx, occID, status); err != nil {
		log.Printf("telegram: failed to update occurrence %d status: %v", occID, err)
		return nil
	}

	// Update message text and remove buttons.
	if cb.Message != nil && h.responder != nil {
		newText := BuildFinalText(cb.Message.Text, status)
		if err := h.responder.EditMessageText(ctx, cb.Message.Chat.ID, cb.Message.MessageID, newText, BuildFinalMarkup()); err != nil {
			log.Printf("telegram: failed to edit message text/markup: %v", err)
		}
	}
	return nil
}
