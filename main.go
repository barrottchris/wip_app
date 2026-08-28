package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"wip/internal/app"
	"wip/internal/config"
)

// Server holds shared state for HTTP handlers.
type Server struct {
	store       *app.Store
	configStore *config.Store
}

func main() {
	server := &Server{
		store:       app.NewStore(),
		configStore: config.NewStore(),
	}

	mux := http.NewServeMux()

	// Apps API
	mux.HandleFunc("/api/apps", server.handleListApps)
	mux.HandleFunc("/api/apps/", server.handleAppSubroutes) // /api/apps/{id}, /api/apps/{id}/start, /stop, /git

	// Settings/config API
	mux.HandleFunc("/api/settings", server.handleSettings)

	// Static frontend (index.html, main.js, style.css)
	mux.Handle("/", http.FileServer(http.Dir("./frontend/src")))

	addr := "localhost:34115" // arbitrary local port, change freely
	fmt.Printf("WIP running at http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.store.ListApps())
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
		entry, err := s.store.GetApp(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, entry)

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
		// TODO: real process/docker execution — stub for now.
		fmt.Printf("TODO: %s component %q for app %q\n", action, body.Component, id)
		writeJSON(w, map[string]string{"status": "ok"})

	default:
		http.NotFound(w, r)
	}
}

// handleSettings handles GET (read current settings) and POST (update them)
// for the Settings/Config page — managed folder path, GitHub connection.
func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, s.configStore.Get())

	case http.MethodPost:
		var body struct {
			ManagedRoot     string `json:"managedRoot"`
			GitHubUsername  string `json:"githubUsername"`
			GitHubToken     string `json:"githubToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if body.ManagedRoot != "" {
			s.configStore.UpdateManagedRoot(body.ManagedRoot)
		}
		if body.GitHubUsername != "" || body.GitHubToken != "" {
			s.configStore.SetGitHubToken(body.GitHubUsername, body.GitHubToken != "")
		}
		writeJSON(w, s.configStore.Get())

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
