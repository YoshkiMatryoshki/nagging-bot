package sqlite

import (
	"context"
	"database/sql"

	"naggingbot/internal/domain"
)

// UserStore implements domain.UserStore backed by SQLite.
type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, language
		FROM users WHERE id = ?`, id)

	var u domain.User
	if err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &u.Language); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) GetByTelegramID(ctx context.Context, telegramID int64) (*domain.User, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, telegram_id, username, first_name, last_name, language
		FROM users WHERE telegram_id = ?`, telegramID)

	var u domain.User
	if err := row.Scan(&u.ID, &u.TelegramID, &u.Username, &u.FirstName, &u.LastName, &u.Language); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) Upsert(ctx context.Context, user *domain.User) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO users (telegram_id, username, first_name, last_name, language)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(telegram_id) DO UPDATE SET
			username = excluded.username,
			first_name = excluded.first_name,
			last_name = excluded.last_name,
			language = excluded.language
	`, user.TelegramID, user.Username, user.FirstName, user.LastName, user.Language)
	if err != nil {
		return err
	}

	// Reload to populate ID.
	reloaded, err := s.GetByTelegramID(ctx, user.TelegramID)
	if err != nil {
		return err
	}
	if reloaded != nil {
		user.ID = reloaded.ID
	}
	return nil
}
