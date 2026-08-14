package opencart

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// latestOpenCartVersion asks GitHub's releases API for the latest stable
// OpenCart tag (e.g. "4.1.0.4") - used server-side only when the install
// form's version field is left blank ("Latest"), since the install form
// itself populates specific versions client-side straight from the same
// releases list (see opencart_install.html's version-fetch script).
func latestOpenCartVersion(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/repos/opencart/opencart/releases/latest", nil)
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
