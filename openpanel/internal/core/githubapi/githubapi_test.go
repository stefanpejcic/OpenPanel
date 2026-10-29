package githubapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func serve(t *testing.T, status int, headers map[string]string, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestGetReturnsBodyOn200(t *testing.T) {
	url := serve(t, 200, nil, `[{"name":"v1.2.3"}]`)
	body, err := Get(context.Background(), url, time.Second)
	if err != nil || string(body) != `[{"name":"v1.2.3"}]` {
		t.Fatalf("got %q, %v", body, err)
	}
}

func TestGetRateLimited403(t *testing.T) {
	url := serve(t, 403, map[string]string{"X-RateLimit-Remaining": "0", "X-RateLimit-Reset": "1793276876"}, `{"message":"API rate limit exceeded"}`)
	_, err := Get(context.Background(), url, time.Second)
	if !errors.Is(err, ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
	if !strings.Contains(err.Error(), "resets at 12:27 UTC") {
		t.Fatalf("expected reset time in %q", err.Error())
	}
	if PickVersionHint(err) == "" {
		t.Fatal("expected a pick-a-version hint for a rate limit")
	}
}

func TestGetRateLimited429WithoutReset(t *testing.T) {
	url := serve(t, 429, nil, `{}`)
	_, err := Get(context.Background(), url, time.Second)
	if !errors.Is(err, ErrRateLimited) || strings.Contains(err.Error(), "resets") {
		t.Fatalf("got %v", err)
	}
}

func TestGetOther403IsNotRateLimit(t *testing.T) {
	url := serve(t, 403, map[string]string{"X-RateLimit-Remaining": "12"}, `{"message":"Forbidden"}`)
	_, err := Get(context.Background(), url, time.Second)
	if err == nil || errors.Is(err, ErrRateLimited) || PickVersionHint(err) != "" {
		t.Fatalf("got %v", err)
	}
}

func TestGetNotFound(t *testing.T) {
	url := serve(t, 404, nil, `{"message":"Not Found"}`)
	_, err := Get(context.Background(), url, time.Second)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("got %v", err)
	}
}
