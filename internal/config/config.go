package config

import "database/sql"

// Settings holds user-configurable, app-wide settings — the kind of thing
// shown on a Settings/Config page rather than per-app metadata.
type Settings struct {
	ManagedRoot      string `json:"managedRoot"`
	GitHubUsername   string `json:"githubUsername"`
	GitHubTokenIsSet bool   `json:"githubTokenIsSet"`
	DockerAvailable  bool   `json:"dockerAvailable"`
}

// Store persists settings as simple key/value rows in Postgres.
// TODO: GitHub token itself should move to a real secret store (OS
// credential manager or encrypted file) rather than the settings table —
// only whether one is set gets persisted here for now, never the token.
type Store struct {
	conn *sql.DB
}

func NewStore(conn *sql.DB) *Store {
	return &Store{conn: conn}
}

func (s *Store) Get() (Settings, error) {
	settings := Settings{
		ManagedRoot: `C:\Dev\WIP`, // default if not yet set
	}

	rows, err := s.conn.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return settings, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return settings, err
		}
		switch key {
		case "managed_root":
			settings.ManagedRoot = value
		case "github_username":
			settings.GitHubUsername = value
		case "github_token_is_set":
			settings.GitHubTokenIsSet = value == "true"
		}
	}
	return settings, rows.Err()
}

func (s *Store) UpdateManagedRoot(path string) error {
	return s.upsert("managed_root", path)
}

// SetGitHubToken records that a token was provided without storing the
// token value itself in this table.
func (s *Store) SetGitHubToken(username string, tokenProvided bool) error {
	if err := s.upsert("github_username", username); err != nil {
		return err
	}
	value := "false"
	if tokenProvided {
		value = "true"
	}
	return s.upsert("github_token_is_set", value)
}

func (s *Store) upsert(key, value string) error {
	_, err := s.conn.Exec(`
		INSERT INTO settings (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value
	`, key, value)
	return err
}
