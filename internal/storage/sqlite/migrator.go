package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    telegram_id INTEGER NOT NULL UNIQUE,
    username TEXT,
    first_name TEXT,
    last_name TEXT,
    language TEXT
);

CREATE TABLE reminders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    start_date_utc DATETIME NOT NULL,
    end_date_utc DATETIME NOT NULL,
    times_of_day TEXT,
    time_zone TEXT NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE occurrences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    reminder_id INTEGER NOT NULL,
    fire_at_utc DATETIME NOT NULL,
    status INTEGER NOT NULL,
    FOREIGN KEY (reminder_id) REFERENCES reminders(id)
);

CREATE INDEX idx_occurrence_reminder ON occurrences(reminder_id);
CREATE INDEX idx_occurrence_fire_at ON occurrences(fire_at_utc);
`

// EnsureDB creates the SQLite database and applies initial schema if the file does not exist.
// If the file already exists, nothing is applied.
func EnsureDB(ctx context.Context, path string) error {
	if path == "" {
		return fmt.Errorf("db path is empty")
	}

	// If DB file exists, skip migrations.
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	// Ensure directory exists.
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create db dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open sqlite db: %w", err)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping sqlite db: %w", err)
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}

	return nil
}
