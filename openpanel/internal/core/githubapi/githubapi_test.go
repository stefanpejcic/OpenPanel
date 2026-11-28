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

func TestDoMapsStatusesAndSendsTheToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/ok":
			w.Header().Set("github-authentication-token-expiration", "2030-05-06 07:08:09 UTC")
			_, _ = w.Write([]byte(`{"a":1}`))
		case "/limited":
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusForbidden)
		case "/denied":
			w.WriteHeader(http.StatusUnprocessableEntity)
			_, _ = w.Write([]byte(`{"message":"Hook already exists on this repository"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	old := APIBase
	APIBase = srv.URL
	defer func() { APIBase = old }()
	ctx := context.Background()

	resp, err := Do(ctx, http.MethodGet, "/ok", "tok", nil, time.Second)
	if err != nil || string(resp.Body) != `{"a":1}` {
		t.Fatalf("ok: %v %v", resp, err)
	}
	if exp := TokenExpiry(resp.Header); exp.Year() != 2030 || exp.Hour() != 7 {
		t.Fatalf("expiry = %v", exp)
	}
	if !TokenExpiry(http.Header{}).IsZero() {
		t.Fatal("no header should mean no expiry")
	}
	if _, err := Do(ctx, http.MethodGet, "/ok", "wrong", nil, time.Second); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("bad token: %v", err)
	}
	if _, err := Do(ctx, http.MethodGet, "/missing", "tok", nil, time.Second); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing: %v", err)
	}
	if _, err := Do(ctx, http.MethodGet, "/limited", "tok", nil, time.Second); !errors.Is(err, ErrRateLimited) {
		t.Fatalf("limited: %v", err)
	}
	if _, err := Do(ctx, http.MethodPost, "/denied", "tok", map[string]string{"x": "y"}, time.Second); err == nil || !strings.Contains(err.Error(), "Hook already exists") {
		t.Fatalf("GitHub's own message should be passed on, got %v", err)
	}
}
