package db

import (
	"path/filepath"
	"testing"
)

func TestOpenCreatesSchema(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(filepath.Join(dir, "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// entries table should exist
	if _, err := store.DB.Exec(`INSERT INTO entries(entry_date, started_at, updated_at) VALUES('2026-04-20','2026-04-20T07:14:00Z','2026-04-20T07:14:00Z')`); err != nil {
		t.Fatalf("insert: %v", err)
	}
	// ai_messages table should exist
	if _, err := store.DB.Exec(`INSERT INTO ai_messages(mode, prompt, response, created_at) VALUES('rewrite','p','r','2026-04-20T07:14:00Z')`); err != nil {
		t.Fatalf("insert ai: %v", err)
	}
}
