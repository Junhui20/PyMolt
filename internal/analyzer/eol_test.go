package analyzer

import "testing"

func TestPythonEOL_KnownAndUnknown(t *testing.T) {
	if s := PythonEOL("2.7"); s.Status != "eol" {
		t.Errorf("Python 2.7 should be eol, got %q", s.Status)
	}
	if s := PythonEOL("3.99"); s.Status != "unknown" {
		t.Errorf("unknown version should report 'unknown', got %q", s.Status)
	}
	if s := PythonEOL(""); s.Status != "unknown" {
		t.Errorf("empty version should report 'unknown', got %q", s.Status)
	}
}

func TestPythonEOL_RecentVersionsAreClassified(t *testing.T) {
	// Avoid asserting supported-vs-security for a given date (that changes over
	// time); just require these lines resolve to a concrete, non-unknown status.
	for _, mm := range []string{"3.11", "3.12", "3.13", "3.14"} {
		s := PythonEOL(mm)
		if s.Status == "" || s.Status == "unknown" {
			t.Errorf("PythonEOL(%q) should be a known status, got %q", mm, s.Status)
		}
		if s.EndOfLife == "" {
			t.Errorf("PythonEOL(%q) should carry an end-of-life date", mm)
		}
	}
}
