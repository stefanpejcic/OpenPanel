// Package search implements the sidebar's entity search box, covering
// feature/page search (everyone) and, for Enterprise licenses, cross-entity
// search over databases/domains/emails/ftp/containers/services/websites/
// crons.
package search

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/searchdata"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/crons"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/dashboard"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/emails"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/mysql"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/postgresql"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/services"
	"gopkg.in/yaml.v3"
)

// enterpriseSearchTypes lists the search sub-types gated behind an
// Enterprise license.
var enterpriseSearchTypes = map[string]bool{
	"mysql_databases": true, "mysql_users": true,
	"postgresql_databases": true, "postgresql_users": true,
	"domains": true, "emails": true, "ftp": true, "containers": true,
	"services": true, "websites": true, "crons": true,
}

// gateFor returns the feature gate for a search sub-type: nil means no
// feature required, a non-empty slice means "at least one of these
// features".
func gateFor(what string) (required []string, ok bool) {
	gates := map[string][]string{
		"files": {"filemanager"}, "folders": {"filemanager"},
		"features":             nil,
		"websites":             {"wordpress", "website_builder", "nodejs", "python", "mautic", "flarum"},
		"mysql_databases":      {"mysql"},
		"mysql_users":          {"mysql"},
		"postgresql_databases": {"postgresql"},
		"postgresql_users":     {"postgresql"},
		"domains":              {"domains"},
		"emails":               {"emails"},
		"ftp":                  {"ftp"},
		"containers":           {"docker"},
		"services":             {"services"},
		"crons":                {"crons"},
	}
	required, ok = gates[what]
	return required, ok
}

func isEnterprise(a *appctx.App) bool {
	return strings.HasPrefix(a.LicenseKey, "enterprise")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// item is the {name, link} shape almost every sub-handler returns.
type item struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// HandleSearch dispatches a /json/search/{what} request to its sub-type
// handler after checking the Enterprise and feature gates.
func HandleSearch(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	what := r.PathValue("what")
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userAllowedSlice, _ := injected["user_allowed"].([]string)
	userAllowed := make(map[string]bool, len(userAllowedSlice))
	for _, m := range userAllowedSlice {
		userAllowed[m] = true
	}
	currentUsername, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)

	required, known := gateFor(what)
	if !known {
		writeJSONError(w, http.StatusNotFound, "Unknown search type")
		return
	}
	if enterpriseSearchTypes[what] && !isEnterprise(a) {
		writeJSONError(w, http.StatusForbidden, "Feature not available")
		return
	}
	if len(required) > 0 {
		allowed := false
		for _, m := range required {
			if userAllowed[m] {
				allowed = true
				break
			}
		}
		if !allowed {
			writeJSONError(w, http.StatusForbidden, "Feature not available")
			return
		}
	}

	switch what {
	case "files":
		searchFiles(a, w, r, userContext)
	case "folders":
		searchFolders(a, w, r, userContext)
	case "features":
		searchFeatures(w, userAllowed)
	case "websites":
		searchWebsites(a, w, r, userID)
	case "mysql_databases":
		searchMySQLDatabases(a, w, r, userContext)
	case "mysql_users":
		searchMySQLUsers(a, w, r, userContext)
	case "postgresql_databases":
		searchPostgresDatabases(a, w, r, userContext)
	case "postgresql_users":
		searchPostgresUsers(a, w, r, userContext)
	case "domains":
		searchDomains(a, w, r, userID)
	case "emails":
		searchEmails(a, w, r, userID, currentUsername)
	case "ftp":
		searchFTP(w, userContext)
	case "containers":
		searchContainers(a, w, r, userContext)
	case "services":
		searchServices(w, userContext)
	case "crons":
		searchCrons(w, userContext)
	}
}

func limitItems(items []item, n int) []item {
	if len(items) > n {
		return items[:n]
	}
	return items
}

// FEATURES
func searchFeatures(w http.ResponseWriter, userAllowed map[string]bool) {
	var allRoutes []map[string]string
	_ = json.Unmarshal(searchdata.FeaturesJSON, &allRoutes)

	filtered := make([]map[string]string, 0, len(allRoutes))
	for _, route := range allRoutes {
		if userAllowed[route["module"]] {
			filtered = append(filtered, route)
			if len(filtered) >= 100 {
				break
			}
		}
	}
	writeJSON(w, http.StatusOK, filtered)
}

// WEBSITES
func searchWebsites(a *appctx.App, w http.ResponseWriter, r *http.Request, userID int) {
	sites, err := dashboard.GetUserWebsites(a, r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	limited := sites
	if len(limited) > 10 {
		limited = limited[:10]
	}
	writeJSON(w, http.StatusOK, limited)
}

// MYSQL DATABASES / USERS
func searchMySQLDatabases(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext string) {
	data, err := mysql.ComputeDatabasesInfo(r.Context(), userContext)
	if err != nil {
		writeJSON(w, http.StatusOK, []item{})
		return
	}
	items := make([]item, 0, len(data.Databases))
	for _, name := range data.Databases {
		items = append(items, item{Name: name, Link: "/mysql"})
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

func searchMySQLUsers(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext string) {
	data, err := mysql.ComputeDatabasesInfo(r.Context(), userContext)
	if err != nil {
		writeJSON(w, http.StatusOK, []item{})
		return
	}
	items := make([]item, 0, len(data.Users))
	for _, name := range data.Users {
		items = append(items, item{Name: name, Link: "/mysql/users"})
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

// POSTGRESQL DATABASES / USERS
func searchPostgresDatabases(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext string) {
	databases, _, err := postgresql.ComputeDatabaseAndUserNames(r.Context(), userContext)
	if err != nil {
		writeJSON(w, http.StatusOK, []item{})
		return
	}
	items := make([]item, 0, len(databases))
	for _, name := range databases {
		items = append(items, item{Name: name, Link: "/postgresql"})
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

func searchPostgresUsers(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext string) {
	_, users, err := postgresql.ComputeDatabaseAndUserNames(r.Context(), userContext)
	if err != nil {
		writeJSON(w, http.StatusOK, []item{})
		return
	}
	items := make([]item, 0, len(users))
	for _, name := range users {
		items = append(items, item{Name: name, Link: "/postgresql/users"})
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

// DOMAINS
func searchDomains(a *appctx.App, w http.ResponseWriter, r *http.Request, userID int) {
	domains, err := a.AllDomainsForUser(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	items := make([]item, 0, len(domains))
	for _, d := range domains {
		items = append(items, item{Name: d.DomainURL, Link: "/domains"})
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

// EMAIL ACCOUNTS
func searchEmails(a *appctx.App, w http.ResponseWriter, r *http.Request, userID int, currentUsername string) {
	ctx := r.Context()
	domains, err := a.AllDomainsForUser(ctx, userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	userDomains := make(map[string]bool, len(domains))
	for _, d := range domains {
		userDomains[d.DomainURL] = true
	}
	raw := emails.GetEmailList(ctx, a, userID, currentUsername, userDomains)
	items := make([]item, 0, len(raw))
	for _, entry := range raw {
		parts := strings.Split(entry, " ")
		if len(parts) >= 2 {
			items = append(items, item{Name: parts[1], Link: "/emails"})
		}
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

// FTP ACCOUNTS
func searchFTP(w http.ResponseWriter, userContext string) {
	usersListFile := "/etc/openpanel/ftp/users/" + userContext + "/users.list"
	var items []item
	if content, err := os.ReadFile(usersListFile); err == nil {
		for _, line := range strings.Split(string(content), "\n") {
			parts := strings.Split(strings.TrimSpace(line), "|")
			if len(parts) >= 3 && parts[0] != "" {
				items = append(items, item{Name: parts[0], Link: "/ftp"})
			}
		}
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

// DOCKER CONTAINERS
func searchContainers(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext string) {
	names, err := docker.GetRunningContainers(r.Context(), userContext)
	if err != nil {
		writeJSON(w, http.StatusOK, []item{})
		return
	}
	items := make([]item, 0, len(names))
	for _, n := range names {
		if n == "" {
			continue
		}
		items = append(items, item{Name: n, Link: "/containers/edit/" + n})
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

// SERVICES
func searchServices(w http.ResponseWriter, userContext string) {
	composePath := "/home/" + userContext + "/docker-compose.yml"
	var names []string
	if data, err := os.ReadFile(composePath); err == nil {
		var composeData map[string]any
		if yaml.Unmarshal(data, &composeData) == nil {
			if svcMap, ok := composeData["services"].(map[string]any); ok {
				var all []string
				for name := range svcMap {
					all = append(all, name)
				}
				webserver, _ := docker.GetEnvValue(userContext, "WEB_SERVER")
				mysqlType, _ := docker.GetEnvValue(userContext, "MYSQL_TYPE")
				names = services.FilterServices(all, webserver, mysqlType)
			}
		}
	}
	items := make([]item, 0, len(names))
	for _, n := range names {
		items = append(items, item{Name: n, Link: "/services/" + n})
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

// CRON JOBS
func searchCrons(w http.ResponseWriter, userContext string) {
	cronFilePath := "/home/" + userContext + "/crons.ini"
	var items []item
	if content, err := os.ReadFile(cronFilePath); err == nil {
		for _, job := range crons.ParseCronFile(string(content)) {
			name := job.Comment
			if name == "" {
				name = job.Schedule
			}
			if name != "" {
				items = append(items, item{Name: name, Link: "/cronjobs"})
			}
		}
	}
	writeJSON(w, http.StatusOK, limitItems(items, 50))
}

// FILES
var maxSearchDepth = 5

func searchFiles(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext string) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	folder := strings.TrimSpace(r.URL.Query().Get("folder"))
	ext := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("ext")))

	baseDirectory := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data"
	userHomeDirectory, _ := filepath.Abs(baseDirectory)
	requestedDirectory, _ := filepath.Abs(filepath.Join(userHomeDirectory, folder))

	if !strings.HasPrefix(requestedDirectory, userHomeDirectory) {
		writeJSONError(w, http.StatusBadRequest, "Invalid file path")
		return
	}

	baseDepth := strings.Count(requestedDirectory, string(os.PathSeparator))
	var results []map[string]string

	_ = walkLimitedDepth(requestedDirectory, baseDepth, maxSearchDepth, func(root string, dirEntries []os.DirEntry) bool {
		for _, de := range dirEntries {
			if de.IsDir() {
				continue
			}
			nameLower := strings.ToLower(de.Name())
			if !strings.Contains(nameLower, q) {
				continue
			}
			if ext != "" && !strings.HasSuffix(nameLower, ext) {
				continue
			}
			rel, _ := filepath.Rel(userHomeDirectory, root)
			results = append(results, map[string]string{"name": de.Name(), "path": rel})
			if len(results) >= 10 {
				return false
			}
		}
		return len(results) < 10
	})

	if results == nil {
		results = []map[string]string{}
	}
	writeJSON(w, http.StatusOK, results)
}

// FOLDERS
func searchFolders(a *appctx.App, w http.ResponseWriter, r *http.Request, userContext string) {
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	folder := strings.TrimSpace(r.URL.Query().Get("folder"))

	baseDirectory := "/home/" + userContext + "/docker-data/volumes/" + userContext + "_html_data/_data"
	userHomeDirectory, _ := filepath.Abs(baseDirectory)
	requestedDirectory, _ := filepath.Abs(filepath.Join(userHomeDirectory, folder))

	if !strings.HasPrefix(requestedDirectory, userHomeDirectory) {
		writeJSONError(w, http.StatusBadRequest, "Invalid folder path")
		return
	}

	baseDepth := strings.Count(requestedDirectory, string(os.PathSeparator))
	var results []map[string]string

	_ = walkLimitedDepth(requestedDirectory, baseDepth, maxSearchDepth, func(root string, dirEntries []os.DirEntry) bool {
		for _, de := range dirEntries {
			if !de.IsDir() {
				continue
			}
			if !strings.Contains(strings.ToLower(de.Name()), q) {
				continue
			}
			rel, _ := filepath.Rel(userHomeDirectory, filepath.Join(root, de.Name()))
			results = append(results, map[string]string{"name": de.Name(), "path": rel})
			if len(results) >= 10 {
				return false
			}
		}
		return len(results) < 10
	})

	if results == nil {
		results = []map[string]string{}
	}
	writeJSON(w, http.StatusOK, results)
}

// walkLimitedDepth visits root and its subdirectories up to maxDepth levels
// below root, calling visit(dir, entries) for each - visit returns false to
// stop the walk early (result cap reached).
func walkLimitedDepth(root string, baseDepth, maxDepth int, visit func(dir string, entries []os.DirEntry) bool) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil //nolint:nilerr // unreadable dirs are skipped silently rather than failing the whole search
	}
	if !visit(root, entries) {
		return nil
	}
	if strings.Count(root, string(os.PathSeparator))-baseDepth >= maxDepth {
		return nil
	}
	for _, de := range entries {
		if de.IsDir() {
			if err := walkLimitedDepth(filepath.Join(root, de.Name()), baseDepth, maxDepth, visit); err != nil {
				return err
			}
		}
	}
	return nil
}
