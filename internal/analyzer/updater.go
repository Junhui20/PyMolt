package analyzer

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// UpdateInfo holds version check result.
type UpdateInfo struct {
	CurrentVersion string `json:"currentVersion"`
	LatestVersion  string `json:"latestVersion"`
	UpdateURL      string `json:"updateUrl"`
	HasUpdate      bool   `json:"hasUpdate"`
	ReleaseNotes   string `json:"releaseNotes"`
}

// AppVersion is set at build time via -ldflags.
var AppVersion = "0.2.0"

const githubRepo = "Junhui20/PyMolt"

// CheckForUpdate queries GitHub Releases API for the latest version.
func CheckForUpdate() *UpdateInfo {
	info := &UpdateInfo{CurrentVersion: AppVersion}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := "https://api.github.com/repos/" + githubRepo + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return info
	}
	req.Header.Set("User-Agent", "PyMolt/"+AppVersion)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return info
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return info
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return info
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Body    string `json:"body"`
	}
	if json.Unmarshal(body, &release) != nil {
		return info
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	info.LatestVersion = latest
	info.UpdateURL = release.HTMLURL
	info.ReleaseNotes = release.Body

	if latest != "" && latest != AppVersion && isNewer(latest, AppVersion) {
		info.HasUpdate = true
	}

	return info
}

// isNewer returns true if a > b using simple semver comparison.
func isNewer(a, b string) bool {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		if aParts[i] > bParts[i] {
			return true
		}
		if aParts[i] < bParts[i] {
			return false
		}
	}
	return len(aParts) > len(bParts)
}
