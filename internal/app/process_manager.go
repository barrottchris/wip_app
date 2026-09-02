package app

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// ProcessManager tracks live OS processes spawned for components.
// This is separate from RuntimeTracker: the tracker reflects UI-visible
// running state, while the process manager owns the actual OS process handles.
type ProcessManager struct {
	mu            sync.Mutex
	procs         map[string]map[string]*ProcessSession
	stopped       map[string]map[string]bool
	onStateChange func(appID, component string, running bool)
}

// NewProcessManager creates an in-memory manager for live app component processes.
// It accepts an optional callback so a caller can bind runtime state changes to
// the portal's own tracker when a process starts or exits.
func NewProcessManager(onStateChange ...func(appID, component string, running bool)) *ProcessManager {
	var hook func(appID, component string, running bool)
	if len(onStateChange) > 0 {
		hook = onStateChange[0]
	}
	return NewProcessManagerWithHook(hook)
}

// NewProcessManagerWithHook adds a callback that fires when a component's live
// process state changes, so the portal can keep its running status in sync with
// the actual OS process lifecycle.
func NewProcessManagerWithHook(onStateChange func(appID, component string, running bool)) *ProcessManager {
	return &ProcessManager{
		procs:         make(map[string]map[string]*ProcessSession),
		stopped:       make(map[string]map[string]bool),
		onStateChange: onStateChange,
	}
}

// IsRunning reports whether the supplied component for the given app is still tracked as active.
func (pm *ProcessManager) IsRunning(appID, component string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.procs[appID] == nil {
		return false
	}
	session := pm.procs[appID][component]
	if session == nil || session.cmd == nil || session.cmd.Process == nil {
		return false
	}
	if session.hasExited() {
		return false
	}
	if session.cmd.ProcessState != nil && session.cmd.ProcessState.Exited() {
		return false
	}
	return true
}

// Start launches a component in the app's local directory and records the live
// process session with its log output so the portal can display a console-like
// stream and react when the process exits.
func (pm *ProcessManager) Start(appID, appPath string, component Component) error {
	if appID == "" || component.Name == "" {
		return fmt.Errorf("app id and component name are required")
	}
	if component.StartCommand == "" {
		return fmt.Errorf("component %q has no start command", component.Name)
	}
	if _, err := os.Stat(appPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("app directory %q does not exist: %w", appPath, err)
		}
		return fmt.Errorf("checking app directory %q: %w", appPath, err)
	}
	if pm.IsRunning(appID, component.Name) {
		return fmt.Errorf("component %q is already running", component.Name)
	}

	cmd := buildTerminalSessionCommand(appPath, component)
	if err := prepareVisibleTerminal(cmd); err != nil {
		return err
	}
	cmd.Dir = appPath
	session := &ProcessSession{cmd: cmd, terminal: true}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("creating terminal input: %w", err)
	}
	session.stdin = stdin
	session.appendLog(fmt.Sprintf("starting %s", component.Name))
	if captureTerminalOutput() {
		cmd.Stdout = &processLogWriter{session: session}
		cmd.Stderr = &processLogWriter{session: session}
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting %q: %w", component.Name, err)
	}

	pm.mu.Lock()
	if pm.procs[appID] == nil {
		pm.procs[appID] = make(map[string]*ProcessSession)
	}
	if pm.stopped[appID] == nil {
		pm.stopped[appID] = make(map[string]bool)
	}
	pm.procs[appID][component.Name] = session
	pm.stopped[appID][component.Name] = false
	pm.mu.Unlock()

	if pm.onStateChange != nil {
		pm.onStateChange(appID, component.Name, true)
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			session.setExitError(err)
			session.appendLog(fmt.Sprintf("process exited: %v", err))
			if len(session.snapshotLogs()) <= 2 {
				session.addDiagnostic("no stdout/stderr was emitted before the process exited")
			}
		} else {
			session.mu.Lock()
			session.exited = true
			session.mu.Unlock()
		}
		if session.stdin != nil {
			_ = session.stdin.Close()
		}
		session.mu.Lock()
		session.exited = true
		session.mu.Unlock()
		if pm.onStateChange != nil {
			pm.onStateChange(appID, component.Name, false)
		}
	}()

	return nil
}

// OpenTerminal reuses the visible command prompt owned by a running session.
// It intentionally does not execute the component start command again.
func (pm *ProcessManager) OpenTerminal(appID, component string) error {
	pm.mu.Lock()
	if pm.procs[appID] == nil {
		pm.mu.Unlock()
		return fmt.Errorf("no terminal session exists for component %q", component)
	}
	session := pm.procs[appID][component]
	pm.mu.Unlock()

	if session == nil || !session.terminal {
		return fmt.Errorf("no attachable terminal session exists for component %q", component)
	}
	if !pm.IsRunning(appID, component) {
		return fmt.Errorf("terminal session for component %q is no longer running", component)
	}
	return nil
}

// SendTerminalInput writes a command or keystrokes to the existing session.
func (pm *ProcessManager) SendTerminalInput(appID, component, input string) error {
	pm.mu.Lock()
	if pm.procs[appID] == nil {
		pm.mu.Unlock()
		return fmt.Errorf("no terminal session exists for component %q", component)
	}
	session := pm.procs[appID][component]
	pm.mu.Unlock()
	if session == nil || session.stdin == nil || !pm.IsRunning(appID, component) {
		return fmt.Errorf("terminal session for component %q is no longer running", component)
	}
	if _, err := session.stdin.Write([]byte(input)); err != nil {
		return fmt.Errorf("writing to terminal for %q: %w", component, err)
	}
	return nil
}

// Stop shuts down a component using its explicit stop command when available, then falls back to terminating the tracked process tree.
func (pm *ProcessManager) Stop(appID, appPath string, component Component) error {
	pm.mu.Lock()
	session := pm.procs[appID][component.Name]
	pm.mu.Unlock()

	if component.StopCommand != "" {
		stopCmd := buildCommand(component.RunMode, component.StopCommand)
		stopCmd.Dir = appPath
		if err := stopCmd.Run(); err != nil {
			fmt.Printf("stop command for %q failed: %v\n", component.Name, err)
		}
	}

	if session == nil || session.cmd == nil || session.cmd.Process == nil {
		pm.mu.Lock()
		if pm.procs[appID] != nil {
			delete(pm.procs[appID], component.Name)
		}
		pm.mu.Unlock()
		if pm.onStateChange != nil {
			pm.onStateChange(appID, component.Name, false)
		}
		return nil
	}

	if err := terminateProcessTree(session.cmd); err != nil {
		if strings.Contains(err.Error(), "exit status 128") || strings.Contains(err.Error(), "not found") {
			pm.mu.Lock()
			if pm.procs[appID] != nil {
				delete(pm.procs[appID], component.Name)
			}
			pm.mu.Unlock()
			if pm.onStateChange != nil {
				pm.onStateChange(appID, component.Name, false)
			}
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
	if pm.onStateChange != nil {
		pm.onStateChange(appID, component.Name, false)
	}
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

// buildTerminalSessionCommand creates the Windows shell command used to keep a
// component's visible terminal session attached to its process.
func buildTerminalSessionCommand(appPath string, component Component) *exec.Cmd {
	command := strings.TrimSpace(component.StartCommand)
	if command == "" {
		command = "echo no start command configured"
	}
	return exec.Command("cmd.exe", "/K", fmt.Sprintf("title %s && %s & exit", component.Name, command))
}

// BuildTerminalCommand creates the Windows command line used to open an
// interactive terminal in the target app directory for a component.
func BuildTerminalCommand(appPath string, component Component) string {
	appPath = strings.TrimSpace(appPath)
	command := strings.TrimSpace(component.StartCommand)
	if appPath == "" {
		return command
	}
	if command == "" {
		return fmt.Sprintf("cd /d \"%s\" && title %s", appPath, component.Name)
	}
	return fmt.Sprintf("cd /d \"%s\" && title %s && %s", appPath, component.Name, command)
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

// GetComponentLogs returns the captured console output for a component, if any.
func (pm *ProcessManager) GetComponentLogs(appID, component string) []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.procs[appID] == nil {
		return nil
	}
	session := pm.procs[appID][component]
	if session == nil {
		return nil
	}
	return session.snapshotLogs()
}

// GetComponentURL returns a URL discovered in the component's captured output.
func (pm *ProcessManager) GetComponentURL(appID, component string) string {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.procs[appID] == nil {
		return ""
	}
	session := pm.procs[appID][component]
	if session == nil {
		return ""
	}
	if url := InferBrowseURLFromLogs(session.snapshotLogs()); url != "" {
		return url
	}
	return ""
}

// GetComponentLastError returns the latest recorded process error for a
// component, if its session is still available.
func (pm *ProcessManager) GetComponentLastError(appID, component string) string {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.procs[appID] == nil {
		return ""
	}
	session := pm.procs[appID][component]
	if session == nil {
		return ""
	}
	return session.getLastError()
}
