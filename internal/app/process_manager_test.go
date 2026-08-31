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

func TestInferBrowseURL(t *testing.T) {
	cases := []struct {
		name      string
		command   string
		expected  string
	}{
		{name: "python http server", command: "python -m http.server 8765 --bind 127.0.0.1", expected: "http://127.0.0.1:8765"},
		{name: "vite port flag", command: "npm run dev -- --host 0.0.0.0 --port 5173", expected: "http://localhost:5173"},
		{name: "explicit url", command: "npx serve -l 3000", expected: "http://localhost:3000"},
		{name: "no port", command: "python app.py", expected: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			component := Component{Name: "Server", StartCommand: tc.command}
			if got := InferBrowseURL(component); got != tc.expected {
				t.Fatalf("InferBrowseURL(%q) = %q, want %q", tc.command, got, tc.expected)
			}
		})
	}
}
