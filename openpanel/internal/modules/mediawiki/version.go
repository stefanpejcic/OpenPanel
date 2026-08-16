package mediawiki

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	mediawikiBranchRE  = regexp.MustCompile(`href="(\d+\.\d+)/"`)
	mediawikiTarballRE = regexp.MustCompile(`href="(mediawiki-(\d+\.\d+\.\d+)\.tar\.gz)"`)
)

func fetchDirListing(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	// releases.wikimedia.org returns 403 to Go's default "Go-http-client"
	// User-Agent - confirmed live, a plain browser/curl-like UA is required.
	req.Header.Set("User-Agent", "OpenPanel/1.0 (+https://openpanel.com)")
	client := &http.Client{Timeout: 15 * time.Second}
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

// listMediaWikiBranches scrapes releases.wikimedia.org/mediawiki/'s
// top-level listing for branch directories (e.g. "1.42/", "1.43/") -
// confirmed live: this index lists nothing but branch directories, newest
// first is not guaranteed so the result is sorted numerically here.
func listMediaWikiBranches(ctx context.Context) ([]string, error) {
	body, err := fetchDirListing(ctx, "https://releases.wikimedia.org/mediawiki/")
	if err != nil {
		return nil, err
	}
	matches := mediawikiBranchRE.FindAllStringSubmatch(body, -1)
	seen := map[string]bool{}
	var branches []string
	for _, m := range matches {
		if !seen[m[1]] {
			seen[m[1]] = true
			branches = append(branches, m[1])
		}
	}
	if len(branches) == 0 {
		return nil, errors.New("no MediaWiki release branches found")
	}
	sort.Slice(branches, func(i, j int) bool { return compareVersions(branches[i], branches[j]) > 0 })
	return branches, nil
}

// listMediaWikiVersionsForBranch scrapes one branch directory's listing for
// patch-level tarballs (e.g. "mediawiki-1.42.7.tar.gz"), newest first.
func listMediaWikiVersionsForBranch(ctx context.Context, branch string) ([]string, error) {
	body, err := fetchDirListing(ctx, "https://releases.wikimedia.org/mediawiki/"+branch+"/")
	if err != nil {
		return nil, err
	}
	matches := mediawikiTarballRE.FindAllStringSubmatch(body, -1)
	var versions []string
	for _, m := range matches {
		versions = append(versions, m[2])
	}
	if len(versions) == 0 {
		return nil, errors.New("no MediaWiki tarballs found for branch " + branch)
	}
	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// listMediaWikiVersions returns every patch-level version across every
// branch, newest first - used by the install form's version dropdown.
func listMediaWikiVersions(ctx context.Context) ([]string, error) {
	branches, err := listMediaWikiBranches(ctx)
	if err != nil {
		return nil, err
	}
	var all []string
	for _, branch := range branches {
		versions, verErr := listMediaWikiVersionsForBranch(ctx, branch)
		if verErr != nil {
			continue
		}
		all = append(all, versions...)
	}
	if len(all) == 0 {
		return nil, errors.New("no MediaWiki versions found")
	}
	sort.Slice(all, func(i, j int) bool { return compareVersions(all[i], all[j]) > 0 })
	return all, nil
}

// latestMediaWikiVersion returns the highest available patch version off
// the highest available branch - used server-side when the install form's
// version field is left blank.
func latestMediaWikiVersion(ctx context.Context) (string, error) {
	branches, err := listMediaWikiBranches(ctx)
	if err != nil {
		return "", err
	}
	for _, branch := range branches {
		versions, verErr := listMediaWikiVersionsForBranch(ctx, branch)
		if verErr != nil || len(versions) == 0 {
			continue
		}
		return versions[0], nil
	}
	return "", errors.New("no MediaWiki versions found")
}

// mediawikiBranchForVersion converts a "X.Y.Z"-style patch version back
// into its "X.Y" branch directory name.
func mediawikiBranchForVersion(version string) string {
	parts := strings.SplitN(version, ".", 3)
	if len(parts) < 2 {
		return ""
	}
	return parts[0] + "." + parts[1]
}

// mediawikiComposerPHPRequirementRE pulls the minimum PHP version out of a
// downloaded release's own composer.json ("php": ">=8.3.0") - read straight
// from the archive rather than hardcoding a branch->PHP table, since the
// minimum climbs across branches (1.39 LTS wants 7.4.3+, 1.46 wants 8.3+ -
// confirmed live by downloading both and reading their composer.json).
var mediawikiComposerPHPRequirementRE = regexp.MustCompile(`"php"\s*:\s*"[^0-9]*(\d+\.\d+(?:\.\d+)?)`)

// minPHPVersionFromComposerJSON reads composer.json's require.php constraint
// out of an already-extracted MediaWiki tree. Returns "" if it can't be
// determined (e.g. some future release restructures composer.json) - the
// caller treats that as "unknown, don't block the install".
func minPHPVersionFromComposerJSON(installDir string) string {
	content, err := os.ReadFile(filepath.Join(installDir, "composer.json"))
	if err != nil {
		return ""
	}
	m := mediawikiComposerPHPRequirementRE.FindSubmatch(content)
	if m == nil {
		return ""
	}
	return string(m[1])
}

// compareVersions compares two dotted numeric versions ("1.42.7" vs
// "1.9.3"); returns >0 if a > b.
func compareVersions(a, b string) int {
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
