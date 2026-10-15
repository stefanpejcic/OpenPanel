package web

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// CaddyDomainsDir is a var so tests can point it at a temp dir
var CaddyDomainsDir = "/etc/openpanel/caddy/domains"

// DomainSuspension is what opencli domains-suspend writes as header lines above the caddy block
type DomainSuspension struct {
	By      string // "admin", "user", or "" for older suspensions
	Reason  string
	Comment string
}

// ReadDomainSuspension reads the vhost directly, no cache, since it gates writes
func ReadDomainSuspension(domain string) DomainSuspension {
	var s DomainSuspension
	if domain == "" || strings.ContainsAny(domain, `/\`) || strings.HasPrefix(domain, ".") {
		return s
	}
	f, err := os.Open(filepath.Join(CaddyDomainsDir, domain+".conf"))
	if err != nil {
		return s
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, "# ") {
			break
		}
		if v, ok := strings.CutPrefix(line, "# suspended_by: "); ok {
			s.By = v
		} else if v, ok := strings.CutPrefix(line, "# reason: "); ok {
			s.Reason = v
		} else if v, ok := strings.CutPrefix(line, "# comment: "); ok {
			s.Comment = v
		}
	}
	return s
}

func IsAdminSuspended(domain string) bool {
	return ReadDomainSuspension(strings.ToLower(strings.TrimSpace(domain))).By == "admin"
}

// requestDomains collects every domain a request could be acting on, path wildcards and the form field names the handlers use
func requestDomains(r *http.Request) []string {
	var out []string
	for _, k := range []string{"domain", "domain_name", "domain_url"} {
		if v := r.PathValue(k); v != "" {
			out = append(out, v)
		}
	}
	// API catch-alls look like {rest...} = example.com/dns/records
	if rest := r.PathValue("rest"); rest != "" {
		out = append(out, strings.SplitN(rest, "/", 2)[0])
	}
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") || strings.HasPrefix(ct, "multipart/form-data") {
		for _, k := range []string{"domain", "Domain", "domain_name", "domain_url"} {
			if v := r.FormValue(k); v != "" {
				out = append(out, v)
			}
		}
	} else if ct == "" {
		out = append(out, r.URL.Query().Get("domain"), r.URL.Query().Get("domain_name"))
	}
	return out
}

// BlockAdminSuspended refuses any write to a domain the admin suspended, reads still go through
func BlockAdminSuspended(a *appctx.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet || r.Method == http.MethodHead {
				next.ServeHTTP(w, r)
				return
			}
			for _, d := range requestDomains(r) {
				if !IsAdminSuspended(d) {
					continue
				}
				msg := "Domain " + d + " was suspended by the administrator and can't be changed. Please contact support."
				if strings.HasPrefix(r.URL.Path, "/api/") || strings.Contains(r.Header.Get("Accept"), "application/json") || r.Header.Get("X-Requested-With") != "" {
					WriteJSON(w, http.StatusForbidden, map[string]string{"error": msg})
				} else {
					FlashRedirect(a, w, r, "error", msg, "/domains")
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
