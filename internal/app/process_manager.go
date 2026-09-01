package app

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// processLogWriter captures command output so the portal can show a live
// terminal-like log, not just a boolean running flag.
type processLogWriter struct {
	session *ProcessSession
}

func (w *processLogWriter) Write(p []byte) (int, error) {
	if w == nil || w.session == nil {
		return len(p), nil
	}
	for _, line := range strings.Split(strings.TrimRight(string(p), "\r\n"), "\n") {
		msg := strings.TrimSpace(line)
		if msg == "" {
			continue
		}
		w.session.appendLog(msg)
	}
	return len(p), nil
}

// ProcessSession holds a real OS process plus its captured output so the portal
// can display terminal-like log output and react to termination.
type ProcessSession struct {
	mu          sync.Mutex
	cmd         *exec.Cmd
	logs        []string
	lastError   string
	lastExitErr error
	exited      bool
}

func (s *ProcessSession) appendLog(msg string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.logs) >= 200 {
		s.logs = append([]string(nil), s.logs[len(s.logs)-199:]...)
	}
	s.logs = append(s.logs, msg)
}

func (s *ProcessSession) snapshotLogs() []string {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.logs))
	copy(out, s.logs)
	return out
}

func (s *ProcessSession) setExitError(err error) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastExitErr = err
	s.exited = true
	if err != nil {
		s.lastError = err.Error()
		if !strings.Contains(strings.Join(s.logs, "\n"), s.lastError) {
			s.logs = append(s.logs, s.lastError)
		}
	}
}

func (s *ProcessSession) addDiagnostic(message string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !strings.Contains(strings.Join(s.logs, "\n"), message) {
		s.logs = append(s.logs, message)
	}
}

func (s *ProcessSession) getLastError() string {
	if s == nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastError
}

func (s *ProcessSession) hasExited() bool {
	if s == nil {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exited
}

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

func InferBrowseURLFromLogs(logs []string) string {
	for _, line := range logs {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if idx := strings.Index(trimmed, "http://"); idx >= 0 {
			if url, ok := readURLToken(trimmed[idx:]); ok {
				return url
			}
		}
		if idx := strings.Index(trimmed, "https://"); idx >= 0 {
			if url, ok := readURLToken(trimmed[idx:]); ok {
				return url
			}
		}
	}
	return ""
}

func readURLToken(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	for _, end := range []string{" ", "\n", "\t", "\"", "'", ")", "]", "}"} {
		if idx := strings.Index(trimmed, end); idx >= 0 {
			trimmed = trimmed[:idx]
		}
	}
	if trimmed == "" {
		return "", false
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed, true
	}
	return "", false
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

	cmd := buildCommand(component.RunMode, component.StartCommand)
	cmd.Dir = appPath
	session := &ProcessSession{cmd: cmd}
	session.appendLog(fmt.Sprintf("starting %s", component.Name))
	cmd.Stdout = &processLogWriter{session: session}
	cmd.Stderr = &processLogWriter{session: session}

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
		session.mu.Lock()
		session.exited = true
		session.mu.Unlock()
		if pm.onStateChange != nil {
			pm.onStateChange(appID, component.Name, false)
		}
	}()

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
