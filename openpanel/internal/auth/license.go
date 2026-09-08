package auth

import (
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// licenseExemptPaths stay reachable even when the license check fails: login and the forgot-password flow (route is /reset_password, not /forgot_password)
var licenseExemptPaths = []string{"/login", "/reset_password"}

// CheckLicense gates every request behind a valid license, calling render instead of the wrapped handler on failure - render is injected so this package doesn't need to depend on the template engine
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
