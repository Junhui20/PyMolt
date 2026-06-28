//go:build !windows

package analyzer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Junhui20/PyMolt/internal/models"
)

// OpenTerminal opens a new terminal window with the given Python or venv activated.
func OpenTerminal(inst models.PythonInstallation) error {
	// Defense in depth: a path containing newlines cannot be safely embedded in a
	// shell command or AppleScript string, and no legitimate Python path has them.
	if strings.ContainsAny(inst.Path, "\n\r") {
		return fmt.Errorf("refusing to open terminal: path contains invalid characters")
	}

	var shellCmd string

	if inst.Source == models.SourceVenv {
		activateScript := filepath.Join(inst.Path, "bin", "activate")
		if _, err := os.Stat(activateScript); err != nil {
			return fmt.Errorf("activate script not found in %s", inst.Path)
		}
		// shellQuote prevents shell metacharacters in the path from being executed.
		shellCmd = "source " + shellQuote(activateScript) + " && exec $SHELL"
	} else {
		shellCmd = "export PATH=" + shellQuote(inst.Path) + ":$PATH && python3 --version && exec $SHELL"
	}

	var cmd *exec.Cmd

	if runtime.GOOS == "darwin" {
		// Use osascript to open Terminal.app. The shell command is escaped for the
		// AppleScript string context so it cannot break out of `do script "..."`.
		script := fmt.Sprintf(`tell application "Terminal" to do script "%s"`, appleScriptEscape(shellCmd))
		cmd = exec.Command("osascript", "-e", script)
	} else {
		// Linux: try common terminal emulators
		terminal := findTerminalEmulator()
		if terminal == "" {
			return fmt.Errorf("no terminal emulator found")
		}
		cmd = exec.Command(terminal, "-e", "bash", "-c", shellCmd)
	}

	return cmd.Start()
}

// shellQuote wraps a string in single quotes for safe use in a POSIX shell,
// escaping any embedded single quotes. This neutralizes shell metacharacters in
// paths (e.g. "$(...)", ";", "&&") so they are treated as literal text.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// appleScriptEscape escapes backslashes and double quotes so a string can be
// safely embedded inside an AppleScript double-quoted string literal.
func appleScriptEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// AddToPATH adds a Python installation's directory to the PATH.
func AddToPATH(pythonDir string) error {
	return SetDefaultPython(pythonDir)
}

func findTerminalEmulator() string {
	for _, term := range []string{
		"gnome-terminal", "konsole", "xfce4-terminal",
		"mate-terminal", "tilix", "alacritty", "kitty",
		"xterm",
	} {
		if path, err := exec.LookPath(term); err == nil {
			return path
		}
	}
	return ""
}
