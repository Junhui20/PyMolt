package analyzer

import (
	"time"

	"github.com/Junhui20/PyMolt/internal/models"
)

// pythonEOL maps a Python "major.minor" line to {endOfFullSupport, endOfLife}
// dates (YYYY-MM-DD). endOfFullSupport is when bugfix releases stop and the
// version becomes security-only; endOfLife is when security support ends.
//
// This is a bundled snapshot. Sources: https://devguide.python.org/versions/
// and https://endoflife.date/python. A future enhancement (see docs/ROADMAP.md)
// can refresh it from endoflife.date at runtime and cache it like the catalog.
var pythonEOL = map[string][2]string{
	"2.7":  {"2020-01-01", "2020-01-01"},
	"3.0":  {"2009-06-27", "2009-06-27"},
	"3.1":  {"2012-04-09", "2012-04-09"},
	"3.2":  {"2016-02-20", "2016-02-20"},
	"3.3":  {"2017-09-29", "2017-09-29"},
	"3.4":  {"2019-03-18", "2019-03-18"},
	"3.5":  {"2017-09-13", "2020-09-30"},
	"3.6":  {"2018-12-24", "2021-12-23"},
	"3.7":  {"2020-06-27", "2023-06-27"},
	"3.8":  {"2021-05-03", "2024-10-07"},
	"3.9":  {"2022-05-17", "2025-10-31"},
	"3.10": {"2023-04-05", "2026-10-31"},
	"3.11": {"2024-04-02", "2027-10-31"},
	"3.12": {"2025-04-02", "2028-10-31"},
	"3.13": {"2026-10-31", "2029-10-31"},
	"3.14": {"2027-10-31", "2030-10-31"},
	"3.15": {"2028-10-31", "2031-10-31"},
}

// PythonEOL returns the upstream support status for a "major.minor" version
// string (e.g. "3.12"). Unknown versions return status "unknown".
func PythonEOL(majorMinor string) models.EOLStatus {
	rng, ok := pythonEOL[majorMinor]
	if !ok {
		return models.EOLStatus{Status: "unknown", Label: "Support status unknown"}
	}

	bugfixEnd, _ := time.Parse("2006-01-02", rng[0])
	eolEnd, _ := time.Parse("2006-01-02", rng[1])
	now := time.Now()

	st := models.EOLStatus{EndOfFullSupport: rng[0], EndOfLife: rng[1]}
	switch {
	case now.Before(bugfixEnd):
		st.Status = "supported"
		st.Label = "Actively supported (bug + security fixes)"
	case now.Before(eolEnd):
		st.Status = "security"
		st.Label = "Security fixes only until " + rng[1] + " — consider upgrading"
	default:
		st.Status = "eol"
		st.Label = "End of life since " + rng[1] + " — no more security updates"
	}
	return st
}
