package app

import "sync"

// RuntimeTracker holds whether each app's components are currently running.
// This is deliberately NOT persisted to the database — it reflects live
// process state, which resets to "not running" on every WIP restart (the
// same way real processes wouldn't survive a restart either). This is
// separate from Entry.Status, which is a user-set lifecycle label (active
// development / paused / abandoned / shipped) and does not change based on
// whether something happens to be running right now.
type RuntimeTracker struct {
	mu    sync.Mutex
	state map[string]map[string]bool // appID -> componentName -> running
}

func NewRuntimeTracker() *RuntimeTracker {
	return &RuntimeTracker{state: make(map[string]map[string]bool)}
}

func (t *RuntimeTracker) SetRunning(appID, component string, running bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.state[appID] == nil {
		t.state[appID] = make(map[string]bool)
	}
	t.state[appID][component] = running
}

func (t *RuntimeTracker) IsRunning(appID, component string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.state[appID][component]
}

// ApplyTo overlays known running state onto an Entry's components in place.
// Call this after reading entries from the store, before returning them
// over the API — the database never stores this, so it has to be merged
// in at read time.
func (t *RuntimeTracker) ApplyTo(entry *Entry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	componentStates := t.state[entry.ID]
	for i := range entry.Components {
		entry.Components[i].Running = componentStates[entry.Components[i].Name]
		if entry.Components[i].Running {
			if entry.Components[i].URL == "" {
				entry.Components[i].URL = InferBrowseURL(entry.Components[i])
			}
		}
	}
}
