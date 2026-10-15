// Package errorlogs shows PHP, webserver, database and HTTP 5xx errors for all of a user's sites on one page, so nobody has to know which container to dig through.
package errorlogs

import (
	"net/http"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const (
	defaultLimit = 500
	maxLimit     = 5000
)

// parseBound takes a YYYY-MM-DD day from the picker or an RFC3339 time from the API, endOfDay picks the last moment of a day
func parseBound(s string, loc *time.Location, endOfDay bool) (time.Time, string) {
	if d, err := time.ParseInLocation("2006-01-02", s, loc); err == nil {
		if endOfDay {
			return d.AddDate(0, 0, 1).Add(-time.Nanosecond), s
		}
		return d, s
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, ""
	}
	return time.Time{}, ""
}

func parseQuery(r *http.Request, now time.Time) Query {
	v := r.URL.Query()
	q := Query{
		Domain: strings.TrimSpace(v.Get("domain")), Service: v.Get("service"),
		Level: v.Get("level"), Search: strings.TrimSpace(v.Get("q")),
		Limit: web.AtoiDefault(v.Get("limit"), defaultLimit),
		// exact leaves out the sites in the domain's subfolders, the site manager wants just its own site
		Exact: v.Get("exact") == "1",
	}
	if q.Limit < 1 || q.Limit > maxLimit {
		q.Limit = defaultLimit
	}
	if q.Level != "warning" && q.Level != "error" {
		q.Level = ""
	}
	// tz is the browser's UTC offset in minutes, so a picked day starts at the viewer's midnight
	q.TZ = web.AtoiDefault(v.Get("tz"), 0)
	if q.TZ < -14*60 || q.TZ > 14*60 {
		q.TZ = 0
	}
	loc := time.FixedZone("", q.TZ*60)
	from, to := v.Get("from"), v.Get("to")
	if from != "" && to != "" && len(from) == len(to) && to < from {
		from, to = to, from
	}
	q.From, q.FromDay = parseBound(from, loc, false)
	q.To, q.ToDay = parseBound(to, loc, true)
	// a range ending today or later is open ended, so new errors keep showing up
	if !q.To.IsZero() && !q.To.Before(now) {
		q.To = time.Time{}
	}
	return q
}

// load resolves the user, checks the domain/service filters belong to them and runs the collection
func load(a *appctx.App, r *http.Request) (Query, Result, []Target, int, string) {
	ctx := r.Context()
	q := parseQuery(r, time.Now())

	userID, _, userContext, err := auth.Injected(a, r)
	if err != nil {
		return q, Result{}, nil, http.StatusInternalServerError, "internal error"
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	sites := loadSites(ctx, a, userID)
	targets := buildTargets(domains, sites)
	if q.Domain != "" {
		owned := false
		for _, t := range targets {
			owned = owned || t.Value == q.Domain
		}
		if !owned {
			return q, Result{}, targets, http.StatusForbidden, "You do not own this domain."
		}
	}

	res := Collect(ctx, a, userContext, domains, sites, q)
	if q.Service != "" {
		known := false
		for _, s := range res.Sources {
			known = known || s.Name == q.Service
		}
		if !known {
			return q, Result{Sources: res.Sources}, targets, http.StatusBadRequest, "Unknown service."
		}
	}
	return q, res, targets, http.StatusOK, ""
}

// exportLine is the plain-text form used for both the txt download and the copy button
func exportLine(e Entry) string {
	domain := e.Domain
	if domain == "" {
		domain = "-"
	}
	return e.Time.UTC().Format("2006-01-02 15:04:05") + " UTC [" + e.Level + "] [" + e.Service + "] [" + domain + "] " + e.Message
}

func handlePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	q, res, targets, status, errMsg := load(a, r)
	if status == http.StatusInternalServerError {
		http.Error(w, errMsg, status)
		return
	}

	// session twin of the API, the site manager's debugging tab reads it
	if r.URL.Query().Get("format") == "json" {
		writeJSONResult(w, q, res, status, errMsg)
		return
	}

	if r.URL.Query().Get("format") == "txt" {
		if errMsg != "" {
			http.Error(w, errMsg, status)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="error-logs-`+time.Now().UTC().Format("20060102-150405")+`.txt"`)
		var b strings.Builder
		for _, e := range res.Entries {
			b.WriteString(exportLine(e))
			b.WriteByte('\n')
		}
		_, _ = w.Write([]byte(b.String()))
		return
	}

	renderErrorLogsPage(a, w, r, q, res, targets, status, errMsg)
}

func handleAPI(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	q, res, _, status, errMsg := load(a, r)
	writeJSONResult(w, q, res, status, errMsg)
}

func writeJSONResult(w http.ResponseWriter, q Query, res Result, status int, errMsg string) {
	if errMsg != "" {
		web.WriteJSON(w, status, map[string]string{"error": errMsg})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{
		"from": q.From, "to": q.To, "total": res.Total, "truncated": res.Truncated,
		"sources": res.Sources, "entries": res.Entries,
	})
}
