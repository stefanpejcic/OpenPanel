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

// regression test: gorilla/csrf defaults to field name "gorilla.csrf.Token" but our templates render "csrf_token", so every POST got rejected without csrf.FieldName("csrf_token") set - simulates a real browser via GET /login then POST back
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
	// use the real exemptAPIFromCSRF wiring, not a bare csrf.Protect(mux), so the plain-http handling main.go does is covered too
	protected := exemptAPIFromCSRF(csrfMiddleware, mux)

	// GET /login gets us the CSRF cookie and the token embedded in the form
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
	// browsers auto-decode escaped attribute values (e.g. "+" -> "&#43;"), so decode it here too instead of sending the raw escaped text
	token := html.UnescapeString(m[1])

	// password is empty on purpose - handleLoginPassword rejects that before touching the database, so this only exercises the CSRF gate, not a real login
	form := url.Values{"csrf_token": {token}, "username": {"someone"}, "password": {""}}
	postReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(cookie)
	postW := httptest.NewRecorder()
	protected.ServeHTTP(postW, postReq)

	if postW.Code == http.StatusForbidden && strings.Contains(postW.Body.String(), "CSRF") {
		t.Fatalf("POST /login was rejected as a CSRF failure despite a valid token/cookie pair; body: %s", postW.Body.String())
	}
	// wrong creds still render 200 with an error - we just care it's not a CSRF rejection
	if postW.Code != http.StatusOK {
		t.Errorf("POST /login status = %d, want 200 (login form re-render with an error)", postW.Code)
	}
}
