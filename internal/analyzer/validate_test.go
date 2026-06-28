package analyzer

import "testing"

func TestCleanArg_AcceptsValid(t *testing.T) {
	for _, v := range []string{"requests", "flask==2.0.1", "numpy>=1.20", "3.12", "3.13.1"} {
		got, err := cleanArg("package name", v)
		if err != nil {
			t.Errorf("cleanArg(%q) returned error: %v", v, err)
		}
		if got != v {
			t.Errorf("cleanArg(%q) = %q, want unchanged", v, got)
		}
	}
}

func TestCleanArg_RejectsInjection(t *testing.T) {
	// empty, leading-dash (flag injection), and whitespace (argument splitting).
	for _, v := range []string{"", "   ", "-rrequirements.txt", "--upgrade", "foo bar", "a\tb", "x\ny"} {
		if _, err := cleanArg("package name", v); err == nil {
			t.Errorf("cleanArg(%q) should have been rejected", v)
		}
	}
}
