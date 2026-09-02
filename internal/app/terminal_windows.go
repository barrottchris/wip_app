//go:build windows

package app

import (
	"fmt"
	"os/exec"
	"syscall"
)

func prepareVisibleTerminal(cmd *exec.Cmd) error {
	if cmd == nil {
		return fmt.Errorf("terminal command is required")
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	return nil
}

func captureTerminalOutput() bool {
	return true
}