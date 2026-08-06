package emails

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/sieveparser"
)

var (
	errInvalidEmailFormat = errors.New("Invalid email format")
	errPathTraversal      = errors.New("Path traversal detected")
)

// resolveSievePath resolves an email address to its Sieve filter path,
// validating the address and confirming the user owns the domain.
// Domain-ownership failure is reported separately from other errors so the
// caller can return 403 instead of a generic 400.
func resolveSievePath(email string, userDomains map[string]bool) (path string, forbidden bool, err error) {
	if !isValidEmail(email) {
		return "", false, errInvalidEmailFormat
	}
	user, domain, _ := strings.Cut(email, "@")
	if !userDomains[domain] {
		return "", true, nil
	}

	candidate := filepath.Join(baseMailPath, domain, user, "home", ".dovecot.sieve")
	base := filepath.Clean(baseMailPath)
	if candidate != base && !strings.HasPrefix(candidate, base+string(filepath.Separator)) {
		return "", false, errPathTraversal
	}
	return candidate, false, nil
}

// writeSieve writes the Sieve filter content to disk, creating the parent
// directory first if it doesn't already exist.
func writeSieve(resolvedPath, content string) error {
	dir := filepath.Dir(resolvedPath)
	if _, err := os.Stat(dir); err != nil {
		if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
			return mkErr
		}
	}
	return os.WriteFile(resolvedPath, []byte(content), 0o644)
}

// handleFiltersForUser renders the list of mail filters for the current user.
func handleFiltersForUser(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)
	currentEmailsList := GetEmailList(ctx, a, userID, currentUsername, userDomains)

	renderFilterPage(a, w, r, "", currentEmailsList, nil, "", "", "gui")
}

// handleFilterForEmail redirects to the GUI filter editor for an email address.
func handleFilterForEmail(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")
	http.Redirect(w, r, "/emails/filter/"+email+"/gui", http.StatusFound)
}

// handleFilterGUI renders the visual (drag-and-drop) Sieve filter editor.
func handleFilterGUI(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	handleFilterView(a, w, r, "gui")
}

// handleFilterRaw renders the raw-text Sieve filter editor.
func handleFilterRaw(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	handleFilterView(a, w, r, "raw")
}

func handleFilterView(a *appctx.App, w http.ResponseWriter, r *http.Request, viewMode string) {
	ctx := r.Context()
	email := r.PathValue("email")
	userID, _ := auth.UserID(r)
	currentUsername, _, err := injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	domains, _ := a.AllDomainsForUser(ctx, userID)
	userDomains := domainSet(domains)

	resolvedPath, forbidden, resolveErr := resolveSievePath(email, userDomains)
	if forbidden {
		http.Error(w, "You do not own this domain.", http.StatusForbidden)
		return
	}
	if resolveErr != nil {
		http.Error(w, resolveErr.Error(), http.StatusBadRequest)
		return
	}

	rawContent := ""
	if data, readErr := os.ReadFile(resolvedPath); readErr == nil {
		rawContent = string(data)
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		newContent, ok := r.Form["new_content"]
		if !ok {
			flashAndRedirect(a, w, r, "error", "No content provided.", "/emails/filter/"+email+"/"+viewMode)
			return
		}
		if writeErr := writeSieve(resolvedPath, newContent[0]); writeErr != nil {
			flashAndRedirect(a, w, r, "error", "Error saving filter.", "/emails/filter/"+email+"/"+viewMode)
			return
		}
		action := "GUI"
		if viewMode == "raw" {
			action = "raw"
		}
		ipAddress := reqip.ClientIP(r)
		_ = logger.RecordUserAction(a.Config, currentUsername, "edited filter ("+action+") for "+email, ipAddress)
		flashAndRedirect(a, w, r, "success", "Filter saved successfully.", "/emails/filter/"+email+"/"+viewMode)
		return
	}

	parsedFilters := sieveparser.Parse(rawContent)
	renderFilterPage(a, w, r, email, nil, parsedFilters, rawContent, resolvedPath, viewMode)
}
