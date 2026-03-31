//go:build windows

package analyzer

import (
	"os/exec"
	"syscall"
)

// hideWindow prevents a console window from appearing when running a command.
func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}

// newConsole sets SysProcAttr to create a new visible console window.
func newConsole(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x00000010, // CREATE_NEW_CONSOLE
	}
}
