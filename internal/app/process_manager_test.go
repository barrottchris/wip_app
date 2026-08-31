package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProcessManagerStartsAndStopsNativeProcess(t *testing.T) {
	pm := NewProcessManager()
	tempDir := t.TempDir()
	component := Component{
		Name:         "Server",
		StartCommand: "ping -n 20 127.0.0.1 >NUL",
		RunMode:      RunModeNative,
	}

	if err := pm.Start("app-1", tempDir, component); err != nil {
		t.Fatalf("start process: %v", err)
	}
	if !pm.IsRunning("app-1", "Server") {
		t.Fatal("expected process to be marked as running")
	}

	if err := pm.Stop("app-1", tempDir, component); err != nil {
		t.Fatalf("stop process: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
	if pm.IsRunning("app-1", "Server") {
		t.Fatal("expected process to be marked as stopped")
	}
}

func TestProcessManagerUsesConfiguredStopCommand(t *testing.T) {
	pm := NewProcessManager()
	tempDir := t.TempDir()
	markerPath := filepath.Join(tempDir, "stop-marker.txt")
	component := Component{
		Name:         "Server",
		StartCommand: "ping -n 1 127.0.0.1 >NUL",
		StopCommand:  "echo stopped > " + markerPath,
		RunMode:      RunModeNative,
	}

	if err := pm.Start("app-2", tempDir, component); err != nil {
		t.Fatalf("start process: %v", err)
	}
	if err := pm.Stop("app-2", tempDir, component); err != nil {
		t.Fatalf("stop process with stop command: %v", err)
	}

	if _, err := os.Stat(markerPath); err != nil {
		t.Fatalf("expected stop command to run and create marker: %v", err)
	}
}
