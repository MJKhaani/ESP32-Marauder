package store

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// AppSettings represents persisted user preferences.
type AppSettings struct {
	Theme            string `json:"theme"`
	AutoReconnect    bool   `json:"auto_reconnect"`
	DefaultBaud      int    `json:"default_baud"`
	SanitizeReports  bool   `json:"sanitize_reports"`
	RetentionDays    int    `json:"retention_days"`
}

// Store wraps SQLite operations for Marauder Studio.
type Store struct {
	mu sync.RWMutex
	db *sql.DB
}

// NewStore initializes a SQLite database connection and runs migrations.
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening sqlite database: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed running database migrations: %w", err)
	}

	return s, nil
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS devices (
			id TEXT PRIMARY KEY,
			nickname TEXT,
			board_model TEXT,
			chip_family TEXT,
			firmware_version TEXT,
			first_seen DATETIME,
			last_seen DATETIME
		);`,
		`CREATE TABLE IF NOT EXISTS survey_snapshots (
			id TEXT PRIMARY KEY,
			device_id TEXT,
			timestamp DATETIME,
			ap_count INTEGER,
			sta_count INTEGER,
			data_json TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			timestamp DATETIME,
			operator TEXT,
			lab TEXT,
			command TEXT,
			targets_json TEXT,
			duration_seconds INTEGER,
			approved INTEGER
		);`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		);`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// SaveDevice upserts a device entry into SQLite.
func (s *Store) SaveDevice(id, nickname, board, chip, fw string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	query := `INSERT INTO devices (id, nickname, board_model, chip_family, firmware_version, first_seen, last_seen)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			nickname = coalesce(excluded.nickname, devices.nickname),
			board_model = excluded.board_model,
			chip_family = excluded.chip_family,
			firmware_version = excluded.firmware_version,
			last_seen = excluded.last_seen;`

	now := time.Now()
	_, err := s.db.Exec(query, id, nickname, board, chip, fw, now, now)
	return err
}

// SaveAuditLog persists an audit record.
func (s *Store) SaveAuditLog(id, operator, lab, command, targetsJSON string, durationSec int, approved bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	apprInt := 0
	if approved {
		apprInt = 1
	}

	query := `INSERT INTO audit_logs (id, timestamp, operator, lab, command, targets_json, duration_seconds, approved)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?);`

	_, err := s.db.Exec(query, id, time.Now(), operator, lab, command, targetsJSON, durationSec, apprInt)
	return err
}

// Close closes the database connection.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.db.Close()
}
