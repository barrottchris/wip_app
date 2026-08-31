package app

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

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
