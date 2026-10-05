package prestashop

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var prestashopVersionRE = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
	} `json:"assets"`
}

// listPrestashopVersions hits the GitHub releases API and returns every stable version that actually ships a prestashop_X.Y.Z.zip release asset, newest first - PrestaShop stopped attaching that asset starting with 9.x, so filtering on its presence naturally limits the list to installable versions
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
	sort.Slice(versions, func(i, j int) bool { return cmsapp.CompareVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// latestPrestashopVersion returns the highest available installable version, used server-side when the install form's version field is left blank
func latestPrestashopVersion(ctx context.Context) (string, error) {
	versions, err := listPrestashopVersions(ctx)
	if err != nil {
		return "", err
	}
	return versions[0], nil
}
