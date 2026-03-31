package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MarketplacePackage represents a package from the catalog.
type MarketplacePackage struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Homepage    string `json:"homepage"`
	Stars       int    `json:"stars"`
	PypiID      string `json:"pypiId"`
}

// PyPIPackageDetail has real-time info from PyPI JSON API.
type PyPIPackageDetail struct {
	Name           string   `json:"name"`
	Version        string   `json:"version"`
	Summary        string   `json:"summary"`
	Author         string   `json:"author"`
	License        string   `json:"license"`
	HomePage       string   `json:"homePage"`
	RequiresPython string   `json:"requiresPython"`
	Keywords       string   `json:"keywords"`
	Versions       []string `json:"versions"`
}

var (
	catalogCache []MarketplacePackage
	catalogMu    sync.RWMutex
	cacheFile    = filepath.Join(os.Getenv("APPDATA"), "PythonManager", "catalog.json")
)

// LoadCatalog fetches the awesome-python dataset from GitHub or local cache.
func LoadCatalog() ([]MarketplacePackage, error) {
	catalogMu.RLock()
	if catalogCache != nil {
		defer catalogMu.RUnlock()
		return catalogCache, nil
	}
	catalogMu.RUnlock()

	// Try local cache first (valid for 24h)
	if data, err := os.ReadFile(cacheFile); err == nil {
		info, _ := os.Stat(cacheFile)
		if info != nil && time.Since(info.ModTime()) < 24*time.Hour {
			var pkgs []MarketplacePackage
			if json.Unmarshal(data, &pkgs) == nil && len(pkgs) > 0 {
				catalogMu.Lock()
				catalogCache = pkgs
				catalogMu.Unlock()
				return pkgs, nil
			}
		}
	}

	// Fetch from GitHub
	pkgs, err := fetchAwesomePython()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch catalog: %w", err)
	}

	// Save to local cache
	cacheDir := filepath.Dir(cacheFile)
	os.MkdirAll(cacheDir, 0755)
	if data, err := json.Marshal(pkgs); err == nil {
		os.WriteFile(cacheFile, data, 0644)
	}

	catalogMu.Lock()
	catalogCache = pkgs
	catalogMu.Unlock()

	return pkgs, nil
}

func fetchAwesomePython() ([]MarketplacePackage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := "https://raw.githubusercontent.com/dylanhogg/awesome-python/main/github_data.json"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "PythonManager/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// awesome-python JSON structure: {"schema":{...}, "data":[...]}
	var wrapper struct {
		Data []struct {
			Category    string `json:"category"`
			RepoName    string `json:"_reponame"`
			Stars       int    `json:"_stars"`
			Homepage    string `json:"_homepage"`
			GHDesc      string `json:"_github_description"`
			Description string `json:"_description"`
			GithubURL   string `json:"githuburl"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &wrapper); err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	var pkgs []MarketplacePackage
	for _, r := range wrapper.Data {
		name := r.RepoName
		if name == "" {
			continue
		}
		desc := r.GHDesc
		if desc == "" {
			desc = r.Description
		}
		homepage := r.GithubURL
		if homepage == "" {
			homepage = r.Homepage
		}

		pkgs = append(pkgs, MarketplacePackage{
			Name:        name,
			Description: desc,
			Category:    r.Category,
			Homepage:    homepage,
			Stars:       r.Stars,
			PypiID:      name, // most Python packages use repo name as PyPI name
		})
	}

	return pkgs, nil
}

// GetCategories returns all unique categories from the catalog.
func GetCategories(pkgs []MarketplacePackage) []string {
	seen := map[string]bool{}
	var cats []string
	for _, p := range pkgs {
		if p.Category != "" && !seen[p.Category] {
			seen[p.Category] = true
			cats = append(cats, p.Category)
		}
	}
	return cats
}

// SearchCatalog filters the catalog by query string.
func SearchCatalog(pkgs []MarketplacePackage, query string) []MarketplacePackage {
	if query == "" {
		return pkgs
	}
	q := strings.ToLower(query)
	var results []MarketplacePackage
	for _, p := range pkgs {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Description), q) ||
			strings.Contains(strings.ToLower(p.PypiID), q) {
			results = append(results, p)
		}
	}
	return results
}

// FetchPyPIDetail gets real-time package info from PyPI JSON API.
func FetchPyPIDetail(packageName string) (*PyPIPackageDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", packageName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "PythonManager/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("package not found on PyPI")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		Info struct {
			Name           string `json:"name"`
			Version        string `json:"version"`
			Summary        string `json:"summary"`
			Author         string `json:"author"`
			License        string `json:"license"`
			HomePage       string `json:"home_page"`
			RequiresPython string `json:"requires_python"`
			Keywords       string `json:"keywords"`
		} `json:"info"`
		Releases map[string]json.RawMessage `json:"releases"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	// Get last 10 versions
	var versions []string
	for v := range data.Releases {
		versions = append(versions, v)
	}
	// Sort descending (simple reverse since PyPI returns them roughly ordered)
	if len(versions) > 10 {
		versions = versions[len(versions)-10:]
	}
	// Reverse
	for i, j := 0, len(versions)-1; i < j; i, j = i+1, j-1 {
		versions[i], versions[j] = versions[j], versions[i]
	}

	return &PyPIPackageDetail{
		Name:           data.Info.Name,
		Version:        data.Info.Version,
		Summary:        data.Info.Summary,
		Author:         data.Info.Author,
		License:        data.Info.License,
		HomePage:       data.Info.HomePage,
		RequiresPython: data.Info.RequiresPython,
		Keywords:       data.Info.Keywords,
		Versions:       versions,
	}, nil
}

// SearchPyPI searches by trying exact match on PyPI, then falls back to catalog search.
func SearchPyPI(query string) ([]MarketplacePackage, error) {
	if query == "" {
		return nil, nil
	}

	// Try exact match on PyPI
	detail, err := FetchPyPIDetail(query)
	if err == nil && detail != nil {
		return []MarketplacePackage{{
			Name:        detail.Name,
			Description: detail.Summary,
			Category:    "PyPI",
			Homepage:    detail.HomePage,
			PypiID:      detail.Name,
		}}, nil
	}

	// Fall back to catalog search
	catalog, err := LoadCatalog()
	if err != nil {
		return nil, err
	}
	return SearchCatalog(catalog, query), nil
}
