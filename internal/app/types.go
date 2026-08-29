package app

import "time"

// Status is the user-set lifecycle stage of a tracked app — NOT whether it
// is currently running. Whether something is actually running is a
// separate, computed concept (see RuntimeTracker) based on live component
// process state, and is never stored here. A "shipped" app and a "paused"
// app are both, correctly, never "running" from WIP's point of view unless
// something has actually been started.
type Status string

const (
	StatusActive    Status = "active" // actively being worked on
	StatusPaused    Status = "paused"
	StatusAbandoned Status = "abandoned"
	StatusShipped   Status = "shipped"
)

// RunMode indicates how a component is executed.
type RunMode string

const (
	RunModeDocker RunMode = "docker"
	RunModeNative RunMode = "native"
)

// Component is a single runnable unit within an app (e.g. "Frontend", "Backend").
type Component struct {
	Name         string  `json:"name"`
	StartCommand string  `json:"startCommand"`
	StopCommand  string  `json:"stopCommand"`
	RunMode      RunMode `json:"runMode"`
	// Running is computed at runtime, not persisted.
	Running bool `json:"running"`
}

// Branch reflects a git branch discovered for an app's repo.
// MVP: read-only, sourced from git — no branch management actions yet.
type Branch struct {
	Name         string    `json:"name"`
	LastCommitAt time.Time `json:"lastCommitAt"`
	IsDefault    bool      `json:"isDefault"`
}

// Entry is a single tracked app/project — the core object in the registry.
type Entry struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Description   string      `json:"description"`
	Stack         []string    `json:"stack"`
	Status        Status      `json:"status"`
	Notes         string      `json:"notes"`
	LocalPath     string      `json:"localPath"`
	RepoURL       string      `json:"repoUrl"`
	DefaultBranch string      `json:"defaultBranch"`
	Branches      []Branch    `json:"branches"`
	Components    []Component `json:"components"`
	CreatedAt     time.Time   `json:"createdAt"`
	LastTouchedAt time.Time   `json:"lastTouchedAt"`
	Archived      bool        `json:"archived"`
}
