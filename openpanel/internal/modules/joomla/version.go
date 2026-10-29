package joomla

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/core/githubapi"
)

// latestJoomlaVersion asks GitHub's releases API for the latest stable Joomla tag (e.g. "6.1.2") - used server-side only when the install form's version field is left blank ("Latest"), since the form itself populates specific versions client-side straight from the same releases list (see drupal_install.html's analogous version-fetch script)
func latestJoomlaVersion(ctx context.Context) (string, error) {
	body, err := githubapi.Get(ctx, "https://api.github.com/repos/joomla/joomla-cms/releases/latest", 10*time.Second)
	if err != nil {
		return "", err
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if decodeErr := json.Unmarshal(body, &release); decodeErr != nil {
		return "", decodeErr
	}
	if release.TagName == "" {
		return "", errors.New("GitHub returned no Joomla release")
	}
	return release.TagName, nil
}
