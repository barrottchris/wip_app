package httpapi

import ("encoding/json"; "net/http")

// handleSettings reads or updates the local application settings.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet: settings, err := s.configStore.Get(); if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }; writeJSON(w, settings)
	case http.MethodPost:
		var body struct { ManagedRoot string `json:"managedRoot"`; GitHubUsername string `json:"githubUsername"`; GitHubToken string `json:"githubToken"` }; if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, "invalid request body", http.StatusBadRequest); return }
		if body.ManagedRoot != "" { if err := s.configStore.UpdateManagedRoot(body.ManagedRoot); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return } }; if body.GitHubUsername != "" || body.GitHubToken != "" { if err := s.configStore.SetGitHubToken(body.GitHubUsername, body.GitHubToken != ""); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return } }; settings, err := s.configStore.Get(); if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }; writeJSON(w, settings)
	default: http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}