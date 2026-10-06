// https://discuss.flarum.org/d/39930-what-features-should-a-flarum-manager-have
package flarum

import (
	"context"
	"net/http"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/appkit"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/cmsapp"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// install, clone, remove and the toggle must all derive the same comment to find the same job
func flarumCronComment(selectedDomain string) string {
	return "flarum-" + appkit.SiteSlug(selectedDomain)
}

func schedulerCommand(installPath string) string {
	return "php " + installPath + "/flarum schedule:run"
}

func addScheduler(ctx context.Context, userContext, selectedDomain, phpContainer, installPath string) error {
	return crons.AddJob(ctx, userContext, flarumCronComment(selectedDomain), "0 * * * * *", phpContainer, schedulerCommand(installPath), true)
}

// cloneAddCron gives the clone its own scheduler job
func cloneAddCron(c *cmsapp.Clone) {
	webServer := webserver.GetEnvFileValue(c.UserContext, "WEB_SERVER")
	phpContainer := webServer
	if !strings.Contains(strings.ToLower(webServer), "litespeed") {
		phpContainer = "php-fpm-" + c.PHPVersion
	}
	_ = addScheduler(c.Ctx, c.UserContext, c.DstDomainWithSubdir, phpContainer, c.Docroot)
}

// handleFlarumScheduler reports the schedule:run cron job on GET and turns it on/off on POST with state=on|off
func handleFlarumScheduler(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	domain, docroot, phpContainer, ok := cmsapp.RequestParams(ctx, a, r, userID, userContext)
	if !ok {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "domain and docroot are required, or you do not own this domain"})
		return
	}
	comment := flarumCronComment(domain)

	if r.Method == http.MethodGet {
		web.WriteJSON(w, http.StatusOK, map[string]any{"enabled": crons.HasJob(userContext, comment)})
		return
	}

	switch r.URL.Query().Get("state") {
	case "on":
		// drops a copy the user disabled from Cron Jobs so we don't end up with two
		_ = crons.RemoveJobByComment(ctx, userContext, comment)
		if addErr := addScheduler(ctx, userContext, domain, phpContainer, docroot); addErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to add the scheduler cron job: " + addErr.Error()})
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "enabled Flarum scheduler for "+domain, reqip.ClientIP(r))
		web.WriteJSON(w, http.StatusOK, map[string]any{"enabled": true, "message": "Scheduler enabled."})
	case "off":
		if rmErr := crons.RemoveJobByComment(ctx, userContext, comment); rmErr != nil {
			web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to remove the scheduler cron job: " + rmErr.Error()})
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "disabled Flarum scheduler for "+domain, reqip.ClientIP(r))
		web.WriteJSON(w, http.StatusOK, map[string]any{"enabled": false, "message": "Scheduler disabled."})
	default:
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "state must be on or off"})
	}
}
