package app

import (
	"fmt"
	"time"
)

// Store is a placeholder persistence layer.
// TODO: replace with real storage once decided (e.g. SQLite vs JSON files —
// see open questions in mvp-scope.md / data-model.md). For now this is just
// an in-memory slice seeded with sample data so the frontend has something
// to render while the UI is being built.
type Store struct {
	entries []Entry
}

func NewStore() *Store {
	return &Store{
		entries: seedData(),
	}
}

func (s *Store) ListApps() []Entry {
	return s.entries
}

func (s *Store) GetApp(id string) (Entry, error) {
	for _, e := range s.entries {
		if e.ID == id {
			return e, nil
		}
	}
	return Entry{}, fmt.Errorf("app not found: %s", id)
}

func seedData() []Entry {
	now := time.Now()
	return []Entry{
		{
			ID:            "bordle",
			Name:          "Bordle",
			Description:   "Geography-based word game",
			Stack:         []string{"Node.js"},
			Status:        StatusActive,
			LocalPath:     `C:\Dev\WIP\bordle`,
			RepoURL:       "https://github.com/example/bordle",
			DefaultBranch: "main",
			Branches: []Branch{
				{Name: "main", LastCommitAt: now.Add(-48 * time.Hour), IsDefault: true},
			},
			Components: []Component{
				{Name: "App", StartCommand: "npm start", StopCommand: "", RunMode: RunModeNative},
			},
			CreatedAt:     now.Add(-90 * 24 * time.Hour),
			LastTouchedAt: now.Add(-48 * time.Hour),
		},
	}
}
