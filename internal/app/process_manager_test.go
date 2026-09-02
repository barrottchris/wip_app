package app

import (
	"os"
	"path/filepath"
	"strings"
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

func TestProcessManagerOpenTerminalRequiresRunningSession(t *testing.T) {
	pm := NewProcessManager()

	if err := pm.OpenTerminal("missing-app", "Server"); err == nil {
		t.Fatal("expected opening a missing terminal session to fail")
	}
}

func TestProcessManagerRejectsDuplicateStart(t *testing.T) {
	pm := NewProcessManager()
	tempDir := t.TempDir()
	component := Component{
		Name:         "Server",
		StartCommand: "ping -n 20 127.0.0.1 >NUL",
		RunMode:      RunModeNative,
	}

	if err := pm.Start("app-duplicate", tempDir, component); err != nil {
		t.Fatalf("start process: %v", err)
	}
	t.Cleanup(func() {
		if err := pm.Stop("app-duplicate", tempDir, component); err != nil {
			t.Logf("stop process: %v", err)
		}
	})

	if err := pm.OpenTerminal("app-duplicate", "Server"); err != nil {
		t.Fatalf("open running terminal: %v", err)
	}
	if err := pm.Start("app-duplicate", tempDir, component); err == nil {
		t.Fatal("expected duplicate start to fail")
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

func TestProcessManagerRejectsMissingAppDirectory(t *testing.T) {
	pm := NewProcessManager()
	component := Component{
		Name:         "App",
		StartCommand: "npm start",
		RunMode:      RunModeNative,
	}

	if err := pm.Start("missing-dir", filepath.Join(t.TempDir(), "ghost-project"), component); err == nil {
		t.Fatal("expected start to fail when app directory does not exist")
	}
}

func TestProcessManagerCapturesEarlyExitOutput(t *testing.T) {
	pm := NewProcessManager()
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "exit_script.py")
	script := "import sys\nprint(\"hello stdout\")\nsys.stderr.write(\"hello stderr\\n\")\nraise SystemExit(3)\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o600); err != nil {
		t.Fatalf("write temp script: %v", err)
	}

	component := Component{
		Name:         "App",
		StartCommand: "python exit_script.py",
		RunMode:      RunModeNative,
	}

	if err := pm.Start("app-logs", tempDir, component); err != nil {
		t.Fatalf("start process: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		logs := pm.GetComponentLogs("app-logs", "App")
		if len(logs) > 0 {
			joined := strings.Join(logs, "\n")
			if strings.Contains(joined, "hello stdout") && strings.Contains(joined, "hello stderr") {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("expected captured stdout/stderr from early-exit process, got logs: %#v", pm.GetComponentLogs("app-logs", "App"))
}

func TestInferBrowseURLFromLogs(t *testing.T) {
	logs := []string{
		"Starting app...",
		"Local:   http://localhost:3000",
		"Network: http://0.0.0.0:3000",
	}

	if got := InferBrowseURLFromLogs(logs); got != "http://localhost:3000" {
		t.Fatalf("InferBrowseURLFromLogs() = %q, want %q", got, "http://localhost:3000")
	}
}

func TestInferBrowseURL(t *testing.T) {
	cases := []struct {
		name     string
		command  string
		expected string
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

func TestBuildTerminalCommandUsesAppDirectoryAndComponentName(t *testing.T) {
	appPath := `C:\work\demo-app`
	cmd := BuildTerminalCommand(appPath, Component{Name: "Frontend", StartCommand: "npm run dev"})

	if !strings.Contains(cmd, `cd /d "C:\work\demo-app"`) {
		t.Fatalf("terminal command should cd into app directory, got %q", cmd)
	}
	if !strings.Contains(cmd, "title Frontend") {
		t.Fatalf("terminal command should set the title to the component name, got %q", cmd)
	}
	if !strings.Contains(cmd, "npm run dev") {
		t.Fatalf("terminal command should include the start command, got %q", cmd)
	}
}
