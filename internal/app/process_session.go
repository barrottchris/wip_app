package app

import (
	"io"
	"os/exec"
	"strings"
	"sync"
)

// processLogWriter captures command output so the portal can show a live
// terminal-like log, not just a boolean running flag.
type processLogWriter struct {
	session *ProcessSession
}

// Write splits terminal output into trimmed lines and appends non-empty lines
// to the associated process session.
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
	stdin       io.WriteCloser
	terminal    bool
	logs        []string
	lastError   string
	lastExitErr error
	exited      bool
}

// appendLog adds a message while retaining only the most recent 200 entries.
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

// snapshotLogs returns a copy of the session log so callers cannot mutate it.
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

// setExitError records process termination and exposes the error in the log.
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

// addDiagnostic adds a diagnostic message once when the process gives little
// or no output to explain its termination.
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

// getLastError returns the most recent process exit error, if one exists.
func (s *ProcessSession) getLastError() string {
	if s == nil {
		return ""
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastError
}

// hasExited reports whether the session has observed process termination.
func (s *ProcessSession) hasExited() bool {
	if s == nil {
		return true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exited
}