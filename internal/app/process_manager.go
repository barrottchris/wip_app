package app

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// InferBrowseURL derives a likely local browse URL for a component from the
// start command. This is intentionally heuristic: many dev servers listen on a
// known port and a human-friendly URL is easier to act on than reading the raw
// command text. The result is used only as a UI hint and is never treated as a
// hard guarantee that the process is healthy.
func InferBrowseURL(component Component) string {
	cmd := strings.TrimSpace(component.StartCommand)
	if cmd == "" {
		return ""
	}

	patterns := []string{
		"--port ",
		"-p ",
		"-l ",
		"--listen ",
		"--host ",
		"--bind ",
	}

	for _, token := range patterns {
		if idx := strings.Index(cmd, token); idx >= 0 {
			value := strings.TrimSpace(cmd[idx+len(token):])
			if value != "" {
				if portMatch := extractPort(value); portMatch != "" {
					return "http://localhost:" + portMatch
				}
			}
		}
	}

	if idx := strings.Index(cmd, "http.server"); idx >= 0 {
		if portMatch := extractPort(cmd[idx:]); portMatch != "" {
			return "http://127.0.0.1:" + portMatch
		}
	}

	if idx := strings.Index(cmd, "serve"); idx >= 0 {
		if portMatch := extractPort(cmd[idx:]); portMatch != "" {
			return "http://localhost:" + portMatch
		}
	}

	if portMatch := extractPort(cmd); portMatch != "" {
		return "http://localhost:" + portMatch
	}

	return ""
}

func extractPort(value string) string {
	parts := strings.Fields(value)
	for i, part := range parts {
		if strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://") {
			if i+1 < len(parts) {
				if port, ok := parsePort(parts[i+1]); ok {
					return port
				}
			}
		}
		if port, ok := parsePort(part); ok {
			return port
		}
		if strings.Contains(part, ":") {
			if port, ok := parsePort(strings.TrimSpace(strings.Split(part, ":")[1])); ok {
				return port
			}
		}
	}
	return ""
}

func parsePort(value string) (string, bool) {
	trimmed := strings.Trim(strings.TrimSpace(value), "\"'")
	trimmed = strings.TrimSuffix(trimmed, ",")
	trimmed = strings.TrimSuffix(trimmed, ";")
	trimmed = strings.TrimSuffix(trimmed, ")")
	trimmed = strings.TrimSuffix(trimmed, "]")
	if trimmed == "" {
		return "", false
	}
	if _, err := strconv.Atoi(trimmed); err == nil {
		return trimmed, true
	}
	return "", false
}

// ProcessManager tracks live OS processes spawned for components.
// This is separate from RuntimeTracker: the tracker reflects UI-visible
// running state, while the process manager owns the actual OS process handles.
type ProcessManager struct {
	mu     sync.Mutex
	procs  map[string]map[string]*exec.Cmd
	stopped map[string]map[string]bool
}

// NewProcessManager creates an in-memory manager for live app component processes.
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		procs:   make(map[string]map[string]*exec.Cmd),
		stopped: make(map[string]map[string]bool),
	}
}

// IsRunning reports whether the supplied component for the given app is still tracked as active.
func (pm *ProcessManager) IsRunning(appID, component string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.procs[appID] == nil {
		return false
	}
	cmd := pm.procs[appID][component]
	if cmd == nil || cmd.Process == nil {
		return false
	}
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		return false
	}
	return true
}

// Start launches a component in the app's local directory and records the running process handle.
func (pm *ProcessManager) Start(appID, appPath string, component Component) error {
	if appID == "" || component.Name == "" {
		return fmt.Errorf("app id and component name are required")
	}
	if component.StartCommand == "" {
		return fmt.Errorf("component %q has no start command", component.Name)
	}

	cmd := buildCommand(component.RunMode, component.StartCommand)
	cmd.Dir = appPath

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %q: %w", component.Name, err)
	}

	pm.mu.Lock()
	if pm.procs[appID] == nil {
		pm.procs[appID] = make(map[string]*exec.Cmd)
	}
	if pm.stopped[appID] == nil {
		pm.stopped[appID] = make(map[string]bool)
	}
	pm.procs[appID][component.Name] = cmd
	pm.stopped[appID][component.Name] = false
	pm.mu.Unlock()
	return nil
}

// Stop shuts down a component using its explicit stop command when available, then falls back to terminating the tracked process tree.
func (pm *ProcessManager) Stop(appID, appPath string, component Component) error {
	pm.mu.Lock()
	cmd := pm.procs[appID][component.Name]
	pm.mu.Unlock()

	if component.StopCommand != "" {
		stopCmd := buildCommand(component.RunMode, component.StopCommand)
		stopCmd.Dir = appPath
		if err := stopCmd.Run(); err != nil {
			// A configured stop command is authoritative for app-specific shutdown.
			// If it fails, we still try to clean up the tracked process tree below.
			// This preserves the app's own shutdown semantics without leaving the
			// backend in a stale running state.
			fmt.Printf("stop command for %q failed: %v\n", component.Name, err)
		}
	}

	if cmd == nil || cmd.Process == nil {
		pm.mu.Lock()
		if pm.procs[appID] != nil {
			delete(pm.procs[appID], component.Name)
		}
		pm.mu.Unlock()
		return nil
	}

	if err := terminateProcessTree(cmd); err != nil {
		if strings.Contains(err.Error(), "exit status 128") || strings.Contains(err.Error(), "not found") {
			pm.mu.Lock()
			if pm.procs[appID] != nil {
				delete(pm.procs[appID], component.Name)
			}
			pm.mu.Unlock()
			return nil
		}
		return err
	}

	pm.mu.Lock()
	if pm.procs[appID] != nil {
		delete(pm.procs[appID], component.Name)
	}
	if pm.stopped[appID] != nil {
		pm.stopped[appID][component.Name] = true
	}
	pm.mu.Unlock()
	return nil
}

// buildCommand constructs the OS command used to launch a component based on its execution mode.
func buildCommand(runMode RunMode, command string) *exec.Cmd {
	if runMode == RunModeDocker {
		return exec.Command("cmd", "/C", command)
	}
	if strings.Contains(command, "&&") || strings.Contains(command, "||") || strings.Contains(command, ";") {
		return exec.Command("cmd", "/C", command)
	}
	// Keep the simplest native case working cross-platform and shell-aware for
	// multi-word commands.
	return exec.Command("cmd", "/C", command)
}

// terminateProcessTree kills the tracked process and its children on Windows using taskkill.
func terminateProcessTree(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}

	pid := strconv.Itoa(cmd.Process.Pid)
	kill := exec.Command("taskkill", "/PID", pid, "/T", "/F")
	if err := kill.Run(); err != nil {
		return fmt.Errorf("taskkill for pid %s: %w", pid, err)
	}
	return nil
}
