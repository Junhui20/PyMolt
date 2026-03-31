//go:build !windows

package analyzer

import (
	"os/exec"
)

// hideWindow is a no-op on Unix (no console window to hide).
func hideWindow(cmd *exec.Cmd) {
	_ = cmd
}

// newConsole is a no-op on Unix (terminal apps handle their own windows).
func newConsole(cmd *exec.Cmd) {
	_ = cmd
}
