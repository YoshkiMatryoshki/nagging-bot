package sqlite

import (
	"context"
	"database/sql"
	"time"

	"naggingbot/internal/domain"
)

// OccurrenceStore implements domain.OccurrenceStore backed by SQLite.
type OccurrenceStore struct {
	db *sql.DB
}

func NewOccurrenceStore(db *sql.DB) *OccurrenceStore {
	return &OccurrenceStore{db: db}
}

func (s *OccurrenceStore) GetByID(ctx context.Context, id int64) (*domain.Occurrence, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, reminder_id, fire_at_utc, status
		FROM occurrences WHERE id = ?`, id)

	var occ domain.Occurrence
	if err := row.Scan(&occ.ID, &occ.ReminderID, &occ.FireAtUtc, &occ.Status); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &occ, nil
}

func (s *OccurrenceStore) ListByReminder(ctx context.Context, reminderID int64) ([]*domain.Occurrence, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, reminder_id, fire_at_utc, status
		FROM occurrences WHERE reminder_id = ?`, reminderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Occurrence
	for rows.Next() {
		var occ domain.Occurrence
		if err := rows.Scan(&occ.ID, &occ.ReminderID, &occ.FireAtUtc, &occ.Status); err != nil {
			return nil, err
		}
		out = append(out, &occ)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OccurrenceStore) ListPendingInRange(ctx context.Context, startUTC, endUTC time.Time) ([]*domain.Occurrence, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, reminder_id, fire_at_utc, status
		FROM occurrences
		WHERE status = ?
		  AND fire_at_utc >= ?
		  AND fire_at_utc <= ?`,
		domain.OccurrenceCreated, startUTC, endUTC)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Occurrence
	for rows.Next() {
		var occ domain.Occurrence
		if err := rows.Scan(&occ.ID, &occ.ReminderID, &occ.FireAtUtc, &occ.Status); err != nil {
			return nil, err
		}
		out = append(out, &occ)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *OccurrenceStore) Create(ctx context.Context, occ *domain.Occurrence) error {
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO occurrences (reminder_id, fire_at_utc, status)
		VALUES (?, ?, ?)`,
		occ.ReminderID, occ.FireAtUtc, occ.Status)
	if err != nil {
		return err
	}
	if id, err := res.LastInsertId(); err == nil {
		occ.ID = id
	}
	return nil
}

func (s *OccurrenceStore) UpdateStatus(ctx context.Context, id int64, status domain.OccurrenceStatus) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE occurrences SET status = ? WHERE id = ?`, status, id)
	return err
}

func (s *OccurrenceStore) DeleteByReminder(ctx context.Context, reminderID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM occurrences WHERE reminder_id = ?`, reminderID)
	return err
}
