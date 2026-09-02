package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"wip/internal/fsbrowse"
)

// handleBrowse lists directories for the onboarding folder picker.
func (s *Server) handleBrowse(w http.ResponseWriter, r *http.Request) {
	listing, err := fsbrowse.List(r.URL.Query().Get("path"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, listing)
}

// handleOpenFolder opens a tracked directory in Windows Explorer.
func (s *Server) handleOpenFolder(w http.ResponseWriter, r *http.Request) {
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
	info, err := os.Stat(body.Path)
	if err != nil || !info.IsDir() {
		http.Error(w, "directory is not available", http.StatusBadRequest)
		return
	}
	if runtime.GOOS != "windows" {
		http.Error(w, "opening folders is only supported on Windows", http.StatusNotImplemented)
		return
	}
	if err := exec.Command("explorer.exe", body.Path).Start(); err != nil {
		http.Error(w, fmt.Sprintf("could not open directory: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}
