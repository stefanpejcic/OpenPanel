// Package session configures the panel's cookie store: cookie name and
// max-age/lifetime settings. Session contents are not compatible with the
// previous panel version's signed cookies by design — cutover forces a
// one-time re-login, which was an accepted tradeoff.
package session

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
)

// CookieName is the session cookie name, kept the same as the previous
// panel version so reverse proxies / browser devtools during the migration
// show a familiar name.
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
