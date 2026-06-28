package analyzer

import "strings"

// compareVersions compares two dotted version strings numerically, returning
// -1 if a < b, 0 if equal, and 1 if a > b. Only the leading integer of each
// dot-separated segment is considered, so a non-numeric suffix (e.g. "4rc1")
// is ignored. Missing trailing segments are treated as 0, so "3.12" == "3.12.0"
// and — unlike a lexical string compare — "0.10.0" > "0.9.0".
func compareVersions(a, b string) int {
	as := splitVersion(a)
	bs := splitVersion(b)
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		var x, y int
		if i < len(as) {
			x = as[i]
		}
		if i < len(bs) {
			y = bs[i]
		}
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	return 0
}

// splitVersion parses the leading integer of each dot-separated segment of v.
func splitVersion(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	nums := make([]int, 0, len(parts))
	for _, p := range parts {
		n := 0
		for i := 0; i < len(p) && p[i] >= '0' && p[i] <= '9'; i++ {
			n = n*10 + int(p[i]-'0')
		}
		nums = append(nums, n)
	}
	return nums
}
