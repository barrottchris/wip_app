package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
	"wip/internal/app"
	"wip/internal/gitutil"
)

type EntryWithConnections struct { app.Entry; GitConnected bool `json:"gitConnected"`; GitDetails *GitDetails `json:"gitDetails,omitempty"`; JiraConnected bool `json:"jiraConnected"`; JiraComingSoon bool `json:"jiraComingSoon"`; ConfluenceConnected bool `json:"confluenceConnected"`; ConfluenceComingSoon bool `json:"confluenceComingSoon"` }
type GitDetails struct { RepositoryName string `json:"repositoryName"`; Branch string `json:"branch"`; LastUpdate *time.Time `json:"lastUpdate,omitempty"`; RepoURL string `json:"repoUrl,omitempty"` }

// withConnections adds live runtime, log, URL, and Git information to an app.
func (s *Server) withConnections(entry app.Entry) EntryWithConnections {
	s.runtime.ApplyTo(&entry)
	for i := range entry.Components {
		if entry.Components[i].URL == "" { entry.Components[i].URL = app.InferBrowseURL(entry.Components[i]) }
		entry.Components[i].Logs = s.processManager.GetComponentLogs(entry.ID, entry.Components[i].Name)
		if entry.Components[i].URL == "" { entry.Components[i].URL = s.processManager.GetComponentURL(entry.ID, entry.Components[i].Name) }
		if !entry.Components[i].Running && entry.Components[i].URL == "" { entry.Components[i].URL = app.InferBrowseURL(entry.Components[i]) }
	}
	gitConnected := gitutil.HasGit(entry.LocalPath)
	var gitDetails *GitDetails
	if gitConnected {
		gitDetails = &GitDetails{}
		if branch, err := gitutil.DefaultBranch(entry.LocalPath); err == nil && branch != "" { entry.DefaultBranch, gitDetails.Branch = branch, branch }
		entry.RepoURL = ""
		if repoURL, err := gitutil.RemoteURL(entry.LocalPath); err == nil { entry.RepoURL, gitDetails.RepoURL = repoURL, repoURL }
		gitDetails.RepositoryName = gitutil.RepositoryName(entry.RepoURL, entry.LocalPath)
		if lastUpdate, err := gitutil.LastCommitAt(entry.LocalPath); err == nil { gitDetails.LastUpdate = &lastUpdate }
	}
	return EntryWithConnections{Entry: entry, GitConnected: gitConnected, GitDetails: gitDetails, JiraComingSoon: true, ConfluenceComingSoon: true}
}

// writeJSON encodes a response using the API's JSON content type.
func writeJSON(w http.ResponseWriter, value interface{}) { w.Header().Set("Content-Type", "application/json"); if err := json.NewEncoder(w).Encode(value); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError) } }