package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"wip/internal/app"
	"wip/internal/config"
	"wip/internal/db"
	"wip/internal/fsbrowse"
	"wip/internal/gitutil"
)

// Server holds shared state for HTTP handlers.
type Server struct {
	store       *app.Store
	configStore *config.Store
	runtime     *app.RuntimeTracker // live running/stopped state, never persisted — see internal/app/runtime.go
}

func main() {
	fmt.Println("Starting embedded Postgres (first run downloads the binary — may take a moment)...")
	database, err := db.Start()
	if err != nil {
		log.Fatalf("failed to start database: %v", err)
	}
	// Ensure the embedded Postgres process is stopped cleanly on exit,
	// including on Ctrl+C — leaving it running would leak a process and
	// hold the port on next start.
	defer database.Stop()
	handleGracefulShutdown(database)

	store := app.NewStore(database.Conn)
	if err := store.SeedIfEmpty(); err != nil {
		log.Fatalf("failed to seed initial data: %v", err)
	}

	server := &Server{
		store:       store,
		configStore: config.NewStore(database.Conn),
		runtime:     app.NewRuntimeTracker(),
	}

	mux := http.NewServeMux()

	// Apps API
	mux.HandleFunc("/api/apps", server.handleListApps)      // GET list, POST create
	mux.HandleFunc("/api/apps/", server.handleAppSubroutes) // /api/apps/{id}, /api/apps/{id}/start, /stop, /git, /connections

	// Onboarding helpers
	mux.HandleFunc("/api/browse", server.handleBrowse)
	mux.HandleFunc("/api/git-status", server.handleGitStatus)
	mux.HandleFunc("/api/git-init", server.handleGitInit)

	// Settings/config API
	mux.HandleFunc("/api/settings", server.handleSettings)

	// Static frontend (index.html, main.js, style.css)
	mux.Handle("/", http.FileServer(http.Dir("./frontend/src")))

	addr := "localhost:34115" // arbitrary local port, change freely
	fmt.Printf("WIP running at http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// handleGracefulShutdown stops the embedded Postgres process on Ctrl+C /
// SIGTERM, not just on normal return from main(), since log.Fatal from the
// HTTP server would otherwise skip the deferred Stop().
func handleGracefulShutdown(database *db.DB) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		_ = database.Stop()
		os.Exit(0)
	}()
}

// EntryWithConnections wraps an app Entry with computed, non-persisted
// connection info for the UI's pill tags. GitConnected is checked live
// against the filesystem each time — not cached — so it can't go stale if
// someone runs `git init` outside WIP. Jira/Confluence are hardcoded false
// for now since those integrations don't exist yet (see mvp-scope.md).
type EntryWithConnections struct {
	app.Entry
	GitConnected         bool `json:"gitConnected"`
	JiraConnected        bool `json:"jiraConnected"`
	JiraComingSoon       bool `json:"jiraComingSoon"`
	ConfluenceConnected  bool `json:"confluenceConnected"`
	ConfluenceComingSoon bool `json:"confluenceComingSoon"`
}

func (s *Server) withConnections(entry app.Entry) EntryWithConnections {
	s.runtime.ApplyTo(&entry)
	return EntryWithConnections{
		Entry:                entry,
		GitConnected:         gitutil.HasGit(entry.LocalPath),
		JiraConnected:        false,
		JiraComingSoon:       true,
		ConfluenceConnected:  false,
		ConfluenceComingSoon: true,
	}
}

func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var entries []app.Entry
		var err error
		if r.URL.Query().Get("archived") == "true" {
			entries, err = s.store.ListArchivedApps()
		} else {
			entries, err = s.store.ListApps()
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result := make([]EntryWithConnections, len(entries))
		for i, e := range entries {
			result[i] = s.withConnections(e)
		}
		writeJSON(w, result)

	case http.MethodPost:
		s.handleCreateApp(w, r)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateApp backs the "Add app" onboarding flow, covering both the
// "existing folder" and "create new" cases (see mvp-scope.md). Git
// initialization, if needed, is a separate explicit step the frontend
// calls via /api/git-init after the user confirms — this endpoint never
// touches git itself.
func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		LocalPath   string `json:"localPath"`
		CreateNew   bool   `json:"createNew"` // true = "create new" flow, false = "existing folder"
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Name == "" || body.LocalPath == "" {
		http.Error(w, "name and localPath are required", http.StatusBadRequest)
		return
	}

	if body.CreateNew {
		if err := os.MkdirAll(body.LocalPath, 0o755); err != nil {
			http.Error(w, fmt.Sprintf("creating folder: %v", err), http.StatusInternalServerError)
			return
		}
	} else if _, err := os.Stat(body.LocalPath); err != nil {
		http.Error(w, fmt.Sprintf("folder does not exist: %v", err), http.StatusBadRequest)
		return
	}

	baseID := app.Slugify(body.Name)
	id, err := s.store.EnsureUniqueID(baseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	entry := app.Entry{
		ID:          id,
		Name:        body.Name,
		Description: body.Description,
		Status:      app.StatusActive,
		LocalPath:   body.LocalPath,
	}
	if err := s.store.CreateApp(entry); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	created, err := s.store.GetApp(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, created)
}

// handleBrowse powers the folder picker — lists subdirectories of a given
// path (or the user's home directory if none given) so the frontend can
// render a clickable tree without needing native OS file-picker access.
func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	listing, err := fsbrowse.List(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, listing)
}

// handleGitStatus reports whether a given path is already a git repo —
// used by onboarding to decide whether to show the "initialize git?" prompt.
func (s *Server) handleGitStatus(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]bool{"hasGit": gitutil.HasGit(path)})
}

// handleGitInit runs `git init` on a path. Only ever called after the user
// explicitly confirms in the UI — WIP never initializes git silently.
func (s *Server) handleGitInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	if err := gitutil.Init(body.Path); err != nil {
		http.Error(w, fmt.Sprintf("git init failed: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleAppSubroutes is a very simple manual router for MVP purposes.
// TODO: swap for a real router (chi, gorilla/mux, or Go 1.22's enhanced
// http.ServeMux patterns) once routes grow beyond a handful.
func (s *Server) handleAppSubroutes(w http.ResponseWriter, r *http.Request) {
	// Expected paths: /api/apps/{id}, /api/apps/{id}/start, /api/apps/{id}/stop, /api/apps/{id}/git
	path := r.URL.Path[len("/api/apps/"):]

	var id, action string
	for i, c := range path {
		if c == '/' {
			id = path[:i]
			action = path[i+1:]
			break
		}
	}
	if id == "" {
		id = path
	}

	switch action {
	case "":
		switch r.Method {
		case http.MethodGet:
			entry, err := s.store.GetApp(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			writeJSON(w, s.withConnections(entry))

		case http.MethodPut:
			s.handleUpdateApp(w, r, id)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}

	case "git":
		entry, err := s.store.GetApp(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, entry.Branches)

	case "start", "stop":
		var body struct {
			Component string `json:"component"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		// TODO: real process/docker execution is still a stub — but the
		// running/stopped state itself is now tracked for real, so the UI
		// accurately reflects what you last clicked rather than always
		// showing the same status.
		s.runtime.SetRunning(id, body.Component, action == "start")
		fmt.Printf("TODO: %s component %q for app %q\n", action, body.Component, id)
		writeJSON(w, map[string]string{"status": "ok"})

	case "archive":
		s.handleArchiveApp(w, r, id)

	case "unarchive":
		if err := s.store.UnarchiveApp(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"status": "ok"})

	default:
		http.NotFound(w, r)
	}
}

// handleUpdateApp backs the app edit form: name, description, folder path,
// and lifecycle status. Deliberately narrow — components, notes, and stack
// have their own edit flows, not yet built (see README "Still not built").
func (s *Server) handleUpdateApp(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		LocalPath   string `json:"localPath"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if body.Name == "" || body.LocalPath == "" {
		http.Error(w, "name and localPath are required", http.StatusBadRequest)
		return
	}
	status := app.Status(body.Status)
	switch status {
	case app.StatusActive, app.StatusPaused, app.StatusAbandoned, app.StatusShipped:
		// valid
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	if err := s.store.UpdateApp(id, body.Name, body.Description, body.LocalPath, status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	entry, err := s.store.GetApp(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, s.withConnections(entry))
}

// handleArchiveApp soft-archives an app. Deleting the folder on disk is a
// separate, explicit opt-in (deleteFolder: true in the request body) — the
// frontend prompts for this as its own confirmation, distinct from the
// archive confirmation itself, since it's the one genuinely irreversible
// action in the app so far.
func (s *Server) handleArchiveApp(w http.ResponseWriter, r *http.Request, id string) {
	var body struct {
		DeleteFolder bool `json:"deleteFolder"`
	}
	// Body is optional — archiving alone doesn't require one. Ignore a
	// decode error here rather than failing the whole request over it.
	_ = json.NewDecoder(r.Body).Decode(&body)

	entry, err := s.store.GetApp(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := s.store.ArchiveApp(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if body.DeleteFolder && entry.LocalPath != "" {
		if err := os.RemoveAll(entry.LocalPath); err != nil {
			// The app is already archived at this point — report the
			// folder-deletion failure but don't pretend archiving failed too.
			http.Error(w, fmt.Sprintf("archived, but folder deletion failed: %v", err), http.StatusInternalServerError)
			return
		}
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

// handleSettings handles GET (read current settings) and POST (update them)
// for the Settings/Config page — managed folder path, GitHub connection.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		settings, err := s.configStore.Get()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, settings)

	case http.MethodPost:
		var body struct {
			ManagedRoot    string `json:"managedRoot"`
			GitHubUsername string `json:"githubUsername"`
			GitHubToken    string `json:"githubToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if body.ManagedRoot != "" {
			if err := s.configStore.UpdateManagedRoot(body.ManagedRoot); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if body.GitHubUsername != "" || body.GitHubToken != "" {
			if err := s.configStore.SetGitHubToken(body.GitHubUsername, body.GitHubToken != ""); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		settings, err := s.configStore.Get()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, settings)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
