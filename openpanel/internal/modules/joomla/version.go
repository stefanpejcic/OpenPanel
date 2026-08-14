package joomla

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// latestJoomlaVersion asks GitHub's releases API for the latest stable
// Joomla tag (e.g. "6.1.2") - used server-side only when the install form's
// version field is left blank ("Latest"), since the install form itself
// populates specific versions client-side straight from the same releases
// list (see drupal_install.html's analogous version-fetch script).
func latestJoomlaVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/joomla/joomla-cms/releases/latest", nil)
	if err != nil {
		return "", err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var release struct {
		TagName string `json:"tag_name"`
	}
	if decodeErr := json.NewDecoder(resp.Body).Decode(&release); decodeErr != nil {
		return "", decodeErr
	}
	return release.TagName, nil
}
