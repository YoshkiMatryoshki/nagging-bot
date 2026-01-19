package telegram

import (
	"strings"

	"naggingbot/internal/domain"
)

const (
	textDoneInit    = "✅ Done"
	textIgnoreInit  = "🚫 Ignore"
	textDoneFinal   = "Status: ✅ Done"
	textIgnoreFinal = "Status: 🚫 Ignored"
)

// BuildInitialMarkup returns the inline keyboard for initial Done/Ignore.
func BuildInitialMarkup(occID int64) map[string]any {
	doneData := BuildOccurrenceCallback(occID, OccurrenceActionDone)
	ignoreData := BuildOccurrenceCallback(occID, OccurrenceActionIgnore)
	return map[string]any{
		"inline_keyboard": [][]map[string]any{
			{
				{"text": textDoneInit, "callback_data": doneData},
				{"text": textIgnoreInit, "callback_data": ignoreData},
			},
		},
	}
}

// BuildFinalMarkup removes buttons after handling.
func BuildFinalMarkup() map[string]any {
	return map[string]any{
		"inline_keyboard": [][]map[string]any{},
	}
}

// BuildFinalText appends status line to the original message text if not present.
func BuildFinalText(original string, status domain.OccurrenceStatus) string {
	switch status {
	case domain.OccurrenceDone:
		return appendStatus(original, textDoneFinal)
	case domain.OccurrenceIgnored:
		return appendStatus(original, textIgnoreFinal)
	default:
		return original
	}
}

func appendStatus(text, statusLine string) string {
	if statusLine == "" {
		return text
	}
	if strings.Contains(text, statusLine) {
		return text
	}
	return text + "\n\n" + statusLine
}
