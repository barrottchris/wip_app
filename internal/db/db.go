// Package db manages WIP's persistent, app-local SQLite database.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const databaseFileName = "wip.db"

// DB wraps the SQLite connection used by the application.
type DB struct {
	Conn *sql.DB
}

// Start opens WIP's database in the current user's application data directory.
// The directory is stable across restarts and does not require a server.
func Start() (*DB, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("locating user config directory: %w", err)
	}
	return StartAt(filepath.Join(configDir, "WIP", databaseFileName))
}

// StartAt opens a database at path. It is also used by tests to isolate their
// data from the user's real application database.
func StartAt(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening SQLite database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("pinging SQLite database: %w", err)
	}

	d := &DB{Conn: conn}

	if err := d.migrate(); err != nil {
		_ = d.Stop()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return d, nil
}

// Stop closes the SQLite connection. The database file remains on disk.
func (d *DB) Stop() error {
	if d.Conn == nil {
		return nil
	}
	return d.Conn.Close()
}

// migrate creates the schema if it doesn't already exist. For MVP this is
// a plain idempotent CREATE TABLE IF NOT EXISTS — a proper migration tool
// (e.g. golang-migrate) is worth introducing once the schema starts
// changing often.
func (d *DB) migrate() error {
	_, err := d.Conn.Exec(`
		CREATE TABLE IF NOT EXISTS apps (
			id              TEXT PRIMARY KEY,
			name            TEXT NOT NULL,
			description     TEXT NOT NULL DEFAULT '',
			stack           TEXT NOT NULL DEFAULT '[]',
			status          TEXT NOT NULL DEFAULT 'active',
			notes           TEXT NOT NULL DEFAULT '',
			local_path      TEXT NOT NULL DEFAULT '',
			repo_url        TEXT NOT NULL DEFAULT '',
			default_branch  TEXT NOT NULL DEFAULT '',
			branches        TEXT NOT NULL DEFAULT '[]',
			components      TEXT NOT NULL DEFAULT '[]',
			created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_touched_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			archived        BOOLEAN NOT NULL DEFAULT 0
		);

		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT ''
		);
	`)
	return err
}
