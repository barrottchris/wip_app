package httpapi

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"wip/internal/app"
)

// handleActivity returns filtered activity events for the activity view.
func (s *Server) handleActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	offset, _ := strconv.Atoi(query.Get("offset"))
	events, err := s.store.ListActivity(limit, offset, query.Get("appId"), query.Get("eventType"), query.Get("outcome"), query.Get("branch"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, events)
}

// recordActivity persists an operation without allowing audit failure to break the request.
func (s *Server) recordActivity(entry app.Entry, eventType, summary, branch, build, runtimeStatus, changes, outcome, detail string) {
	if err := s.store.RecordActivity(app.ActivityEvent{OccurredAt: time.Now(), AppID: entry.ID, AppName: entry.Name, EventType: eventType, Summary: summary, Branch: branch, Build: build, LifecycleStatus: string(entry.Status), RuntimeStatus: runtimeStatus, Changes: changes, Outcome: outcome, Detail: detail}); err != nil {
		log.Printf("recording activity: %v", err)
	}
}
