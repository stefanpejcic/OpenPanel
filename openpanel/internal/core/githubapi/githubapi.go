// Package githubapi does the unauthenticated GitHub API GETs the CMS modules use to look up versions, and the token calls the Git module makes for connected accounts.
package githubapi

import (
	"bytes"
	"context"
	"encoding/json"
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

// APIBase is swapped out in tests
var APIBase = "https://api.github.com"

var (
	// ErrUnauthorized means GitHub refused the token, it is wrong, expired or revoked
	ErrUnauthorized = errors.New("GitHub did not accept the access token")
	// ErrNotFound also covers things the token is not allowed to see, GitHub answers 404 for both
	ErrNotFound = errors.New("not found on GitHub, or the access token has no permission for it")
)

// Response is what an authenticated call hands back
type Response struct {
	Body   []byte
	Header http.Header
}

// maxBody bounds what a single API answer can take in memory
const maxBody = 8 << 20

// Do makes an authenticated API call, path starts with a slash and payload is sent as JSON when not nil
func Do(ctx context.Context, method, path, token string, payload any, timeout time.Duration) (*Response, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, APIBase+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+token)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := (&http.Client{Timeout: timeout}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, err
	}

	switch {
	case resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0"):
		rl := &RateLimitError{}
		if secs, parseErr := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); parseErr == nil {
			rl.Reset = time.Unix(secs, 0)
		}
		return nil, rl
	case resp.StatusCode == http.StatusUnauthorized:
		return nil, ErrUnauthorized
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case resp.StatusCode < 200 || resp.StatusCode > 299:
		return nil, fmt.Errorf("GitHub API returned %s%s", resp.Status, apiMessage(data))
	}
	return &Response{Body: data, Header: resp.Header}, nil
}

// apiMessage pulls GitHub's own explanation out of an error body
func apiMessage(body []byte) string {
	var e struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &e) != nil || e.Message == "" {
		return ""
	}
	if len(e.Message) > 200 {
		e.Message = e.Message[:200]
	}
	return ": " + e.Message
}

// TokenExpiry reads when the token used for a call runs out, zero when it never does or GitHub didn't say
func TokenExpiry(h http.Header) time.Time {
	v := h.Get("github-authentication-token-expiration")
	for _, layout := range []string{"2006-01-02 15:04:05 MST", "2006-01-02 15:04:05 -0700", time.RFC3339} {
		if t, err := time.Parse(layout, v); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
