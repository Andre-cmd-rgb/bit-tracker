package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS entries (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	entry_date TEXT NOT NULL UNIQUE,
	started_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	raw_text TEXT NOT NULL DEFAULT '',
	cleaned_text TEXT NOT NULL DEFAULT '',
	mood INTEGER NOT NULL DEFAULT 0,
	study_minutes INTEGER NOT NULL DEFAULT 0,
	scroll_minutes INTEGER NOT NULL DEFAULT 0,
	project_minutes INTEGER NOT NULL DEFAULT 0,
	tags TEXT NOT NULL DEFAULT '',
	project_name TEXT NOT NULL DEFAULT '',
	project_note TEXT NOT NULL DEFAULT '',
	project_completed INTEGER NOT NULL DEFAULT 0,
	is_bad_day INTEGER NOT NULL DEFAULT 0,
	is_good_day INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_entries_date ON entries(entry_date);

CREATE TABLE IF NOT EXISTS ai_messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	entry_id INTEGER,
	mode TEXT NOT NULL,
	prompt TEXT NOT NULL,
	response TEXT NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY(entry_id) REFERENCES entries(id) ON DELETE SET NULL
);
`

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir data: %w", err)
	}
	db, err := sql.Open("sqlite", path+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("schema: %w", err)
	}
	return &Store{DB: db}, nil
}

func (s *Store) Close() error { return s.DB.Close() }

func Now() time.Time { return time.Now() }
