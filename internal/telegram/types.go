package telegram

// Update represents a Telegram update payload.
type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message,omitempty"`
}

// Message represents a Telegram message.
type Message struct {
	MessageID int64   `json:"message_id"`
	From      *User   `json:"from,omitempty"`
	Chat      Chat    `json:"chat"`
	Date      int64   `json:"date"`
	Text      string  `json:"text,omitempty"`
	Entities  []Entity `json:"entities,omitempty"`
}

// Entity represents message entities (e.g., bot commands).
type Entity struct {
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Type   string `json:"type"`
}

// Chat represents Telegram chat.
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// User represents Telegram user info from the message.
type User struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}
