package flarum

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

var flarumVersionRE = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)

// listFlarumVersions hits the GitHub tags API and returns every stable
// "vX.Y.Z" tag, newest first - mirrors the exact filtering the install/
// update tab's own client-side JS already does (flarum_install.html /
// flarum_app.html's versionRe), so the server-side "latest" fallback below
// can never disagree with what's shown in the UI. flarum/core has no
// stable v2.0.0 yet (still v2.0.0-rc.N/beta.N as of this writing - verified
// live against the real tags feed), and this regex excludes every such
// pre-release tag, same as the frontend's.
func listFlarumVersions(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/flarum/core/tags?per_page=100", nil)
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
		if flarumVersionRE.MatchString(t.Name) {
			versions = append(versions, strings.TrimPrefix(t.Name, "v"))
		}
	}
	if len(versions) == 0 {
		return nil, errors.New("no Flarum versions found")
	}
	sort.Slice(versions, func(i, j int) bool { return compareFlarumVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// latestFlarumVersion returns the highest available stable version - used
// server-side when the install form's version field is left blank/"latest".
func latestFlarumVersion(ctx context.Context) (string, error) {
	versions, err := listFlarumVersions(ctx)
	if err != nil {
		return "", err
	}
	return versions[0], nil
}

// compareFlarumVersions compares two dotted numeric versions ("1.8.19" vs
// "2.0.0"); returns >0 if a > b.
func compareFlarumVersions(a, b string) int {
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
