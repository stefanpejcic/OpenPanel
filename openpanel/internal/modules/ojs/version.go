package ojs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ojsVersionRE matches pkp/ojs's GitHub tag naming - inconsistent across
// history: newer tags are bare "3_5_0-5", older ones are prefixed
// "ojs-3_1_2-4" (confirmed live against the tags API). A real release tag
// always ends in "-<build>"; anything else (e.g. "3_4_0rc3", beta/alpha
// tags) is a pre-release and is excluded by this pattern requiring the
// string to end right after the build number.
var ojsVersionRE = regexp.MustCompile(`^(?:ojs-)?(\d+)_(\d+)_(\d+)-(\d+)$`)

type githubTag struct {
	Name string `json:"name"`
}

// ojsVersion is one parsed, displayable OJS release.
type ojsVersion struct {
	Dotted                 string // e.g. "3.5.0-5" - used in the download URL and shown to the user
	Major, Minor, Patch, Build int
}

// listOJSVersions hits the GitHub tags API (paginated, since the full
// history of real releases spans more than one page of 100) and returns
// every stable "x.y.z-build" release, newest first. Pre-release tags
// (rc/beta/alpha) are excluded by ojsVersionRE itself.
func listOJSVersions(ctx context.Context) ([]ojsVersion, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	var versions []ojsVersion
	for page := 1; page <= 3; page++ {
		tags, err := fetchOJSTagsPage(ctx, client, page)
		if err != nil {
			return nil, err
		}
		if len(tags) == 0 {
			break
		}
		for _, t := range tags {
			m := ojsVersionRE.FindStringSubmatch(t.Name)
			if m == nil {
				continue
			}
			if strings.Contains(strings.ToLower(t.Name), "rc") || strings.Contains(strings.ToLower(t.Name), "beta") || strings.Contains(strings.ToLower(t.Name), "alpha") {
				continue
			}
			major, _ := strconv.Atoi(m[1])
			minor, _ := strconv.Atoi(m[2])
			patch, _ := strconv.Atoi(m[3])
			build, _ := strconv.Atoi(m[4])
			versions = append(versions, ojsVersion{
				Dotted: m[1] + "." + m[2] + "." + m[3] + "-" + m[4],
				Major:  major, Minor: minor, Patch: patch, Build: build,
			})
		}
		if len(tags) < 100 {
			break
		}
	}
	if len(versions) == 0 {
		return nil, errors.New("no OJS versions found")
	}
	sort.Slice(versions, func(i, j int) bool { return compareOJSVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

func fetchOJSTagsPage(ctx context.Context, client *http.Client, page int) ([]githubTag, error) {
	url := "https://api.github.com/repos/pkp/ojs/tags?per_page=100&page=" + strconv.Itoa(page)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tags []githubTag
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

// latestOJSVersion returns the highest available version - used
// server-side when the install form's version field is left blank.
func latestOJSVersion(ctx context.Context) (ojsVersion, error) {
	versions, err := listOJSVersions(ctx)
	if err != nil {
		return ojsVersion{}, err
	}
	return versions[0], nil
}

// findOJSVersion resolves a "3.5.0-5"-style dotted version string (as
// submitted by the install form) back into an ojsVersion, re-validating it
// against the live tag list so an install can't be pointed at an
// arbitrary/non-existent tarball URL.
func findOJSVersion(ctx context.Context, dotted string) (ojsVersion, error) {
	versions, err := listOJSVersions(ctx)
	if err != nil {
		return ojsVersion{}, err
	}
	for _, v := range versions {
		if v.Dotted == dotted {
			return v, nil
		}
	}
	return ojsVersion{}, errors.New("unknown OJS version " + dotted)
}

// compareOJSVersions returns >0 if a > b.
func compareOJSVersions(a, b ojsVersion) int {
	if a.Major != b.Major {
		return a.Major - b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor - b.Minor
	}
	if a.Patch != b.Patch {
		return a.Patch - b.Patch
	}
	return a.Build - b.Build
}

// ojsDownloadURL builds PKP's own direct-hosted, submodule-bundled release
// package URL for a dotted version (see ojs.go's package doc comment for
// why GitHub's own archive endpoints can't be used instead).
func ojsDownloadURL(dotted string) string {
	return "https://pkp.sfu.ca/ojs/download/ojs-" + dotted + ".tar.gz"
}
