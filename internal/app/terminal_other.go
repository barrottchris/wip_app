//go:build !windows

package app

import "os/exec"

func prepareVisibleTerminal(cmd *exec.Cmd) error {
	return nil
}

func captureTerminalOutput() bool {
	return true
}