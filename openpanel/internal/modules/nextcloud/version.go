package nextcloud

import (
	"context"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sort"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var nextcloudArchiveNameRE = regexp.MustCompile(`nextcloud-(\d+\.\d+\.\d+)\.zip`)

// listNextcloudVersions scrapes the HTML directory listing at download.nextcloud.com/server/releases/ since there's no release-assets JSON API like Joomla/OpenCart's GitHub releases - returns versions sorted newest first
func listNextcloudVersions(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://download.nextcloud.com/server/releases/", nil)
	if err != nil {
		return nil, err
	}
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

	matches := nextcloudArchiveNameRE.FindAllStringSubmatch(string(body), -1)
	seen := map[string]bool{}
	versions := make([]string, 0, len(matches))
	for _, m := range matches {
		if !seen[m[1]] {
			seen[m[1]] = true
			versions = append(versions, m[1])
		}
	}
	if len(versions) == 0 {
		return nil, errors.New("no Nextcloud versions found")
	}
	sort.Slice(versions, func(i, j int) bool { return cmsapp.CompareVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// latestNextcloudVersion returns the highest available version, used server-side when the install form's version field is left blank
func latestNextcloudVersion(ctx context.Context) (string, error) {
	versions, err := listNextcloudVersions(ctx)
	if err != nil {
		return "", err
	}
	return versions[0], nil
}
