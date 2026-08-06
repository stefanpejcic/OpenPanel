package auth

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// licenseExemptPaths lists the paths that stay reachable even when the
// license check fails: /login, and /reset_password (the "forgot password"
// flow - note the route path is /reset_password, not /forgot_password).
var licenseExemptPaths = []string{"/login", "/reset_password"}

// CheckLicense gates every request behind a valid license: when the check
// fails, render is called instead of the wrapped handler (typically
// rendering system/license_error.html). render is injected so this
// package doesn't need to depend on the template engine, which doesn't
// exist until a later step.
func CheckLicense(a *appctx.App, render http.HandlerFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isLicenseExempt(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			if !a.LicenseCheckPasses() {
				render(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isLicenseExempt(path string) bool {
	if strings.HasPrefix(path, "/static/") {
		return true
	}
	for _, p := range licenseExemptPaths {
		if path == p || strings.HasPrefix(path, p+"/") {
			return true
		}
	}
	return false
}
