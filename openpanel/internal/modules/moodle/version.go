package moodle

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/githubapi"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
)

var moodleVersionRE = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

type githubTag struct {
	Name string `json:"name"`
}

// listMoodleVersions hits the GitHub tags API and returns every stable "vX.Y.Z" tag (prerelease tags carry a "-rcN"/"-beta" suffix and are excluded by the exact-three-part regex), newest first - the tarball itself is GitHub's tag archive too (see install.go's moodleArchiveURL)
func listMoodleVersions(ctx context.Context) ([]string, error) {
	body, err := githubapi.Get(ctx, "https://api.github.com/repos/moodle/moodle/tags?per_page=100", 15*time.Second)
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
	sort.Slice(versions, func(i, j int) bool { return cmsapp.CompareVersions(versions[i], versions[j]) > 0 })
	return versions, nil
}

// latestMoodleVersion returns the highest available version - used server-side when the install form's version field is left blank
func latestMoodleVersion(ctx context.Context) (string, error) {
	versions, err := listMoodleVersions(ctx)
	if err != nil {
		return "", err
	}
	return versions[0], nil
}
