package app_test

import (
	"path/filepath"
	"testing"
	"time"

	"wip/internal/app"
	"wip/internal/config"
	"wip/internal/db"
)

func newTestDB(t *testing.T) *db.DB {
	t.Helper()

	d, err := db.StartAt(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("starting test database: %v", err)
	}

	t.Cleanup(func() {
		if err := d.Stop(); err != nil {
			t.Logf("stopping test database: %v", err)
		}
	})

	return d
}

func TestConfigStorePersistsSettings(t *testing.T) {
	d := newTestDB(t)
	store := config.NewStore(d.Conn)

	if err := store.UpdateManagedRoot(`C:\Temp\WIP`); err != nil {
		t.Fatalf("UpdateManagedRoot() failed: %v", err)
	}

	settings, err := store.Get()
	if err != nil {
		t.Fatalf("Get() after managed root update failed: %v", err)
	}
	if settings.ManagedRoot != `C:\Temp\WIP` {
		t.Fatalf("ManagedRoot = %q; want %q", settings.ManagedRoot, `C:\Temp\WIP`)
	}
	if settings.GitHubUsername != "" {
		t.Fatalf("GitHubUsername = %q; want empty", settings.GitHubUsername)
	}
	if settings.GitHubTokenIsSet {
		t.Fatal("GitHubTokenIsSet = true; want false before token set")
	}

	if err := store.SetGitHubToken("octocat", true); err != nil {
		t.Fatalf("SetGitHubToken(username, true) failed: %v", err)
	}

	settings, err = store.Get()
	if err != nil {
		t.Fatalf("Get() after token set failed: %v", err)
	}
	if settings.GitHubUsername != "octocat" {
		t.Fatalf("GitHubUsername = %q; want %q", settings.GitHubUsername, "octocat")
	}
	if !settings.GitHubTokenIsSet {
		t.Fatal("GitHubTokenIsSet = false; want true")
	}

	if err := store.SetGitHubToken("octocat", false); err != nil {
		t.Fatalf("SetGitHubToken(username, false) failed: %v", err)
	}

	settings, err = store.Get()
	if err != nil {
		t.Fatalf("Get() after clearing token failed: %v", err)
	}
	if settings.GitHubTokenIsSet {
		t.Fatal("GitHubTokenIsSet = true; want false after clearing token")
	}
}

func TestStoreCreateAndListAppRoundTrip(t *testing.T) {
	d := newTestDB(t)
	store := app.NewStore(d.Conn)

	now := time.Now()
	entry := app.Entry{
		ID:            "demo-app",
		Name:          "Demo App",
		Description:   "Example app",
		Stack:         []string{"Go"},
		Status:        app.StatusActive,
		Notes:         "notes",
		LocalPath:     `C:\Work\demo-app`,
		RepoURL:       "https://github.com/example/demo-app",
		DefaultBranch: "main",
		Branches: []app.Branch{
			{Name: "main", LastCommitAt: now.Add(-2 * time.Hour), IsDefault: true},
		},
		Components: []app.Component{
			{Name: "Frontend", StartCommand: "go run .", RunMode: app.RunModeNative},
		},
	}

	if err := store.CreateApp(entry); err != nil {
		t.Fatalf("CreateApp() failed: %v", err)
	}

	got, err := store.GetApp("demo-app")
	if err != nil {
		t.Fatalf("GetApp() failed: %v", err)
	}
	if got.Name != entry.Name {
		t.Fatalf("GetApp().Name = %q; want %q", got.Name, entry.Name)
	}
	if got.Description != entry.Description {
		t.Fatalf("GetApp().Description = %q; want %q", got.Description, entry.Description)
	}
	if len(got.Branches) != 1 || got.Branches[0].Name != "main" {
		t.Fatalf("GetApp().Branches = %+v; want single main branch", got.Branches)
	}
	if len(got.Components) != 1 || got.Components[0].Name != "Frontend" {
		t.Fatalf("GetApp().Components = %+v; want one Frontend component", got.Components)
	}

	apps, err := store.ListApps()
	if err != nil {
		t.Fatalf("ListApps() failed: %v", err)
	}
	if len(apps) != 1 {
		t.Fatalf("len(ListApps()) = %d; want 1", len(apps))
	}

	if err := store.UpdateComponents("demo-app", []app.Component{{Name: "Backend", StartCommand: "npm start", RunMode: app.RunModeNative}}); err != nil {
		t.Fatalf("UpdateComponents() failed: %v", err)
	}

	updated, err := store.GetApp("demo-app")
	if err != nil {
		t.Fatalf("GetApp() after UpdateComponents() failed: %v", err)
	}
	if len(updated.Components) != 1 || updated.Components[0].Name != "Backend" {
		t.Fatalf("updated.Components = %+v; want one Backend component", updated.Components)
	}

	if err := store.ArchiveApp("demo-app"); err != nil {
		t.Fatalf("ArchiveApp() failed: %v", err)
	}

	apps, err = store.ListApps()
	if err != nil {
		t.Fatalf("ListApps() after ArchiveApp() failed: %v", err)
	}
	if len(apps) != 0 {
		t.Fatalf("len(ListApps()) after archive = %d; want 0", len(apps))
	}

	archived, err := store.ListArchivedApps()
	if err != nil {
		t.Fatalf("ListArchivedApps() failed: %v", err)
	}
	if len(archived) != 1 {
		t.Fatalf("len(ListArchivedApps()) = %d; want 1", len(archived))
	}
}
