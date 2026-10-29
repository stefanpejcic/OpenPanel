package opencart

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/githubapi"
)

// latestOpenCartVersion asks GitHub's releases API for the latest stable tag, used server-side only when the install form's version field is left blank
func latestOpenCartVersion(ctx context.Context) (string, error) {
	body, err := githubapi.Get(ctx, "https://api.github.com/repos/opencart/opencart/releases/latest", 10*time.Second)
	if err != nil {
		return "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if decodeErr := json.Unmarshal(body, &release); decodeErr != nil {
		return "", decodeErr
	}
	// an empty tag would build a download URL like .../download//opencart-.zip
	if release.TagName == "" {
		return "", errors.New("GitHub returned no OpenCart release")
	}
	return release.TagName, nil
}
