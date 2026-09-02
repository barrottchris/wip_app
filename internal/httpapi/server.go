package httpapi

import (
	"net/http"
	"wip/internal/app"
	"wip/internal/config"
)

type Server struct {
	store          *app.Store
	configStore    *config.Store
	runtime        *app.RuntimeTracker
	processManager *app.ProcessManager
}

// NewServer constructs the HTTP API with its storage and runtime dependencies.
func NewServer(store *app.Store, configStore *config.Store, runtime *app.RuntimeTracker, processManager *app.ProcessManager) *Server {
	return &Server{store: store, configStore: configStore, runtime: runtime, processManager: processManager}
}

// Handler builds the API router used by the application's top-level server.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/apps", s.handleListApps)
	mux.HandleFunc("/api/apps/", s.handleAppSubroutes)
	mux.HandleFunc("/api/activity", s.handleActivity)
	mux.HandleFunc("/api/browse", s.handleBrowse)
	mux.HandleFunc("/api/open-folder", s.handleOpenFolder)
	mux.HandleFunc("/api/git-status", s.handleGitStatus)
	mux.HandleFunc("/api/git-init", s.handleGitInit)
	mux.HandleFunc("/api/settings", s.handleSettings)
	return mux
}
