// Package db manages an embedded Postgres instance and the schema WIP
// stores its data in. "Embedded" here means: a real Postgres binary is
// downloaded (once) and run locally by this Go process — no separate
// Postgres install or server needed. The schema and SQL are genuine
// Postgres, so pointing WIP at a real hosted Postgres later is just a
// connection-string change, not a rewrite.
package db

import (
	"database/sql"
	"fmt"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	_ "github.com/lib/pq"
)

const (
	dbPort     = 34199 // arbitrary local port for the embedded Postgres instance
	dbUser     = "wip"
	dbPassword = "wip"
	dbName     = "wip"
)

// DB wraps the embedded Postgres process and the SQL connection to it.
type DB struct {
	postgres *embeddedpostgres.EmbeddedPostgres
	Conn     *sql.DB
}

// Start launches the embedded Postgres instance (downloading the binary on
// first run only — cached after that) and returns a ready-to-use DB with
// the schema already applied.
func Start() (*DB, error) {
	postgres := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(dbPort).
			Username(dbUser).
			Password(dbPassword).
			Database(dbName),
	)

	if err := postgres.Start(); err != nil {
		return nil, fmt.Errorf("starting embedded postgres: %w", err)
	}

	connStr := fmt.Sprintf(
		"host=localhost port=%d user=%s password=%s dbname=%s sslmode=disable",
		dbPort, dbUser, dbPassword, dbName,
	)

	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		_ = postgres.Stop()
		return nil, fmt.Errorf("opening connection: %w", err)
	}

	if err := conn.Ping(); err != nil {
		_ = postgres.Stop()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	d := &DB{postgres: postgres, Conn: conn}

	if err := d.migrate(); err != nil {
		_ = d.Stop()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return d, nil
}

// Stop cleanly shuts down the connection and the embedded Postgres process.
// Call this on application exit.
func (d *DB) Stop() error {
	if d.Conn != nil {
		_ = d.Conn.Close()
	}
	if d.postgres != nil {
		return d.postgres.Stop()
	}
	return nil
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
			stack           JSONB NOT NULL DEFAULT '[]',
			status          TEXT NOT NULL DEFAULT 'active',
			notes           TEXT NOT NULL DEFAULT '',
			local_path      TEXT NOT NULL DEFAULT '',
			repo_url        TEXT NOT NULL DEFAULT '',
			default_branch  TEXT NOT NULL DEFAULT '',
			branches        JSONB NOT NULL DEFAULT '[]',
			components      JSONB NOT NULL DEFAULT '[]',
			created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
			last_touched_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);

		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT ''
		);
	`)
	return err
}
