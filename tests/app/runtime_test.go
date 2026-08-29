package app_test

import (
	"testing"

	"wip/internal/app"
)

func TestRuntimeTrackerAppliesRunningState(t *testing.T) {
	tracker := app.NewRuntimeTracker()

	entry := app.Entry{
		ID: "myapp",
		Components: []app.Component{
			{Name: "Frontend"},
			{Name: "Backend"},
		},
	}

	tracker.ApplyTo(&entry)
	if entry.Components[0].Running || entry.Components[1].Running {
		t.Fatalf("expected both components not running initially, got %+v", entry.Components)
	}

	tracker.SetRunning("myapp", "Frontend", true)
	tracker.ApplyTo(&entry)
	if !entry.Components[0].Running {
		t.Errorf("expected Frontend running after SetRunning(true)")
	}
	if entry.Components[1].Running {
		t.Errorf("expected Backend still not running")
	}

	tracker.SetRunning("myapp", "Frontend", false)
	tracker.ApplyTo(&entry)
	if entry.Components[0].Running {
		t.Errorf("expected Frontend not running after SetRunning(false)")
	}
}

func TestRuntimeTrackerIsolatesApps(t *testing.T) {
	tracker := app.NewRuntimeTracker()
	tracker.SetRunning("app-a", "App", true)

	if tracker.IsRunning("app-b", "App") {
		t.Errorf("expected app-b's component to be unaffected by app-a's state")
	}
	if !tracker.IsRunning("app-a", "App") {
		t.Errorf("expected app-a's component to be running")
	}
}
