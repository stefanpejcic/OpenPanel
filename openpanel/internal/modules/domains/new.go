package domains

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

var domainCharsRE = regexp.MustCompile(`^[a-zA-Z0-9.-]+$`)

func flashAndRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
	http.Redirect(w, r, path, http.StatusFound)
}

// flashSess adds a flash message without redirecting - for the GET
// branches that flash an error but still render the current page.
func flashSess(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

// resolveUnderVarWWWHTML resolves a path lexically (no symlink following -
// the path may not exist yet) and confirms it stays under the docroot base,
// used for onion key paths and docroot in handleDomainsNew.
func resolveUnderVarWWWHTML(raw string) (resolved string, ok bool) {
	const base = "/var/www/html/"
	abs := raw
	if !filepath.IsAbs(abs) {
		abs = filepath.Join("/", abs)
	}
	cleaned := filepath.Clean(abs)
	if cleaned != strings.TrimSuffix(base, "/") && !strings.HasPrefix(cleaned+"/", base) {
		return cleaned, false
	}
	return cleaned, true
}

// handleDomainsNew adds a new domain for this account.
func handleDomainsNew(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		renderNewDomainPage(a, w, r)
		return
	}

	ctx := r.Context()
	userID, _ := auth.UserID(r)
	currentUsername, userContext, err := injected(a, ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = r.ParseForm()
	domainURL := r.Form.Get("domain_url")
	if domainURL == "" {
		flashAndRedirect(a, w, r, "error", "Domain name is required", "/domains/new")
		return
	}

	domainName := r.Form.Get("domain_name")
	if domainName == "" {
		domainName = domainURL
	}
	skipContainers := r.Form.Get("skip_containers")
	providedDocroot := r.Form.Get("docroot")
	if providedDocroot == "" {
		providedDocroot = "/var/www/html/"
	}

	const userHomeDirectory = "/var/www/html/"
	homePrefix := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data/"

	domainURL = strings.TrimSuffix(domainURL, ".")
	domainURL = strings.ToLower(domainURL)

	if !domainCharsRE.MatchString(domainURL) {
		flashAndRedirect(a, w, r, "error", "Domain name contains invalid characters.", "/domains/new")
		return
	}

	var onionFlags []string
	if strings.HasSuffix(domainURL, ".onion") {
		onionPublicKey := r.Form.Get("hs_ed25519_public_key")
		onionSecretKey := r.Form.Get("hs_ed25519_secret_key")

		if onionPublicKey == "" || onionSecretKey == "" {
			flashAndRedirect(a, w, r, "error", "Paths for both public and secret files must be set for importing .onion domain.", "/domains/new")
			return
		}

		normalizedPublic, publicOK := resolveUnderVarWWWHTML(onionPublicKey)
		if !publicOK {
			flashAndRedirect(a, w, r, "error", "Path for public file must start with /var/www/html/", "/domains/new")
			return
		}
		fullPublicKeyPath := filepath.Join(homePrefix, strings.TrimPrefix(normalizedPublic, userHomeDirectory))
		if info, statErr := os.Stat(fullPublicKeyPath); statErr != nil || info.IsDir() {
			flashAndRedirect(a, w, r, "error", "Public key file: "+onionPublicKey+" does not exist!", "/domains/new")
			return
		}

		normalizedSecret, secretOK := resolveUnderVarWWWHTML(onionSecretKey)
		if !secretOK {
			flashAndRedirect(a, w, r, "error", "Path for secret file must start with /var/www/html/", "/domains/new")
			return
		}
		fullSecretKeyPath := filepath.Join(homePrefix, strings.TrimPrefix(normalizedSecret, userHomeDirectory))
		if info, statErr := os.Stat(fullSecretKeyPath); statErr != nil || info.IsDir() {
			flashAndRedirect(a, w, r, "error", "Secret key file: "+onionSecretKey+" does not exist!", "/domains/new")
			return
		}

		onionFlags = []string{"--hs_ed25519_public_key", onionPublicKey, "--hs_ed25519_secret_key", onionSecretKey}
	}

	docroot, docrootOK := resolveUnderVarWWWHTML(providedDocroot)
	if !docrootOK {
		docroot = userHomeDirectory
	}

	// LIMITS
	injectedData, _ := a.InjectData(ctx, userID)
	planID, _ := injectedData["hosting_plan"].(int)
	domainsLimit := 0
	if plan, planErr := a.QueryPlanDetailsByID(ctx, planID); planErr == nil {
		domainsLimit = atoiDefault(plan.DomainsLimit, 0)
	}

	if domainsLimit != 0 {
		existing, _ := a.AllDomainsForUser(ctx, userID)
		urls := make([]appctx.Domain, len(existing))
		for i, d := range existing {
			urls[i] = appctx.Domain{DomainURL: d.DomainURL}
		}
		mains, _ := appctx.Categorize(urls)
		if len(mains) >= domainsLimit {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "Error: Reached the maximum allowed number of domains on this hosting plan.")
			return
		}
	}

	ipAddress := reqip.ClientIP(r)

	args := []string{"domains-add", domainURL, currentUsername, "--docroot", docroot}
	if skipContainers != "" {
		args = append(args, skipContainers)
	}
	args = append(args, onionFlags...)
	args = append(args, "--debug")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, canFlush := w.(http.Flusher)

	cmd := exec.Command("opencli", args...)

	stdout, pipeErr := cmd.StdoutPipe()
	if pipeErr != nil {
		fmt.Fprintf(w, "Error: %s\n\n", pipeErr.Error())
		return
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(w, "Error: %s\n\n", err.Error())
		return
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		fmt.Fprintf(w, "%s\n", scanner.Text())
		if canFlush {
			flusher.Flush()
		}
	}

	if waitErr := cmd.Wait(); waitErr != nil {
		fmt.Fprintf(w, "Error adding domain %s\n\n", domainURL)
	} else {
		_ = logger.RecordUserAction(a.Config, currentUsername, "added domain "+domainURL, ipAddress)
	}
	if canFlush {
		flusher.Flush()
	}
}
