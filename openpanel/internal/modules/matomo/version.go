package matomo

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

var matomoVersionRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
	} `json:"assets"`
}

// listMatomoVersions hits the GitHub releases API and returns every stable
// version that ships a matomo-X.Y.Z.zip release asset, newest first -
// confirmed live against the 5.12.0 release that this asset is attached
// directly to the GitHub release (unlike PrestaShop, no separate download
// host is needed).
func listMatomoVersions(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/matomo-org/matomo/releases?per_page=40", nil)
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

	var releases []githubRelease
	if unmarshalErr := json.Unmarshal(body, &releases); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	versions := make([]string, 0, len(releases))
	for _, rel := range releases {
		if !matomoVersionRE.MatchString(rel.TagName) {
			continue
		}
		hasZip := false
		for _, asset := range rel.Assets {
			if asset.Name == "matomo-"+rel.TagName+".zip" {
				hasZip = true
				break
			}
		}
		if hasZip {
			versions = append(versions, rel.TagName)
		}
	}
	if len(versions) == 0 {
		return nil, errors.New("no Matomo versions found")
	}
	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// latestMatomoVersion returns the highest available installable version -
// used server-side when the install form's version field is left blank.
func latestMatomoVersion(ctx context.Context) (string, error) {
	versions, err := listMatomoVersions(ctx)
	if err != nil {
		return "", err
	}
	return versions[0], nil
}

// compareVersions compares two dotted numeric versions ("5.12.0" vs
// "5.9.1"); returns >0 if a > b.
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
