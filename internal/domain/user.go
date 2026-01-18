package domain

// User represents a Telegram user that interacts with the bot.
type User struct {
	ID         int64
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	Language   string
}
