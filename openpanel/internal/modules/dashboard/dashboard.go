package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/csrf"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/emails"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var dashboardPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"dashboard/dashboard.html",
	"dashboard/custom_message.html",
	"dashboard/custom_section.html",
	"dashboard/twofa.html",
	"dashboard/information.html",
	"dashboard/usage.html",
	"dashboard/howto.html",
)

// Register wires dashboard.py's routes onto mux. "dashboard" is always in
// EnabledModules (app.py forces it via main_modules), so this registers
// unconditionally like the always-on account routes, not through the
// enabled_modules dispatch table.
func Register(mux *http.ServeMux, a *appctx.App) {
	mux.Handle("/json/resource_usage", auth.RequireLogin(a, "dashboard")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleResourceUsage(a, w, r)
	})))
	mux.Handle("/json/disk_inodes", auth.RequireLogin(a, "dashboard")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDiskInodes(a, w, r)
	})))
	mux.HandleFunc("/openpanel", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	})
	// net/http.ServeMux's "/" pattern is a catch-all for any path no other
	// registered pattern matches (Go 1.22+ routing semantics) - unlike
	// Flask's @app.route('/'), which only ever matches the literal root and
	// leaves every other unmatched path to Flask's own 404, with no login
	// check involved at all (routing happens before any decorator runs).
	// The path check has to sit in front of RequireLogin, not behind it -
	// otherwise an unknown route would redirect an unauthenticated visitor
	// to /login instead of 404ing, which still isn't what Flask does.
	rootHandler := auth.RequireLogin(a, "dashboard")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		rootHandler.ServeHTTP(w, r)
	}))
	mux.Handle("POST /dashboard/tour/complete", auth.RequireLogin(a, "dashboard")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleTourComplete(a, w, r)
	})))
	mux.Handle("/dashboard", auth.RequireLogin(a, "dashboard")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDashboard(a, w, r)
	})))
}

func handleDashboard(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)

	injected, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	data, err := buildDashboardData(a, ctx, userID, injected)
	if err != nil {
		log.Printf("DASHBOARD - failed to build dashboard data: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	page := buildDashboardPageData(a, w, r, injected, data)
	if err := dashboardPage.Render(w, http.StatusOK, page); err != nil {
		log.Printf("DASHBOARD - template render error: %v", err)
	}
}

// buildDashboardPageData assembles the shared app-shell data (nav,
// flashes, branding - the Go analogue of app.py's inject_data()/injected()
// context processors) plus dashboard()'s own template kwargs into one
// DashboardPageData.
func buildDashboardPageData(a *appctx.App, w http.ResponseWriter, r *http.Request, injected map[string]any, d DashboardData) DashboardPageData {
	ctx := r.Context()
	sess, _ := a.Sessions.Get(r, session.CookieName)

	userAllowedSlice, _ := injected["user_allowed"].([]string)
	userAllowed := make(map[string]bool, len(userAllowedSlice))
	for _, m := range userAllowedSlice {
		userAllowed[m] = true
	}

	currentUsername, _ := injected["current_username"].(string)
	sessionLocale, _ := sess.Values["locale"].(string)
	userLocale := i18nUserLocale(injected)
	locale := a.I18n.ResolveLocale(ctx, sessionLocale, userLocale, r.Header.Get("Accept-Language"))
	t := a.I18n.Translator(locale)

	adminPort, _ := sess.Values["admin_port"].(string)
	if adminPort == "" {
		adminPort = "2087"
	}
	_, impersonating := sess.Values["impersonate"]

	hostingPlanName, _ := injected["hosting_plan_name"].(string)
	avatarType, _ := injected["avatar_type"].(string)
	gravatarURL, _ := injected["gravatar_image_url"].(string)
	panelVersion, _ := injected["panel_version"].(string)
	panelDir, _ := injected["panel_dir"].(string)
	isEnterprise, _ := injected["is_enterprise"].(bool)

	layout := web.LayoutData{
		Title:           "Dashboard",
		BrandName:       a.Config.Get("brand_name", ""),
		Logo:            a.Config.Get("logo", ""),
		Favicon:         a.Config.Get("favicon", ""),
		CSRFToken:       csrf.Token(r),
		PanelDir:        panelDir,
		FoundABugLink:   a.Config.Get("found_a_bug_link", ""),
		PanelVersion:    panelVersion,
		CustomPlugins:   len(a.PluginNames) > 0,
		CustomCSS:       a.CustomCSS,
		CustomJS:        true, // matches base.html's always-true url_for() guard, see base.html's comment on this - not tied to a.CustomJS on purpose
		NavGroups:       web.BuildSidebarNav(userAllowed, web.NavPath(r)),
		UserAllowed:     userAllowed,
		UserAllowedJSON: web.UserAllowedList(userAllowed),
		IsEnterprise:    isEnterprise,
		CurrentUsername: currentUsername,
		HostingPlanName: hostingPlanName,
		AvatarType:      avatarType,
		GravatarURL:     gravatarURL,
		RequestPath:     r.URL.Path,
		Flashes:         web.BuildFlashDisplay(flash.Pop(a.Sessions, w, r, sess)),
		Impersonating:   impersonating,
		AdminPort:       adminPort,
		T:               t,
	}

	return DashboardPageData{
		LayoutData:            layout,
		Sections:              buildDashboardSections(userAllowed),
		TourShow:              d.TourShow,
		CustomMessage:         template.HTML(d.CustomMessage), //nolint:gosec // matches Jinja's `custom_message|safe`: admin-authored HTML from a local file, not user input
		CustomSectionTitle:    d.CustomSectionTitle,
		CustomSectionItems:    convertCustomSectionItems(d.CustomSectionItems),
		CustomSectionPosition: d.CustomSectionPosition,
		TwofaEnabled:          d.TwofaEnabled,
		TwofaNag:              d.TwofaNag,
		TwofaStatusMessage:    twofaStatusMessage(t, d.TwofaEnabled),
		IPAddress:             d.IPAddress,
		LastIP:                d.LastIP,
		IPCountyFlag:          d.IPCountyFlag,
		NS1:                   d.NS1,
		NS2:                   d.NS2,
		NS3:                   d.NS3,
		NS4:                   d.NS4,
		HowToGuides:           d.HowToGuides,
		HowToTopics:           convertHowToTopics(d.HowToTopics),
		KnowledgeBaseLink:     d.KnowledgeBaseLink,
		UserWebsitesCount:     len(d.UserWebsites),
		MainDomainsCount:      len(d.MainDomains),
		DBUsage:               d.DBUsage,
		EmailCount:            d.EmailCount,
		FTPCount:              d.FTPCount,
		WebsitesLimit:         atoiDefault(d.Plan.WebsitesLimit, 0),
		DomainsLimit:          atoiDefault(d.Plan.DomainsLimit, 0),
		DBLimit:               atoiDefault(d.Plan.DBLimit, 0),
		EmailLimit:            atoiDefault(d.Plan.EmailLimit, 0),
		FTPLimit:              atoiDefault(d.Plan.FTPLimit, 0),
	}
}

// i18nUserLocale mirrors get_locale()'s per-account locale-file tier,
// reading /home/<context>/locale via i18n.UserLocale.
func i18nUserLocale(injected map[string]any) string {
	userContext, _ := injected["context"].(string)
	return i18n.UserLocale(userContext)
}

func convertCustomSectionItems(items []map[string]any) []CustomSectionItem {
	result := make([]CustomSectionItem, 0, len(items))
	for _, m := range items {
		url, _ := m["url"].(string)
		icon, _ := m["icon"].(string)
		label, _ := m["label"].(string)
		result = append(result, CustomSectionItem{URL: url, Icon: icon, Label: label})
	}
	return result
}

func convertHowToTopics(topics []map[string]any) []HowToTopic {
	result := make([]HowToTopic, 0, len(topics))
	for _, m := range topics {
		link, _ := m["link"].(string)
		title, _ := m["title"].(string)
		result = append(result, HowToTopic{Link: link, Title: title})
	}
	return result
}

func handleResourceUsage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	userContext, _ := data["context"].(string)
	if username == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	stats, err := a.GetResourceUsage(r.Context(), username, userContext)
	if err != nil || stats == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

func handleDiskInodes(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	if username == "" {
		writeJSONError(w, http.StatusNotFound, "User ID not found.")
		return
	}

	disk, err := getDiskAndInodesForUser(a, r.Context(), username)
	if err != nil || disk == nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to retrieve disk usage information.")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(disk)
}

func handleTourComplete(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := data["context"].(string)
	tourSkipFile := fmt.Sprintf("/home/%s/skip.tour", userContext)

	if _, err := os.Stat(tourSkipFile); os.IsNotExist(err) {
		if err := os.WriteFile(tourSkipFile, nil, 0o644); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "Could not save tour state.")
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// DashboardData mirrors every value dashboard.py's dashboard() route
// passes to render_template. Ready for the template once it exists.
type DashboardData struct {
	TwofaEnabled          bool
	TwofaNag              string
	DBUsage               int
	IPAddress             string
	IPCountyFlag          string
	HowToGuides           string
	HowToTopics           []map[string]any
	KnowledgeBaseLink     string
	CustomMessage         string
	CustomSectionTitle    string
	CustomSectionItems    []map[string]any
	CustomSectionPosition string
	LastIP                string
	NS1, NS2, NS3, NS4    string
	Plan                  appctx.PlanDetails
	CPULimit              int
	RAMLimit              int
	EmailCount            int
	Domains               []appctx.Domain
	MainDomains           []appctx.MainDomain
	Subdomains            []appctx.Subdomain
	UserWebsites          []UserWebsite
	FTPCount              int
	TourShow              bool
}

type UserWebsite struct {
	SiteName string
	Type     string
}

// buildDashboardData mirrors dashboard()'s full body (everything up to the
// render_template call, which doesn't exist yet - see the note in
// Register).
func buildDashboardData(a *appctx.App, ctx context.Context, userID int, injected map[string]any) (DashboardData, error) {
	currentUsername, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)
	planID, _ := injected["hosting_plan"].(int)
	userAllowedSlice, _ := injected["user_allowed"].([]string)
	userAllowed := map[string]bool{}
	for _, m := range userAllowedSlice {
		userAllowed[m] = true
	}

	var d DashboardData
	d.IPAddress = a.GetCachedIPForUserOrPublicIPv4(ctx, currentUsername)
	d.IPCountyFlag = a.Config.Get("ip_county_flag", "yes")

	if userAllowed["mysql"] {
		d.DBUsage = getDatabaseCount(a, ctx, currentUsername, userContext)
	}

	var userDomains []appctx.Domain
	if userAllowed["domains"] || userAllowed["emails"] {
		var err error
		userDomains, err = a.AllDomainsForUser(ctx, userID)
		if err != nil {
			return DashboardData{}, err
		}
		d.Domains = userDomains
	}
	if userAllowed["domains"] {
		d.MainDomains, d.Subdomains = appctx.Categorize(userDomains)
	}
	if userAllowed["emails"] {
		domainSet := make(map[string]bool, len(userDomains))
		for _, dm := range userDomains {
			domainSet[dm.DomainURL] = true
		}
		d.EmailCount = emails.GetEmailCount(ctx, a, userID, currentUsername, domainSet)
	}

	websites, err := GetUserWebsites(a, ctx, userID)
	if err != nil {
		return DashboardData{}, err
	}
	d.UserWebsites = websites

	if userAllowed["ftp"] {
		d.FTPCount = countFTPAccounts(a, ctx, userContext)
	}

	plan, err := a.QueryPlanDetailsByID(ctx, planID)
	if err != nil {
		return DashboardData{}, err
	}
	d.Plan = plan
	d.CPULimit = atoiDefault(plan.CPU, 0) * 100
	d.RAMLimit = atoiDefault(strings.TrimRight(plan.RAM, "gG"), 0)

	d.CustomMessage = customMessageForUser(a, ctx, currentUsername)

	lastLogins, err := a.GetLastLoginData(ctx, currentUsername)
	if err != nil {
		return DashboardData{}, err
	}
	if len(lastLogins) >= 2 {
		d.LastIP = lastLogins[len(lastLogins)-2].IP
	} else if len(lastLogins) == 1 {
		d.LastIP = lastLogins[0].IP
	}

	d.HowToGuides = a.Config.Get("how_to_guides", "")
	topics, kbLink := howToLinksForDashboard(a, ctx)
	d.HowToTopics, d.KnowledgeBaseLink = topics, kbLink

	title, items, position := customSectionForDashboard(a, ctx)
	d.CustomSectionTitle, d.CustomSectionItems, d.CustomSectionPosition = title, items, position

	if userAllowed["ftp"] && a.Config.Get("twofa_nag", "") == "yes" {
		status, err := a.Get2FAStatusForUser(ctx, userID)
		if err == nil {
			d.TwofaEnabled = status.Enabled
		}
	}
	d.TwofaNag = a.Config.Get("twofa_nag", "")
	d.NS1 = a.Config.Get("ns1", "")
	d.NS2 = a.Config.Get("ns2", "")
	d.NS3 = a.Config.Get("ns3", "")
	d.NS4 = a.Config.Get("ns4", "")

	tourSkipFile := fmt.Sprintf("/home/%s/skip.tour", userContext)
	_, statErr := os.Stat(tourSkipFile)
	d.TourShow = os.IsNotExist(statErr)

	return d, nil
}

// restrictedDatabasesSQL mirrors mysql.py's restricted_databases_sql:
// config's mysql_restricted_databases (space-separated, optionally quoted)
// rendered as a SQL string list for `NOT IN (...)`.
func restrictedDatabasesSQL(cfg config.Config) string {
	raw := stripQuotes(cfg.Get("mysql_restricted_databases",
		"information_schema performance_schema mysql phpmyadmin sys mariadb.sys"))
	fields := strings.Fields(raw)
	quoted := make([]string, len(fields))
	for i, d := range fields {
		quoted[i] = "'" + strings.Trim(strings.TrimSpace(d), `"'`) + "'"
	}
	return strings.Join(quoted, ", ")
}

// stripQuotes mirrors modules/files/filemanager.py's strip_quotes(): strip
// one layer of surrounding quotes if the whole string is quote-wrapped.
func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (strings.HasPrefix(s, `"`) && strings.HasSuffix(s, `"`)) ||
			(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// getDatabaseCount mirrors modules/mysql.py's get_database_count(),
// cached 1h. Connects to the user's own MySQL instance via mysqlmanager
// (a unix socket under /home/<context>/sockets/mysqld/mysqld.sock, not
// the panel's own DB) - returns 0 on any failure (container not running,
// socket not found, ...), same as Python's broad except.
func getDatabaseCount(a *appctx.App, ctx context.Context, username, userContext string) int {
	count, _ := cache.Memoize(ctx, a.Cache, "get_database_count:"+username, time.Hour, func() (int, error) {
		query := "SELECT COUNT(*) AS total FROM information_schema.schemata WHERE schema_name NOT IN (" +
			restrictedDatabasesSQL(a.Config) + ")"
		rows, err := mysqlmanager.Exec(ctx, userContext, query, "")
		if err != nil || len(rows) == 0 || len(rows[0]) == 0 {
			return 0, nil //nolint:nilerr // matches Python: any failure -> 0, not an error response
		}
		return mysqlmanager.ToInt(rows[0][0]), nil
	})
	return count
}

func GetUserWebsites(a *appctx.App, ctx context.Context, userID int) ([]UserWebsite, error) {
	return cache.Memoize(ctx, a.Cache, fmt.Sprintf("get_user_websites:%d", userID), 30*time.Second, func() ([]UserWebsite, error) {
		rows, err := a.DB.QueryContext(ctx, `
			SELECT site_name, type FROM sites
			WHERE domain_id IN (SELECT domain_id FROM domains WHERE user_id = ?) LIMIT 1000`, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var sites []UserWebsite
		for rows.Next() {
			var (
				s        UserWebsite
				siteName sql.NullString
				siteType sql.NullString
			)
			if err := rows.Scan(&siteName, &siteType); err != nil {
				return nil, err
			}
			s.SiteName, s.Type = siteName.String, siteType.String
			sites = append(sites, s)
		}
		return sites, rows.Err()
	})
}

type quotaReport struct {
	Timestamp string      `json:"timestamp"`
	Users     []quotaUser `json:"users"`
}

type quotaUser struct {
	Username   string `json:"username"`
	InodesUsed int64  `json:"inodes_used"`
	InodesSoft int64  `json:"inodes_soft"`
	InodesHard int64  `json:"inodes_hard"`
	DiskUsed   int64  `json:"disk_used"`
	DiskSoft   int64  `json:"disk_soft"`
	DiskHard   int64  `json:"disk_hard"`
	HomePath   string `json:"home_path"`
}

func loadQuotaReport(a *appctx.App, ctx context.Context) (*quotaReport, error) {
	return cache.Memoize(ctx, a.Cache, "_load_quota_report", 1800*time.Second, func() (*quotaReport, error) {
		data, err := os.ReadFile("/etc/openpanel/openpanel/quota_report.json")
		if err != nil {
			return nil, nil //nolint:nilerr // matches Python: missing/bad file -> None, not an error
		}
		var report quotaReport
		if err := json.Unmarshal(data, &report); err != nil {
			return nil, nil //nolint:nilerr
		}
		return &report, nil
	})
}

// DiskInodes mirrors get_disk_and_inodes_for_user()'s dict shape.
type DiskInodes struct {
	InodesUsed int64  `json:"inodes_used"`
	InodesSoft int64  `json:"inodes_soft"`
	InodesHard int64  `json:"inodes_hard"`
	DiskUsed   int64  `json:"disk_used"`
	DiskSoft   int64  `json:"disk_soft"`
	DiskHard   int64  `json:"disk_hard"`
	Device     string `json:"device"`
	Date       string `json:"date"`
}

func getDiskAndInodesForUser(a *appctx.App, ctx context.Context, username string) (*DiskInodes, error) {
	return cache.Memoize(ctx, a.Cache, "get_disk_and_inodes_for_user:"+username, 1800*time.Second, func() (*DiskInodes, error) {
		report, err := loadQuotaReport(a, ctx)
		if err != nil || report == nil {
			return nil, err
		}
		for _, u := range report.Users {
			if u.Username != username {
				continue
			}
			device := u.HomePath
			if device == "" {
				device = "/home"
			}
			return &DiskInodes{
				InodesUsed: u.InodesUsed, InodesSoft: u.InodesSoft, InodesHard: u.InodesHard,
				DiskUsed: u.DiskUsed, DiskSoft: u.DiskSoft, DiskHard: u.DiskHard,
				Device: device, Date: report.Timestamp,
			}, nil
		}
		return nil, nil
	})
}

func countFTPAccounts(a *appctx.App, ctx context.Context, userContext string) int {
	count, _ := cache.Memoize(ctx, a.Cache, "count_ftp_accounts:"+userContext, 360*time.Second, func() (int, error) {
		path := fmt.Sprintf("/etc/openpanel/ftp/users/%s/users.list", userContext)
		data, err := os.ReadFile(path)
		if err != nil {
			return 0, nil //nolint:nilerr
		}
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		if len(data) == 0 {
			return 0, nil
		}
		return len(lines), nil
	})
	return count
}

func customMessageForUser(a *appctx.App, ctx context.Context, username string) string {
	msg, _ := cache.Memoize(ctx, a.Cache, "custom_message_for_user_on_dashboard_page:"+username, 3600*time.Second, func() (string, error) {
		path := fmt.Sprintf("/etc/openpanel/openpanel/core/users/%s/custom_message.html", username)
		data, err := os.ReadFile(path)
		if err != nil {
			return "", nil //nolint:nilerr
		}
		return string(data), nil
	})
	return msg
}

func readJSONFile(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var v map[string]any
	if json.Unmarshal(data, &v) != nil {
		return nil
	}
	return v
}

func howToLinksForDashboard(a *appctx.App, ctx context.Context) ([]map[string]any, string) {
	type result struct {
		Topics []map[string]any
		Link   string
	}
	r, _ := cache.Memoize(ctx, a.Cache, "how_to_links_for_dashboard_page", 86400*time.Second, func() (result, error) {
		if a.Config.Get("how_to_guides", "") == "" {
			return result{}, nil
		}
		kbData := readJSONFile("/etc/openpanel/openpanel/conf/knowledge_base_articles.json")
		if kbData == nil {
			// Matches os.path.join(app.root_path, ...) in Python: app.py
			// (and this default JSON, shipped alongside the Python
			// templates/) lives at the container image's "/".
			kbData = readJSONFile("/templates/dashboard/knowledge_base_articles_default.json")
		}
		if kbData == nil {
			kbData = fetchRemoteKBData()
		}
		if kbData == nil {
			return result{}, nil
		}
		topics, _ := kbData["how_to_topics"].([]any)
		var topicMaps []map[string]any
		for _, t := range topics {
			if m, ok := t.(map[string]any); ok {
				topicMaps = append(topicMaps, m)
			}
		}
		link, _ := kbData["knowledge_base_link"].(string)
		return result{Topics: topicMaps, Link: link}, nil
	})
	return r.Topics, r.Link
}

func fetchRemoteKBData() map[string]any {
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/conf/knowledge_base_articles.json")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var v map[string]any
	if json.NewDecoder(resp.Body).Decode(&v) != nil {
		return nil
	}
	return v
}

func customSectionForDashboard(a *appctx.App, ctx context.Context) (string, []map[string]any, string) {
	type result struct {
		Title    string
		Items    []map[string]any
		Position string
	}
	r, _ := cache.Memoize(ctx, a.Cache, "custom_section_for_dashboard_page", 28800*time.Second, func() (result, error) {
		data := readJSONFile("/etc/openpanel/openpanel/conf/custom_dashboard_section.json")
		if data == nil {
			return result{}, nil
		}
		itemsRaw, ok := data["items"]
		if !ok {
			return result{}, nil
		}
		items, _ := itemsRaw.([]any)
		var itemMaps []map[string]any
		for _, it := range items {
			if m, ok := it.(map[string]any); ok {
				itemMaps = append(itemMaps, m)
			}
		}
		title, _ := data["section_title"].(string)
		position, _ := data["section_position"].(string)
		return result{Title: title, Items: itemMaps, Position: position}, nil
	})
	return r.Title, r.Items, r.Position
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}
