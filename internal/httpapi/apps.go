package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"wip/internal/app"
)

// handleListApps lists active or archived apps and creates new app records.
func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		var entries []app.Entry; var err error
		if r.URL.Query().Get("archived") == "true" { entries, err = s.store.ListArchivedApps() } else { entries, err = s.store.ListApps() }
		if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
		result := make([]EntryWithConnections, len(entries)); for i, entry := range entries { result[i] = s.withConnections(entry) }; writeJSON(w, result)
	case http.MethodPost: s.handleCreateApp(w, r)
	default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateApp validates onboarding data, stores the app, and refreshes Git metadata.
func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) {
	var body struct { Name string `json:"name"`; Description string `json:"description"`; LocalPath string `json:"localPath"`; CreateNew bool `json:"createNew"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, "invalid request body", http.StatusBadRequest); return }
	if body.Name == "" || body.LocalPath == "" { http.Error(w, "name and localPath are required", http.StatusBadRequest); return }
	if body.CreateNew { if err := os.MkdirAll(body.LocalPath, 0o755); err != nil { http.Error(w, fmt.Sprintf("creating folder: %v", err), http.StatusInternalServerError); return } } else if _, err := os.Stat(body.LocalPath); err != nil { http.Error(w, fmt.Sprintf("folder does not exist: %v", err), http.StatusBadRequest); return }
	id, err := s.store.EnsureUniqueID(app.Slugify(body.Name)); if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	entry := app.Entry{ID: id, Name: body.Name, Description: body.Description, Status: app.StatusActive, LocalPath: body.LocalPath}; if err := s.store.CreateApp(entry); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
	s.recordActivity(entry, "app.created", "Added app to WIP", entry.DefaultBranch, "", "stopped", "Created app record", "success", "")
	created, err := s.store.GetApp(id); if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }; if err := s.store.RefreshGitInfo(created.ID, created.LocalPath); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }; created, err = s.store.GetApp(id); if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }; writeJSON(w, s.withConnections(created))
}

// handleAppSubroutes dispatches requests for one app and its component actions.
func (s *Server) handleAppSubroutes(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/api/apps/"):]; id, action := path, ""; for i, character := range path { if character == '/' { id, action = path[:i], path[i+1:]; break } }
	switch action {
	case "":
		switch r.Method { case http.MethodGet: entry, err := s.store.GetApp(id); if err != nil { http.Error(w, err.Error(), http.StatusNotFound); return }; writeJSON(w, s.withConnections(entry)); case http.MethodPut: s.handleUpdateApp(w, r, id); default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed) }
	case "git": entry, err := s.store.GetApp(id); if err != nil { http.Error(w, err.Error(), http.StatusNotFound); return }; writeJSON(w, entry.Branches)
	case "git-refresh": s.handleGitRefresh(w, r, id)
	case "start", "stop", "terminal": s.handleComponentAction(w, r, id, action)
	case "archive": s.handleArchiveApp(w, r, id)
	case "unarchive": entry, err := s.store.GetApp(id); if err != nil { http.Error(w, err.Error(), http.StatusNotFound); return }; if err := s.store.UnarchiveApp(id); err != nil { s.recordActivity(entry, "app.unarchived", "Restored app from archive", entry.DefaultBranch, "", "", "Unarchive failed", "failure", err.Error()); http.Error(w, err.Error(), http.StatusInternalServerError); return }; s.recordActivity(entry, "app.unarchived", "Restored app from archive", entry.DefaultBranch, "", "stopped", "App restored to registry", "success", ""); writeJSON(w, map[string]string{"status": "ok"})
	case "components": s.handleUpdateComponents(w, r, id)
	default: http.NotFound(w, r)
	}
}

// handleUpdateApp saves editable app metadata and lifecycle status.
func (s *Server) handleUpdateApp(w http.ResponseWriter, r *http.Request, id string) {
	var body struct { Name string `json:"name"`; Description string `json:"description"`; LocalPath string `json:"localPath"`; Status string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, "invalid request body", http.StatusBadRequest); return }; if body.Name == "" || body.LocalPath == "" { http.Error(w, "name and localPath are required", http.StatusBadRequest); return }
	status := app.Status(body.Status); switch status { case app.StatusActive, app.StatusPaused, app.StatusAbandoned, app.StatusShipped: default: http.Error(w, "invalid status", http.StatusBadRequest); return }
	if err := s.store.UpdateApp(id, body.Name, body.Description, body.LocalPath, status); err != nil { if entry, getErr := s.store.GetApp(id); getErr == nil { s.recordActivity(entry, "app.updated", "Updated app details", entry.DefaultBranch, "", "", "App update failed", "failure", err.Error()) }; http.Error(w, err.Error(), http.StatusInternalServerError); return }
	if err := s.store.RefreshGitInfo(id, body.LocalPath); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }; entry, err := s.store.GetApp(id); if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }; s.recordActivity(entry, "app.updated", "Updated app details", entry.DefaultBranch, "", "stopped", "Updated name, description, path, or lifecycle status", "success", ""); writeJSON(w, s.withConnections(entry))
}

// handleArchiveApp archives an app and optionally removes its local folder.
func (s *Server) handleArchiveApp(w http.ResponseWriter, r *http.Request, id string) {
	var body struct { DeleteFolder bool `json:"deleteFolder"` }; _ = json.NewDecoder(r.Body).Decode(&body); entry, err := s.store.GetApp(id); if err != nil { http.Error(w, err.Error(), http.StatusNotFound); return }
	if err := s.store.ArchiveApp(id); err != nil { s.recordActivity(entry, "app.archived", "Archived app", entry.DefaultBranch, "", "stopped", "Archive failed", "failure", err.Error()); http.Error(w, err.Error(), http.StatusInternalServerError); return }; if body.DeleteFolder && entry.LocalPath != "" { if err := os.RemoveAll(entry.LocalPath); err != nil { http.Error(w, fmt.Sprintf("archived, but folder deletion failed: %v", err), http.StatusInternalServerError); return } }; s.recordActivity(entry, "app.archived", "Archived app", entry.DefaultBranch, "", "stopped", "App archived", "success", ""); writeJSON(w, map[string]string{"status": "ok"})
}