package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"wip/internal/gitutil"
)

// handleGitStatus reports whether a path contains a Git repository.
func (s *Server) handleGitStatus(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path is required", http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]bool{"hasGit": gitutil.HasGit(path)})
}

// handleGitInit initializes Git after the user explicitly confirms onboarding.
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
	if entry, err := s.store.GetAppByPath(body.Path); err == nil {
		s.recordActivity(entry, "git.init", "Initialized git repository", entry.DefaultBranch, "", "stopped", "Created repository metadata", "success", "")
	}
	if err := s.store.RefreshGitInfoForPath(body.Path); err != nil {
		http.Error(w, fmt.Sprintf("refresh git metadata failed: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

// handleGitRefresh refreshes stored Git metadata and returns the enriched app.
func (s *Server) handleGitRefresh(w http.ResponseWriter, r *http.Request, id string) {
	entry, err := s.store.GetApp(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := s.store.RefreshGitInfo(id, entry.LocalPath); err != nil {
		s.recordActivity(entry, "git.refresh", "Refreshed git metadata", entry.DefaultBranch, "", "", "Git metadata refresh failed", "failure", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.recordActivity(entry, "git.refresh", "Refreshed git metadata", entry.DefaultBranch, "", "", "Updated repository and branch metadata", "success", "")
	updated, err := s.store.GetApp(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, s.withConnections(updated))
}
