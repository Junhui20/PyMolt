package analyzer

import (
	"fmt"
	"strings"
)

// cleanArg validates and trims a value before it is passed as a command-line
// argument to pip or uv. No shell is involved (commands run via exec with an
// argv slice), so shell metacharacters are harmless; the real risks are a value
// that begins with "-" (which pip/uv would treat as a flag — argument injection)
// or one that smuggles extra arguments via whitespace. Returns the trimmed value
// or an error describing why it was rejected. kind is used in the error message
// (e.g. "package name", "version").
func cleanArg(kind, value string) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", fmt.Errorf("%s cannot be empty", kind)
	}
	if strings.HasPrefix(v, "-") {
		return "", fmt.Errorf("invalid %s %q: must not start with '-'", kind, value)
	}
	if strings.ContainsAny(v, " \t\r\n\x00") {
		return "", fmt.Errorf("invalid %s %q: must not contain whitespace", kind, value)
	}
	return v, nil
}
