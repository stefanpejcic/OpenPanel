package phpbb

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

var listPhpbbVersionsRE = regexp.MustCompile(`^release-(\d+\.\d+\.\d+)$`)

// listPhpbbVersions hits the GitHub tags API and returns every stable
// "release-X.Y.Z" tag (stripped of that prefix), newest first - mirrors
// the exact filtering phpbb_install.html's own client-side JS already
// does, so the server-side "latest" fallback below can never disagree
// with what's shown in the UI.
func listPhpbbVersions(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/phpbb/phpbb/tags?per_page=100", nil)
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

	var tags []struct {
		Name string `json:"name"`
	}
	if unmarshalErr := json.Unmarshal(body, &tags); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	versions := make([]string, 0, len(tags))
	for _, t := range tags {
		if m := listPhpbbVersionsRE.FindStringSubmatch(t.Name); m != nil {
			versions = append(versions, m[1])
		}
	}
	if len(versions) == 0 {
		return nil, errors.New("no phpBB versions found")
	}
	sort.Slice(versions, func(i, j int) bool { return comparePhpbbVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// latestPhpbbVersion returns the highest available stable version - used
// server-side when the install form's version field is left blank
// ("Latest").
func latestPhpbbVersion(ctx context.Context) (string, error) {
	versions, err := listPhpbbVersions(ctx)
	if err != nil {
		return "", err
	}
	return versions[0], nil
}

// comparePhpbbVersions compares two dotted numeric versions ("3.3.16" vs
// "3.3.17"); returns >0 if a > b.
func comparePhpbbVersions(a, b string) int {
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
