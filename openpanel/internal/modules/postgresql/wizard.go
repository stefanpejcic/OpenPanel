package postgresql

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
)

// handleDatabasesWizard renders the setup wizard form. Unlike MySQL's
// wizard, all the actual create/assign work happens client-side via three
// sequential fetch() calls to /postgresql/new, /postgresql/user and
// /postgresql/assign (see wizard.html) - this handler only gates access on
// the container being up and renders the form.
func handleDatabasesWizard(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, userContext, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	status := docker.GetContainerStatus(ctx, userContext, "postgres")
	if status.State != "running" {
		flashAndRedirect(a, w, r, "warning", "Postgres service is not ready yet. Please wait for the installation to finish before creating a database.", "/postgresql")
		return
	}

	if !checkPostgresInsideContainer(ctx, userContext) {
		http.Redirect(w, r, "/postgresql", http.StatusFound)
		return
	}

	renderWizardPage(a, w, r)
}
