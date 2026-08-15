package nextcloud

import (
	"context"
	"errors"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var nextcloudArchiveNameRE = regexp.MustCompile(`nextcloud-(\d+\.\d+\.\d+)\.zip`)

// listNextcloudVersions scrapes the HTML directory listing at
// download.nextcloud.com/server/releases/ for available nextcloud-X.Y.Z.zip
// archives - Nextcloud has no release-assets JSON API (unlike
// Joomla/OpenCart's GitHub releases), so this is the only source (also
// used by nextcloud_install.html's version dropdown, via a small internal
// API endpoint backed by this function). Returns versions sorted newest
// first.
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
	sort.Slice(versions, func(i, j int) bool { return compareVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// latestNextcloudVersion returns the highest available version - used
// server-side when the install form's version field is left blank.
func latestNextcloudVersion(ctx context.Context) (string, error) {
	versions, err := listNextcloudVersions(ctx)
	if err != nil {
		return "", err
	}
	return versions[0], nil
}

// compareVersions compares two dotted numeric versions ("34.0.3" vs
// "9.2.1"); returns >0 if a > b.
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
