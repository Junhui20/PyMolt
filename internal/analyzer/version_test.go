package analyzer

import "testing"

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"3.12.0", "3.12.0", 0},
		{"3.12", "3.12.0", 0},   // missing trailing segment treated as 0
		{"0.10.0", "0.9.0", 1},  // the bug the old lexical compare had
		{"0.9.0", "0.10.0", -1}, // ...and its mirror
		{"3.13.1", "3.13.0", 1},
		{"v1.2.3", "1.2.3", 0},     // leading "v" is stripped
		{"3.12.4rc1", "3.12.4", 0}, // non-numeric suffix ignored
		{"2.7.18", "3.0.0", -1},
	}
	for _, c := range cases {
		if got := compareVersions(c.a, c.b); got != c.want {
			t.Errorf("compareVersions(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
