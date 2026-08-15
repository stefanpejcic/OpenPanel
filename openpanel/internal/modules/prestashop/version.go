package prestashop

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

var prestashopVersionRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
	} `json:"assets"`
}

// listPrestashopVersions hits the GitHub releases API and returns every
// stable version that actually ships a `prestashop_X.Y.Z.zip` release
// asset, newest first. Confirmed live: PrestaShop only stopped attaching
// that asset starting with its 9.x line (releases.prestashop.com became the
// sole distribution channel then) - the `assets` array is empty for those,
// so filtering on its presence naturally limits the list to installable
// versions instead of ones with nothing to download.
func listPrestashopVersions(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/PrestaShop/PrestaShop/releases?per_page=40", nil)
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
		if !prestashopVersionRE.MatchString(rel.TagName) {
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
		return nil, errors.New("no PrestaShop versions found")
	}
	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// latestPrestashopVersion returns the highest available installable
// version - used server-side when the install form's version field is left
// blank.
func latestPrestashopVersion(ctx context.Context) (string, error) {
	versions, err := listPrestashopVersions(ctx)
	if err != nil {
		return "", err
	}
	return versions[0], nil
}

// compareVersions compares two dotted numeric versions ("8.2.7" vs
// "1.7.10"); returns >0 if a > b.
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
