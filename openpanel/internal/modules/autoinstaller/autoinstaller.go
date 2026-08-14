// Package autoinstaller implements the "Auto Installer" hub page listing
// every one-click app type (WordPress, Drupal, Joomla, Website Builder,
// Mautic, NodeJS, Python), each showing how many instances of that type
// this user already has installed.
package autoinstaller

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
)

// technologies is every site type counted for the autoinstaller hub. Only
// wordpress/drupal/sitebuilder (website_builder)/mautic/node/python have a
// card in the template - the rest are counted but never displayed.
var technologies = []string{
	"wordpress", "drupal", "joomla", "opencart", "sitebuilder", "node", "python", "php",
	"java", "ruby", "bun", "mautic", "flarum", "fossbilling",
}

// getAutoinstallerData returns every domain the user owns, plus a
// per-technology count of sites whose type contains that technology's name
// (case-insensitively).
func getAutoinstallerData(ctx context.Context, a *appctx.App, userID int) ([]appctx.Domain, map[string]int, error) {
	domains, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	rows, err := a.DB.QueryContext(ctx,
		"SELECT type FROM sites WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?)", userID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	counts := make(map[string]int, len(technologies))
	for _, t := range technologies {
		counts[t] = 0
	}
	for rows.Next() {
		var siteType string
		if scanErr := rows.Scan(&siteType); scanErr != nil {
			return nil, nil, scanErr
		}
		siteTypeLower := strings.ToLower(siteType)
		for _, t := range technologies {
			if strings.Contains(siteTypeLower, t) {
				counts[t]++
			}
		}
	}
	return domains, counts, rows.Err()
}

// handleAutoinstaller renders the autoinstaller hub page. On a database
// error, the response body is just the bare error text with no error
// wrapping and no non-200 status - preserved here exactly since it's an
// already-visible production behavior, not something to improve on
// incidentally.
func handleAutoinstaller(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	data, err := cache.Memoize(ctx, a.Cache, autoinstallerCacheKey(userID), 30*time.Second, func() (autoinstallerData, error) {
		d, c, getErr := getAutoinstallerData(ctx, a, userID)
		if getErr != nil {
			return autoinstallerData{}, getErr
		}
		return autoinstallerData{Domains: d, Counts: c}, nil
	})
	if err != nil {
		_, _ = w.Write([]byte(" " + err.Error()))
		return
	}

	renderAutoinstallerPage(a, w, r, data.Domains, data.Counts)
}

type autoinstallerData struct {
	Domains []appctx.Domain
	Counts  map[string]int
}

func autoinstallerCacheKey(userID int) string {
	return "autoinstaller:" + strconv.Itoa(userID)
}

// Register wires the /auto-installer route onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "autoinstaller")(h)
	}
	mux.Handle("GET /auto-installer", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAutoinstaller(a, w, r) }))
}
