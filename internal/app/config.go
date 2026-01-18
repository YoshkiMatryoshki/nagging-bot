package app

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Config holds runtime settings loaded from environment variables.
type Config struct {
	BotToken          string
	DBPath            string
	PollInterval      time.Duration
	PollTimeout       time.Duration
	SchedulerInterval time.Duration
}

// LoadConfig reads environment variables and validates them.
// Required env vars:
//   BOT_TOKEN            - Telegram bot token
//   DB_PATH              - path to SQLite database file
// Optional with defaults:
//   POLL_INTERVAL        - Cooldown between polling attempts (default: 30s)
//   POLL_TIMEOUT         - Long-poll timeout per request (default: 10s)
//   SCHEDULER_INTERVAL   - Scheduler tick interval (default: 1s)
func LoadConfig() (Config, error) {
	// Best-effort load from repo-root .env: assumes running from cmd/bot (cwd two levels below repo root).
	if err := loadEnvFile(".env"); err != nil {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg := Config{
		BotToken: os.Getenv("BOT_TOKEN"),
		DBPath:   os.Getenv("DB_PATH"),
	}

	cfg.PollInterval = time.Second * 30
	cfg.PollTimeout = time.Second * 10
	cfg.SchedulerInterval = time.Second

	if v := os.Getenv("POLL_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid POLL_INTERVAL: %w", err)
		}
		cfg.PollInterval = d
	}

	if v := os.Getenv("POLL_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid POLL_TIMEOUT: %w", err)
		}
		cfg.PollTimeout = d
	}

	if v := os.Getenv("SCHEDULER_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid SCHEDULER_INTERVAL: %w", err)
		}
		cfg.SchedulerInterval = d
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// loadEnvFile populates process env vars from a file formatted as KEY=VALUE per line.
// It ignores blank lines and lines starting with '#'. Missing file is not an error.
func loadEnvFile(path string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	repoRoot := filepath.Dir(filepath.Dir(cwd))
	envPath := filepath.Join(repoRoot, path)

	f, err := os.Open(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid env line %q", line)
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			return fmt.Errorf("invalid env line %q: empty key", line)
		}
		if err := os.Setenv(key, val); err != nil {
			return fmt.Errorf("set env %s: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// findEnvFile walks up from the current working directory to locate the first existing file named path.
func findEnvFile(path string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		candidate := filepath.Join(dir, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", nil
}

func (c Config) validate() error {
	var problems []string

	if strings.TrimSpace(c.BotToken) == "" {
		problems = append(problems, "BOT_TOKEN is required")
	}
	if strings.TrimSpace(c.DBPath) == "" {
		problems = append(problems, "DB_PATH is required")
	}
	if c.PollInterval <= 0 {
		problems = append(problems, "POLL_INTERVAL must be > 0")
	}
	if c.PollTimeout <= 0 {
		problems = append(problems, "POLL_TIMEOUT must be > 0")
	}
	if c.SchedulerInterval <= 0 {
		problems = append(problems, "SCHEDULER_INTERVAL must be > 0")
	}

	if len(problems) > 0 {
		return fmt.Errorf("invalid config: %s", strings.Join(problems, "; "))
	}
	return nil
}
