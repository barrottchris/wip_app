package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"wip/internal/gitutil"
)

// Store persists App Entries in SQLite (see internal/db). Replaces the
// earlier in-memory placeholder — data now survives a restart.
type Store struct {
	conn *sql.DB
}

// Slugify turns a display name into a URL/ID-safe slug (e.g. "My Cool App"
// -> "my-cool-app"). Used to derive an app's ID from its name during
// onboarding. Not guaranteed unique on its own — callers should check for
// collisions (see EnsureUniqueID).
func Slugify(name string) string {
	lower := strings.ToLower(strings.TrimSpace(name))
	var b strings.Builder
	lastWasDash := false
	for _, r := range lower {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			lastWasDash = false
		default:
			if !lastWasDash {
				b.WriteRune('-')
				lastWasDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// EnsureUniqueID appends -2, -3, etc. to base until it doesn't collide with
// an existing app ID.
func (s *Store) EnsureUniqueID(base string) (string, error) {
	id := base
	for i := 2; ; i++ {
		_, err := s.GetApp(id)
		if err != nil {
			// GetApp returning an error means no app has this ID — free to use.
			return id, nil
		}
		id = fmt.Sprintf("%s-%d", base, i)
	}
}

func NewStore(conn *sql.DB) *Store {
	return &Store{conn: conn}
}

// ListApps returns every non-archived tracked app for the main registry
// view. Archived apps are deliberately excluded — see ListArchivedApps.
func (s *Store) ListApps() ([]Entry, error) {
	return s.listByArchived(false)
}

// ListArchivedApps returns only archived apps, for the dedicated Archived
// view (see mvp-scope.md / next-phase-plan.md — archived apps are hidden
// from the main registry, not just greyed out).
func (s *Store) ListArchivedApps() ([]Entry, error) {
	return s.listByArchived(true)
}

func (s *Store) listByArchived(archived bool) ([]Entry, error) {
	rows, err := s.conn.Query(`
		SELECT id, name, description, stack, status, notes, local_path,
		       repo_url, default_branch, branches, components,
		       created_at, last_touched_at, archived
		FROM apps
		WHERE archived = $1
		ORDER BY last_touched_at DESC
	`, archived)
	if err != nil {
		return nil, fmt.Errorf("listing apps: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// GetApp returns a single app's full detail.
// GetApp returns a single app's full detail, regardless of archived state
// (the detail/edit page needs to work for archived apps too, e.g. to
// unarchive them).
func (s *Store) GetApp(id string) (Entry, error) {
	row := s.conn.QueryRow(`
		SELECT id, name, description, stack, status, notes, local_path,
		       repo_url, default_branch, branches, components,
		       created_at, last_touched_at, archived
		FROM apps WHERE id = $1
	`, id)

	entry, err := scanEntry(row)
	if err == sql.ErrNoRows {
		return Entry{}, fmt.Errorf("app not found: %s", id)
	}
	return entry, err
}

// UpdateApp updates the user-editable fields of an existing app: name,
// description, local path, and lifecycle status. Deliberately does not
// touch branches/components here — those are managed by their own flows
// (git integration, start/stop config) rather than this general edit form.
// Does not update last_touched_at — that reflects real project activity
// (git commits), not metadata edits.
func (s *Store) UpdateApp(id string, name, description, localPath string, status Status) error {
	_, err := s.conn.Exec(`
		UPDATE apps
		SET name = $1, description = $2, local_path = $3, status = $4
		WHERE id = $5
	`, name, description, localPath, status, id)
	return err
}

// ArchiveApp soft-deletes an app — sets archived = true. It never touches
// the filesystem itself; if the folder should also be deleted, that is a
// separate, explicit step the caller (main.go's handler) performs after
// its own confirmation, not something this method does implicitly.
func (s *Store) ArchiveApp(id string) error {
	_, err := s.conn.Exec(`UPDATE apps SET archived = true WHERE id = $1`, id)
	return err
}

// UnarchiveApp reverses ArchiveApp, restoring an app to the main registry.
func (s *Store) UnarchiveApp(id string) error {
	_, err := s.conn.Exec(`UPDATE apps SET archived = false WHERE id = $1`, id)
	return err
}

// UpdateComponents replaces an app's entire component list wholesale —
// the simplest approach (vs. diffing individual rows), chosen since the
// frontend always submits the full current set from its edit form rather
// than individual add/remove operations against the backend.
func (s *Store) UpdateComponents(id string, components []Component) error {
	componentsJSON, err := json.Marshal(components)
	if err != nil {
		return err
	}
	_, err = s.conn.Exec(`UPDATE apps SET components = $1 WHERE id = $2`, componentsJSON, id)
	return err
}

// RefreshGitInfo updates the app's persisted repo metadata from the actual
// repo on disk rather than trusting stale seed/default values.
func (s *Store) RefreshGitInfo(id, path string) error {
	if path == "" || !gitutil.HasGit(path) {
		_, err := s.conn.Exec(`UPDATE apps SET repo_url = $1, default_branch = $2 WHERE id = $3`, "", "", id)
		return err
	}

	repoURL, err := gitutil.RemoteURL(path)
	if err != nil {
		repoURL = ""
	}
	defaultBranch, err := gitutil.DefaultBranch(path)
	if err != nil {
		defaultBranch = ""
	}

	_, err = s.conn.Exec(`UPDATE apps SET repo_url = $1, default_branch = $2 WHERE id = $3`, repoURL, defaultBranch, id)
	return err
}

// RefreshGitInfoForPath refreshes every app that points at a given local path.
func (s *Store) RefreshGitInfoForPath(path string) error {
	rows, err := s.conn.Query(`SELECT id FROM apps WHERE local_path = $1`, path)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return err
		}
		if err := s.RefreshGitInfo(id, path); err != nil {
			return err
		}
	}
	return rows.Err()
}

// CreateApp inserts a new app entry — the eventual backing for the
// "Add app" onboarding flow (not yet wired up to the frontend).
func (s *Store) CreateApp(entry Entry) error {
	stackJSON, err := json.Marshal(entry.Stack)
	if err != nil {
		return err
	}
	branchesJSON, err := json.Marshal(entry.Branches)
	if err != nil {
		return err
	}
	componentsJSON, err := json.Marshal(entry.Components)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = s.conn.Exec(`
		INSERT INTO apps (
			id, name, description, stack, status, notes, local_path,
			repo_url, default_branch, branches, components,
			created_at, last_touched_at, archived
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, false)
	`,
		entry.ID, entry.Name, entry.Description, stackJSON, entry.Status,
		entry.Notes, entry.LocalPath, entry.RepoURL, entry.DefaultBranch,
		branchesJSON, componentsJSON, now, now,
	)
	return err
}

// scanner is satisfied by both *sql.Row and *sql.Rows, letting scanEntry
// serve both GetApp (single row) and ListApps (multiple rows).
type scanner interface {
	Scan(dest ...interface{}) error
}

func scanEntry(row scanner) (Entry, error) {
	var e Entry
	var stackJSON, branchesJSON, componentsJSON []byte

	err := row.Scan(
		&e.ID, &e.Name, &e.Description, &stackJSON, &e.Status, &e.Notes,
		&e.LocalPath, &e.RepoURL, &e.DefaultBranch, &branchesJSON,
		&componentsJSON, &e.CreatedAt, &e.LastTouchedAt, &e.Archived,
	)
	if err != nil {
		return Entry{}, err
	}

	if err := json.Unmarshal(stackJSON, &e.Stack); err != nil {
		return Entry{}, fmt.Errorf("decoding stack: %w", err)
	}
	if err := json.Unmarshal(branchesJSON, &e.Branches); err != nil {
		return Entry{}, fmt.Errorf("decoding branches: %w", err)
	}
	if err := json.Unmarshal(componentsJSON, &e.Components); err != nil {
		return Entry{}, fmt.Errorf("decoding components: %w", err)
	}

	return e, nil
}

// SeedIfEmpty inserts one sample app on first run only, so the UI has
// something to show before real onboarding exists. Safe to call every
// startup — it checks first.
func (s *Store) SeedIfEmpty() error {
	var count int
	if err := s.conn.QueryRow(`SELECT COUNT(*) FROM apps`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	return s.CreateApp(Entry{
		ID:            "bordle",
		Name:          "Bordle",
		Description:   "Geography-based word game",
		Stack:         []string{"Node.js"},
		Status:        StatusActive,
		LocalPath:     `C:\bordle`,
		RepoURL:       "https://github.com/example/bordle",
		DefaultBranch: "main",
		Branches: []Branch{
			{Name: "main", LastCommitAt: now.Add(-48 * time.Hour), IsDefault: true},
		},
		Components: []Component{
			{Name: "App", StartCommand: "python .\\app.py", StopCommand: "", RunMode: RunModeNative},
		},
	})
}
