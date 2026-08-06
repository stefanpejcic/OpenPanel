package php

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// fetchLitespeedTags paginates the relevant litespeedtech Docker Hub repo's
// tags, extracts the lsphpXY suffix from each, and returns the distinct
// versions sorted desc.
func fetchLitespeedTags(ctx context.Context, webServer string) []string {
	var url string
	switch webServer {
	case "openlitespeed":
		url = "https://hub.docker.com/v2/repositories/litespeedtech/openlitespeed/tags?page_size=10"
	case "litespeed":
		url = "https://hub.docker.com/v2/repositories/litespeedtech/litespeed/tags?page_size=10"
	default:
		return nil
	}

	lsphpRE := regexp.MustCompile(`lsphp(\d)(\d)`)
	seen := map[string]bool{}
	var versions []string

	client := &http.Client{Timeout: 5 * time.Second}
	for url != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			break
		}
		resp, err := client.Do(req)
		if err != nil {
			break
		}
		var data struct {
			Results []struct {
				Name string `json:"name"`
			} `json:"results"`
			Next string `json:"next"`
		}
		decErr := json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK || decErr != nil {
			break
		}

		for _, result := range data.Results {
			if m := lsphpRE.FindStringSubmatch(result.Name); m != nil {
				version := m[1] + "." + m[2]
				if !seen[version] {
					seen[version] = true
					versions = append(versions, version)
				}
			}
		}
		url = data.Next
	}

	sortVersionsDesc(versions)
	return versions
}

// phpAPIVersionsFile is the on-disk cache of the PHP versions API response.
const phpAPIVersionsFile = "/etc/openpanel/openpanel/php/api_versions.json"

// phpAPIVersionsTTL is the freshness window for phpAPIVersionsFile.
const phpAPIVersionsTTL = 72 * time.Hour

// VersionInfo is the per-version status shape returned by the PHP versions API.
type VersionInfo struct {
	StatusLabel     string `json:"statusLabel"`
	IsEOLVersion    bool   `json:"isEOLVersion"`
	IsSecureVersion bool   `json:"isSecureVersion"`
	IsLatestVersion bool   `json:"isLatestVersion"`
	IsFutureVersion bool   `json:"isFutureVersion"`
	IsNextVersion   bool   `json:"isNextVersion"`
}

// fetchPHPVersionsAPI maintains a 3-day disk cache of
// https://api.openpanel.com/php-versions/, used by settings.html to
// badge each installed version (latest/next/future/EOL/secure).
func fetchPHPVersionsAPI(ctx context.Context) map[string]VersionInfo {
	if info, err := os.Stat(phpAPIVersionsFile); err == nil {
		if time.Since(info.ModTime()) < phpAPIVersionsTTL {
			if data, ok := readVersionsAPICache(); ok {
				return data
			}
		}
	}

	client := &http.Client{Timeout: 3 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.openpanel.com/php-versions/", nil)
	if err != nil {
		return map[string]VersionInfo{}
	}
	resp, err := client.Do(req)
	if err != nil {
		os.Remove(phpAPIVersionsFile)
		return map[string]VersionInfo{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]VersionInfo{}
	}

	var body struct {
		Data map[string]struct {
			Name            string `json:"name"`
			StatusLabel     string `json:"statusLabel"`
			IsEOLVersion    bool   `json:"isEOLVersion"`
			IsSecureVersion bool   `json:"isSecureVersion"`
			IsLatestVersion bool   `json:"isLatestVersion"`
			IsFutureVersion bool   `json:"isFutureVersion"`
			IsNextVersion   bool   `json:"isNextVersion"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return map[string]VersionInfo{}
	}

	versionsData := make(map[string]VersionInfo, len(body.Data))
	for _, v := range body.Data {
		versionsData[v.Name] = VersionInfo{
			StatusLabel:     v.StatusLabel,
			IsEOLVersion:    v.IsEOLVersion,
			IsSecureVersion: v.IsSecureVersion,
			IsLatestVersion: v.IsLatestVersion,
			IsFutureVersion: v.IsFutureVersion,
			IsNextVersion:   v.IsNextVersion,
		}
	}

	_ = os.MkdirAll(filepath.Dir(phpAPIVersionsFile), 0o755)
	if b, err := json.Marshal(versionsData); err == nil {
		_ = os.WriteFile(phpAPIVersionsFile, b, 0o644)
	}

	return versionsData
}

func readVersionsAPICache() (map[string]VersionInfo, bool) {
	content, err := os.ReadFile(phpAPIVersionsFile)
	if err != nil {
		return nil, false
	}
	var data map[string]VersionInfo
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, false
	}
	return data, true
}
