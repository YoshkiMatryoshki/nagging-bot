package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	"naggingbot/internal/domain"
)

// ReminderStore implements domain.ReminderStore backed by SQLite.
type ReminderStore struct {
	db *sql.DB
}

func NewReminderStore(db *sql.DB) *ReminderStore {
	return &ReminderStore{db: db}
}

func (s *ReminderStore) GetByID(ctx context.Context, id int64) (*domain.Reminder, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, name, description, start_date_utc, end_date_utc, times_of_day, time_zone, is_active
		FROM reminders WHERE id = ?`, id)

	return scanReminder(row)
}

func (s *ReminderStore) ListByUser(ctx context.Context, userID int64) ([]*domain.Reminder, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, name, description, start_date_utc, end_date_utc, times_of_day, time_zone, is_active
		FROM reminders WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Reminder
	for rows.Next() {
		rem, err := scanReminder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rem)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *ReminderStore) Create(ctx context.Context, reminder *domain.Reminder) error {
	timesJSON, err := marshalTimes(reminder.TimesOfDay)
	if err != nil {
		return err
	}

	res, err := s.db.ExecContext(ctx, `
		INSERT INTO reminders (user_id, name, description, start_date_utc, end_date_utc, times_of_day, time_zone, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		reminder.UserID, reminder.Name, reminder.Description, reminder.StartDate, reminder.EndDate, timesJSON, reminder.TimeZone, boolToInt(reminder.IsActive))
	if err != nil {
		return err
	}

	if id, err := res.LastInsertId(); err == nil {
		reminder.ID = id
	}
	return nil
}

func (s *ReminderStore) Update(ctx context.Context, reminder *domain.Reminder) error {
	timesJSON, err := marshalTimes(reminder.TimesOfDay)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE reminders
		SET user_id = ?, name = ?, description = ?, start_date_utc = ?, end_date_utc = ?, times_of_day = ?, time_zone = ?, is_active = ?
		WHERE id = ?`,
		reminder.UserID, reminder.Name, reminder.Description, reminder.StartDate, reminder.EndDate, timesJSON, reminder.TimeZone, boolToInt(reminder.IsActive), reminder.ID)
	return err
}

func scanReminder(scanner interface {
	Scan(dest ...any) error
}) (*domain.Reminder, error) {
	var r domain.Reminder
	var timesJSON sql.NullString
	if err := scanner.Scan(&r.ID, &r.UserID, &r.Name, &r.Description, &r.StartDate, &r.EndDate, &timesJSON, &r.TimeZone, &r.IsActive); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if timesJSON.Valid && timesJSON.String != "" {
		var tod []domain.TimeOfDay
		if err := json.Unmarshal([]byte(timesJSON.String), &tod); err != nil {
			return nil, err
		}
		r.TimesOfDay = tod
	}

	return &r, nil
}

func marshalTimes(times []domain.TimeOfDay) (sql.NullString, error) {
	if len(times) == 0 {
		return sql.NullString{}, nil
	}
	b, err := json.Marshal(times)
	if err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{String: string(b), Valid: true}, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
