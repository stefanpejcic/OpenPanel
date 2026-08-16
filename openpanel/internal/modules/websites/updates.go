package websites

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

// This file mirrors the latest-version lookups every CMS module's own
// version.go already has (see internal/modules/joomla/version.go,
// prestashop/version.go, etc.) - duplicated locally rather than imported,
// same "websites doesn't cross-import the CMS packages" convention
// websites.go's getMoodleVersion/extractMoodleDatabaseInfo etc. already
// establish (and those functions are unexported in their own packages
// anyway, so a cross-import wouldn't compile). Used by handleSitesUpdates
// to tell /sites which installed sites have a newer version available.

func compareUpdateVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")
	for i := 0; i < len(partsA) || i < len(partsB); i++ {
		var na, nb int
		if i < len(partsA) {
			na, _ = strconv.Atoi(partsA[i])
		}
		if i < len(partsB) {
			nb, _ = strconv.Atoi(partsB[i])
		}
		if na != nb {
			return na - nb
		}
	}
	return 0
}

func fetchLatestGitHubRelease(ctx context.Context, repo string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var release struct {
		TagName string `json:"tag_name"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&release); decodeErr != nil {
		return "", decodeErr
	}
	return strings.TrimPrefix(release.TagName, "v"), nil
}

var updatesDrupalTagRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func latestDrupalVersionForUpdates(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/drupal/drupal/tags?per_page=100", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var tags []struct {
		Name string `json:"name"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&tags); decodeErr != nil {
		return "", decodeErr
	}
	var versions []string
	for _, t := range tags {
		if updatesDrupalTagRE.MatchString(t.Name) {
			versions = append(versions, t.Name)
		}
	}
	if len(versions) == 0 {
		return "", nil
	}
	sort.Slice(versions, func(i, j int) bool { return compareUpdateVersions(versions[i], versions[j]) > 0 })
	return versions[0], nil
}

var updatesNextcloudArchiveRE = regexp.MustCompile(`nextcloud-(\d+\.\d+\.\d+)\.zip`)

func latestNextcloudVersionForUpdates(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://download.nextcloud.com/server/releases/", nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	matches := updatesNextcloudArchiveRE.FindAllStringSubmatch(string(body), -1)
	var versions []string
	for _, m := range matches {
		versions = append(versions, m[1])
	}
	if len(versions) == 0 {
		return "", nil
	}
	sort.Slice(versions, func(i, j int) bool { return compareUpdateVersions(versions[i], versions[j]) > 0 })
	return versions[0], nil
}

var updatesGithubReleaseTagRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

// latestVersionWithZipAsset mirrors prestashop/version.go's and
// matomo/version.go's "only count a release that actually ships a
// downloadable asset" filter.
func latestVersionWithZipAsset(ctx context.Context, repo string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/"+repo+"/releases?per_page=40", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var releases []struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name string `json:"name"`
		} `json:"assets"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&releases); decodeErr != nil {
		return "", decodeErr
	}
	var versions []string
	for _, rel := range releases {
		if !updatesGithubReleaseTagRE.MatchString(rel.TagName) {
			continue
		}
		hasZip := false
		for _, asset := range rel.Assets {
			if strings.HasSuffix(asset.Name, ".zip") {
				hasZip = true
				break
			}
		}
		if hasZip {
			versions = append(versions, rel.TagName)
		}
	}
	if len(versions) == 0 {
		return "", nil
	}
	sort.Slice(versions, func(i, j int) bool { return compareUpdateVersions(versions[i], versions[j]) > 0 })
	return versions[0], nil
}

var updatesMoodleTagRE = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

func latestMoodleVersionForUpdates(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/moodle/moodle/tags?per_page=100", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var tags []struct {
		Name string `json:"name"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&tags); decodeErr != nil {
		return "", decodeErr
	}
	var versions []string
	for _, t := range tags {
		if updatesMoodleTagRE.MatchString(t.Name) {
			versions = append(versions, strings.TrimPrefix(t.Name, "v"))
		}
	}
	if len(versions) == 0 {
		return "", nil
	}
	sort.Slice(versions, func(i, j int) bool { return compareUpdateVersions(versions[i], versions[j]) > 0 })
	return versions[0], nil
}

var (
	updatesMediaWikiBranchRE  = regexp.MustCompile(`href="(\d+\.\d+)/"`)
	updatesMediaWikiVersionRE = regexp.MustCompile(`href="mediawiki-(\d+\.\d+\.\d+)\.tar\.gz"`)
)

// latestMediaWikiVersionForUpdates scrapes releases.wikimedia.org/mediawiki/
// (a two-level branch/patch directory listing, not a GitHub-tags API - see
// mediawiki/version.go's identical scraper) for the highest patch version
// off the highest available branch.
func latestMediaWikiVersionForUpdates(ctx context.Context) (string, error) {
	fetch := func(url string) (string, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return "", err
		}
		// releases.wikimedia.org returns 403 to Go's default "Go-http-client"
		// User-Agent - confirmed live, a plain browser/curl-like UA is required.
		req.Header.Set("User-Agent", "OpenPanel/1.0 (+https://openpanel.com)")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	body, err := fetch("https://releases.wikimedia.org/mediawiki/")
	if err != nil {
		return "", err
	}
	branchMatches := updatesMediaWikiBranchRE.FindAllStringSubmatch(body, -1)
	var branches []string
	for _, m := range branchMatches {
		branches = append(branches, m[1])
	}
	sort.Slice(branches, func(i, j int) bool { return compareUpdateVersions(branches[i], branches[j]) > 0 })

	for _, branch := range branches {
		branchBody, branchErr := fetch("https://releases.wikimedia.org/mediawiki/" + branch + "/")
		if branchErr != nil {
			continue
		}
		versionMatches := updatesMediaWikiVersionRE.FindAllStringSubmatch(branchBody, -1)
		var versions []string
		for _, m := range versionMatches {
			versions = append(versions, m[1])
		}
		if len(versions) == 0 {
			continue
		}
		sort.Slice(versions, func(i, j int) bool { return compareUpdateVersions(versions[i], versions[j]) > 0 })
		return versions[0], nil
	}
	return "", nil
}

// latestWordPressVersionForUpdates asks wordpress.org's own stable-check
// API for the current stable release (the same source list.html's
// client-side badge already reads per-version, just used here to get the
// single latest tag server-side).
func latestWordPressVersionForUpdates(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.wordpress.org/core/stable-check/1.0/", nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var statuses map[string]string
	if decodeErr := json.NewDecoder(resp.Body).Decode(&statuses); decodeErr != nil {
		return "", decodeErr
	}
	var latest string
	for version, status := range statuses {
		if status != "latest" {
			continue
		}
		if latest == "" || compareUpdateVersions(version, latest) > 0 {
			latest = version
		}
	}
	return latest, nil
}

// latestVersionsForAllTypes fetches every supported CMS's latest version
// concurrently and returns a lowercase-type -> version map, memoized for
// 6 hours since these barely change and this is 8 outbound API calls.
func latestVersionsForAllTypes(ctx context.Context, a *appctx.App) map[string]string {
	result, _ := cache.Memoize(ctx, a.Cache, "sites_latest_versions", 6*time.Hour, func() (map[string]string, error) {
		fetchers := map[string]func(context.Context) (string, error){
			"wordpress": latestWordPressVersionForUpdates,
			"drupal":    latestDrupalVersionForUpdates,
			"joomla":    func(ctx context.Context) (string, error) { return fetchLatestGitHubRelease(ctx, "joomla/joomla-cms") },
			"opencart":  func(ctx context.Context) (string, error) { return fetchLatestGitHubRelease(ctx, "opencart/opencart") },
			"nextcloud": latestNextcloudVersionForUpdates,
			"prestashop": func(ctx context.Context) (string, error) {
				return latestVersionWithZipAsset(ctx, "PrestaShop/PrestaShop")
			},
			"matomo":    func(ctx context.Context) (string, error) { return latestVersionWithZipAsset(ctx, "matomo-org/matomo") },
			"moodle":    latestMoodleVersionForUpdates,
			"mediawiki": latestMediaWikiVersionForUpdates,
		}

		out := make(map[string]string, len(fetchers))
		var mu sync.Mutex
		var wg sync.WaitGroup
		fetchCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()

		for cmsType, fetch := range fetchers {
			wg.Add(1)
			go func(cmsType string, fetch func(context.Context) (string, error)) {
				defer wg.Done()
				version, err := fetch(fetchCtx)
				if err != nil || version == "" {
					return
				}
				mu.Lock()
				out[cmsType] = version
				mu.Unlock()
			}(cmsType, fetch)
		}
		wg.Wait()
		return out, nil
	})
	return result
}

// handleSitesUpdates returns {"joomla":"5.4.1","wordpress":"6.8.1",...} -
// the latest known version per supported CMS type, for /sites to compare
// each installed site's version against and flag ones that are outdated.
func handleSitesUpdates(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	versions := latestVersionsForAllTypes(r.Context(), a)
	writeJSON(w, http.StatusOK, versions)
}
