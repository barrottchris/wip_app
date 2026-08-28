package config

// Settings holds user-configurable, app-wide settings — the kind of thing
// shown on a Settings/Config page rather than per-app metadata.
type Settings struct {
	// ManagedRoot is the single folder on disk where WIP keeps/expects
	// tracked apps to live (the "organized C drive" requirement).
	ManagedRoot string `json:"managedRoot"`

	// GitHub integration — MVP just stores a username and whether a token
	// has been set (never echo the token itself back to the frontend).
	GitHubUsername    string `json:"githubUsername"`
	GitHubTokenIsSet  bool   `json:"githubTokenIsSet"`

	// DockerAvailable reflects whether WIP could detect a local Docker
	// install — informational only for now, not enforced.
	DockerAvailable bool `json:"dockerAvailable"`
}

// Store is a placeholder in-memory settings store.
// TODO: persist to disk (likely alongside wherever app registry data ends
// up being stored — see open storage question in data-model.md).
type Store struct {
	settings Settings
}

func NewStore() *Store {
	return &Store{
		settings: Settings{
			ManagedRoot:      `C:\Dev\WIP`,
			GitHubUsername:   "",
			GitHubTokenIsSet: false,
			DockerAvailable:  false,
		},
	}
}

func (s *Store) Get() Settings {
	return s.settings
}

func (s *Store) UpdateManagedRoot(path string) {
	s.settings.ManagedRoot = path
}

// SetGitHubToken records that a token was provided without storing it in
// this placeholder store — a real implementation needs a proper secrets
// approach (OS credential store, encrypted file, etc.), not a TODO left in
// memory.
func (s *Store) SetGitHubToken(username string, tokenProvided bool) {
	s.settings.GitHubUsername = username
	s.settings.GitHubTokenIsSet = tokenProvided
}
