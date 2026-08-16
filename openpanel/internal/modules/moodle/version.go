package moodle

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

var moodleVersionRE = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

type githubTag struct {
	Name string `json:"name"`
}

// listMoodleVersions hits the GitHub tags API and returns every stable
// "vX.Y.Z" tag (confirmed live: prerelease tags carry a "-rcN"/"-beta"
// suffix and are excluded by the exact-three-part regex), newest first.
// Unlike prestashop/nextcloud's releases-API approach, Moodle's GitHub tags
// aren't filtered on release-asset presence, since the real download
// artifact lives on download.moodle.org instead (see install.go's
// moodleBranch), not as a GitHub release asset.
func listMoodleVersions(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/moodle/moodle/tags?per_page=100", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	client := &http.Client{Timeout: 15 * time.Second}
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
	if unmarshalErr := json.Unmarshal(body, &tags); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	versions := make([]string, 0, len(tags))
	for _, t := range tags {
		if moodleVersionRE.MatchString(t.Name) {
			versions = append(versions, strings.TrimPrefix(t.Name, "v"))
		}
	}
	if len(versions) == 0 {
		return nil, errors.New("no Moodle versions found")
	}
	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// latestMoodleVersion returns the highest available version - used
// server-side when the install form's version field is left blank.
func latestMoodleVersion(ctx context.Context) (string, error) {
	versions, err := listMoodleVersions(ctx)
	if err != nil {
		return "", err
	}
	return versions[0], nil
}

// compareVersions compares two dotted numeric versions ("5.2.2" vs
// "4.5.10"); returns >0 if a > b.
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
