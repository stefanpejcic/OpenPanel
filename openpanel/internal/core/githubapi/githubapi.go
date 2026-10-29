// Package githubapi does the unauthenticated GitHub API GETs the CMS modules use to look up versions.
package githubapi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// ErrRateLimited matches any RateLimitError via errors.Is
var ErrRateLimited = errors.New("GitHub API rate limit reached")

// RateLimitError is returned once the server's IP used up GitHub's 60 unauthenticated requests per hour
type RateLimitError struct {
	Reset time.Time
}

func (e *RateLimitError) Error() string {
	if e.Reset.IsZero() {
		return "GitHub API rate limit reached for this server"
	}
	return "GitHub API rate limit reached for this server (resets at " + e.Reset.UTC().Format("15:04") + " UTC)"
}

func (e *RateLimitError) Is(target error) bool { return target == ErrRateLimited }

// PickVersionHint is appended to a failed "Latest" lookup, picking a specific version skips the API call
func PickVersionHint(err error) string {
	if errors.Is(err, ErrRateLimited) {
		return ", pick a specific version instead of Latest"
	}
	return ""
}

// Get returns the body of a GitHub API GET, any non-2xx is an error so GitHub's error object never gets decoded as data
func Get(ctx context.Context, url string, timeout time.Duration) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0") {
		rl := &RateLimitError{}
		if secs, parseErr := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); parseErr == nil {
			rl.Reset = time.Unix(secs, 0)
		}
		return nil, rl
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("GitHub API returned %s", resp.Status)
	}
	return body, nil
}
