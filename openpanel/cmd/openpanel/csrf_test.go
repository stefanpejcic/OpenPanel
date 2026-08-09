package main

import (
	"crypto/sha256"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gorilla/csrf"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/modules"
)

// TestCSRFFieldNameMatchesTemplates is a regression test for a real bug
// caught in manual testing: gorilla/csrf's default form field name is
// "gorilla.csrf.Token", but every template renders the hidden input as
// name="csrf_token". Without csrf.FieldName("csrf_token") in the
// middleware setup, every POST request in the app - including login - is
// rejected with "CSRF token not found in request", even with a
// perfectly valid token in the form.
//
// This exercises the exact chain that broke: GET /login (issues a token +
// cookie) -> extract the token from the rendered HTML, exactly as a
// browser would -> POST /login with that token in the "csrf_token" field
// and the cookie attached -> must NOT be rejected for CSRF.
func TestCSRFFieldNameMatchesTemplates(t *testing.T) {
	c := cache.New(filepath.Join(t.TempDir(), "no-redis.sock"))
	a := &appctx.App{
		Sessions:       session.NewStore([]byte("test-secret-key-for-csrf-test")),
		Cache:          c,
		I18n:           i18n.NewManager(t.TempDir(), c),
		EnabledModules: []string{"dashboard", "websites"},
		SecretKey:      []byte("test-secret-key-for-csrf-test"),
	}

	mux := http.NewServeMux()
	modules.RegisterAll(mux, a)

	csrfKey := sha256.Sum256(a.SecretKey)
	csrfMiddleware := csrf.Protect(csrfKey[:],
		csrf.Path("/"),
		csrf.CookieName("OPENPANEL_CSRF"),
		csrf.Secure(false),
		csrf.FieldName("csrf_token"),
	)
	// Goes through the real exemptAPIFromCSRF wiring (not a bare
	// csrf.Protect(...)(mux)) so this also covers the requestScheme /
	// PlaintextHTTPRequest handling main.go does for it - httptest requests
	// carry no TLS/X-Forwarded-Proto, matching a real plain-http://ip:port
	// request, which is exactly what gorilla/csrf >=1.7.3 would otherwise
	// reject as a missing Referer.
	protected := exemptAPIFromCSRF(csrfMiddleware, mux)

	// GET /login: issues the CSRF cookie and renders the token into the form.
	getReq := httptest.NewRequest(http.MethodGet, "/login", nil)
	getW := httptest.NewRecorder()
	protected.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("GET /login status = %d, want 200", getW.Code)
	}

	var cookie *http.Cookie
	for _, c := range getW.Result().Cookies() {
		if c.Name == "OPENPANEL_CSRF" {
			cookie = c
		}
	}
	if cookie == nil {
		t.Fatal("expected an OPENPANEL_CSRF cookie in the GET /login response")
	}

	tokenRE := regexp.MustCompile(`name="csrf_token" value="([^"]+)"`)
	m := tokenRE.FindStringSubmatch(getW.Body.String())
	if m == nil {
		t.Fatal("could not find csrf_token hidden field in rendered login page")
	}
	// html/template HTML-escapes the attribute value (e.g. "+" -> "&#43;"),
	// which a real browser decodes automatically when reading the DOM
	// attribute; replicate that decoding here rather than submitting the
	// raw escaped text.
	token := html.UnescapeString(m[1])

	// POST /login with that token, exactly as a browser form submission
	// would. Password is deliberately empty: handleLoginPassword rejects
	// that before touching the database (which this test doesn't set up),
	// so the only thing exercised past the CSRF gate is the "required"
	// validation message, not a live login attempt.
	form := url.Values{"csrf_token": {token}, "username": {"someone"}, "password": {""}}
	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(cookie)
	postW := httptest.NewRecorder()
	protected.ServeHTTP(postW, postReq)

	if postW.Code == http.StatusForbidden && strings.Contains(postW.Body.String(), "CSRF") {
		t.Fatalf("POST /login was rejected as a CSRF failure despite a valid token/cookie pair; body: %s", postW.Body.String())
	}
	// A wrong username/password still renders 200 with an error message -
	// the point of this test is only that it isn't a CSRF rejection.
	if postW.Code != http.StatusOK {
		t.Errorf("POST /login status = %d, want 200 (login form re-render with an error)", postW.Code)
	}
}
