// Package session configures the panel's cookie store: cookie name and
// max-age/lifetime settings.
package session

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// CookieName is the session cookie name
const CookieName = "OPENPANEL"

// Lifetime is the session max age: 300 minutes.
const Lifetime = 300 * time.Minute

// NewStore builds a cookie store using secretKey for signing/encryption.
func NewStore(secretKey []byte) *sessions.CookieStore {
	store := sessions.NewCookieStore(secretKey)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(Lifetime.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	return store
}
