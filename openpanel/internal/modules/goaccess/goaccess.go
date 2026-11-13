// Package goaccess serves the pre-rendered GoAccess HTML report per domain, generated externally (opencli/cron, once every 24h per the UI copy) and simply read from disk here.
package goaccess

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// handleDomainStats shows the domain picker when no domain_name is given, or reads and serves that domain's pre-rendered GoAccess HTML report file as-is
func handleDomainStats(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	domainName := r.PathValue("domain_name")
	if domainName == "" {
		domainsList, _ := a.AllDomainsForUser(ctx, userID)
		_, currentUsername, _, _ := auth.Injected(a, r)
		picks := make([]web.DomainPick, 0, len(domainsList))
		for _, d := range domainsList {
			p := web.DomainPick{Domain: d.DomainURL}
			if info, statErr := tenantfs.Stat("/var/log/caddy/stats/" + currentUsername + "/" + d.DomainURL + ".html"); statErr == nil {
				p.Detail = info.ModTime().Format("2006-01-02 15:04")
			} else {
				p.Disabled, p.Note = true, "No report yet, they're made every 24 hours"
			}
			picks = append(picks, p)
		}
		selected, summary := pickSummary(a, r, domainsList)
		renderGoaccessSelectPage(a, w, r, picks, selected, summary)
		return
	}

	if !a.CheckDomainBelongsToUser(ctx, userID, domainName) {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}

	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	logFilePath := "/var/log/caddy/stats/" + currentUsername + "/" + domainName + ".html"
	content, readErr := tenantfs.ReadFile(logFilePath)
	if readErr != nil {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Stats file for domain %(domain_name)s not found. Data is generated every 24h.", "domain_name", domainName), "/domains/stats")
		return
	}

	// goaccess_single.html has no {% extends %} - the report is a complete standalone HTML document, not wrapped in the panel's own layout, so this writes the file's bytes directly rather than going through a web.Page
	// the report is built by GoAccess from anonymous visitor data (User-Agent/Referer/URLs); serve it
	// sandboxed so that if GoAccess ever fails to escape something, the injected script can't run with
	// the viewer's panel session (no allow-same-origin).
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "sandbox allow-scripts; default-src 'none'; style-src 'unsafe-inline'; script-src 'unsafe-inline'; img-src data:")
	_, _ = w.Write(content)
}

// pickSummary reads ?domain= or, without one, the first domain that already has a report
func pickSummary(a *appctx.App, r *http.Request, domains []appctx.Domain) (string, *StatsSummary) {
	if len(domains) == 0 {
		return "", nil
	}
	_, currentUsername, _, err := auth.Injected(a, r)
	if err != nil {
		return "", nil
	}
	read := func(d string) *StatsSummary {
		content, err := tenantfs.ReadFile("/var/log/caddy/stats/" + currentUsername + "/" + d + ".html")
		if err != nil {
			return nil
		}
		s, err := parseReport(content)
		if err != nil {
			return nil
		}
		return s
	}
	if want := r.URL.Query().Get("domain"); want != "" {
		for _, d := range domains {
			if d.DomainURL == want {
				return want, read(want)
			}
		}
	}
	for _, d := range domains {
		if s := read(d.DomainURL); s != nil {
			return d.DomainURL, s
		}
	}
	return domains[0].DomainURL, nil
}
