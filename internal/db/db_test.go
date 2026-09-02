package db

import (
	"path/filepath"
	"testing"
)

func TestStartAtPersistsSchemaAcrossReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "wip.db")

	first, err := StartAt(path)
	if err != nil {
		t.Fatalf("StartAt() failed: %v", err)
	}
	if _, err := first.Conn.Exec(`INSERT INTO settings (key, value) VALUES ('test', 'persisted')`); err != nil {
		t.Fatalf("insert failed: %v", err)
	}
	if err := first.Stop(); err != nil {
		t.Fatalf("first Stop() failed: %v", err)
	}

	second, err := StartAt(path)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	defer second.Stop()

	var value string
	if err := second.Conn.QueryRow(`SELECT value FROM settings WHERE key = 'test'`).Scan(&value); err != nil {
		t.Fatalf("persisted value was not readable: %v", err)
	}
	if value != "persisted" {
		t.Fatalf("persisted value = %q, want %q", value, "persisted")
	}
}