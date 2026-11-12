// Package imunifyav shows a user their own ImunifyAV results and lets them queue a scan of their home.
package imunifyav

import (
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

const (
	pagePath    = "/imunifyav"
	summaryPath = "/imunifyav/summary"
)

var contextRE = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "imunifyav")(h)
	}
	mux.Handle("GET "+pagePath, requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleGUI(a, w, r)
	}))
	mux.Handle("GET "+summaryPath, requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handlePage(a, w, r)
	}))
	mux.Handle("POST "+pagePath+"/scan", requireLogin(func(w http.ResponseWriter, r *http.Request) {
		handleScan(a, w, r)
	}))
	proxy := newGUIProxy()
	mux.Handle("/imav/", requireLogin(proxy.ServeHTTP))
}

// timestamps are floats while a scan runs and ints after, so they all decode as float64
// userStatus is the user's row from malware user list, user is the system user which is the panel user's context
type userStatus struct {
	Infected   int      `json:"infected"`
	InfectedDB int      `json:"infected_db"`
	ScanDate   *float64 `json:"scan_date"`
	ScanStatus *string  `json:"scan_status"`
}

func (u userStatus) scanning() bool {
	if u.ScanStatus == nil {
		return false
	}
	switch *u.ScanStatus {
	case "running", "queued", "pending":
		return true
	}
	return false
}

type hit struct {
	File     string  `json:"file"`
	Type     string  `json:"type"`
	Detected float64 `json:"detected"`
	Status   string  `json:"status"`
}

type scan struct {
	Path           string   `json:"path"`
	Started        float64  `json:"started"`
	Completed      *float64 `json:"completed"`
	Error          *string  `json:"error"`
	TotalMalicious int      `json:"total_malicious"`
	TotalResources int      `json:"total_resources"`
	ScanStatus     string   `json:"scan_status"`
	ScanType       string   `json:"scan_type"`
}

func home(userContext string) string { return "/home/" + userContext }

// fetchStatus reads imunify's per-user row, a user it hasn't seen yet comes back as a zero row
var fetchStatus = func(userContext string) (userStatus, error) {
	data, err := call([]string{"malware", "user", "list"}, map[string]any{"user": userContext})
	if err != nil {
		return userStatus{}, err
	}
	var rows []userStatus
	if err := items(data, &rows); err != nil || len(rows) == 0 {
		return userStatus{}, err
	}
	return rows[0], nil
}

// fetchHits is imunify's malicious list filtered to files the user owns, newest first
var fetchHits = func(userContext string) ([]hit, error) {
	data, err := call([]string{"malware", "malicious", "list"}, map[string]any{"user": userContext, "limit": 500})
	if err != nil {
		return nil, err
	}
	var hits []hit
	if err := items(data, &hits); err != nil {
		return nil, err
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Detected > hits[j].Detected })
	return hits, nil
}

// fetchScans keeps the on-demand scans of the user's home, the agent lists every user's scans to root
var fetchScans = func(userContext string) ([]scan, error) {
	data, err := call([]string{"malware", "on-demand", "list"}, map[string]any{"limit": 500})
	if err != nil {
		return nil, err
	}
	var all []scan
	if err := items(data, &all); err != nil {
		return nil, err
	}
	h := home(userContext)
	var mine []scan
	for _, s := range all {
		if s.Path == h || strings.HasPrefix(s.Path, h+"/") {
			mine = append(mine, s)
		}
	}
	sort.SliceStable(mine, func(i, j int) bool { return mine[i].Started > mine[j].Started })
	if len(mine) > 20 {
		mine = mine[:20]
	}
	return mine, nil
}

// queueScan puts the user's home on imunify's on-demand queue, the same scan OpenAdmin starts per user
// scanAllowed is imunify's "users can start scans" setting, the root socket doesn't check it for us
var scanAllowed = func() (bool, error) {
	data, err := call([]string{"config", "show"}, map[string]any{})
	if err != nil {
		return false, err
	}
	var cfg struct {
		Permissions struct {
			AllowMalwareScan bool `json:"allow_malware_scan"`
		} `json:"PERMISSIONS"`
	}
	if err := items(data, &cfg); err != nil {
		return false, err
	}
	return cfg.Permissions.AllowMalwareScan, nil
}

var queueScan = func(userContext string) error {
	_, err := call([]string{"malware", "on-demand", "queue", "put"}, map[string]any{"paths": []string{home(userContext)}})
	return err
}

func userContextOf(injected map[string]any) (string, string, bool) {
	username, _ := injected["current_username"].(string)
	userContext, _ := injected["context"].(string)
	return username, userContext, contextRE.MatchString(userContext)
}

// handleGUI embeds imunify's own ui logged in as the user, or shows the summary when that ui isn't reachable
func handleGUI(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	if !available() || !guiReady() {
		handlePage(a, w, r)
		return
	}
	layout, injected, err := web.BuildLayoutData(a, w, r, "ImunifyAV")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_, userContext, ok := userContextOf(injected)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	token, err := loginToken(userContext)
	if err != nil {
		log.Printf("IMUNIFYAV - login token for %s: %v", userContext, err)
		handlePage(a, w, r)
		return
	}
	renderGUI(w, GUIPageData{LayoutData: layout, Token: token})
}

func handlePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	layout, injected, err := web.BuildLayoutData(a, w, r, "ImunifyAV")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_, userContext, ok := userContextOf(injected)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PageData{LayoutData: layout, GUIReady: available() && guiReady()}
	if !available() {
		data.NotRunning = true
		render(w, data)
		return
	}

	var (
		wg                sync.WaitGroup
		st                userStatus
		hits              []hit
		scans             []scan
		stErr, hErr, sErr error
	)
	wg.Add(4)
	go func() {
		defer wg.Done()
		allowed, err := scanAllowed()
		if err != nil {
			log.Printf("IMUNIFYAV - read scan permission: %v", err)
		}
		data.ScanAllowed = allowed
	}()
	go func() { defer wg.Done(); st, stErr = fetchStatus(userContext) }()
	go func() { defer wg.Done(); hits, hErr = fetchHits(userContext) }()
	go func() { defer wg.Done(); scans, sErr = fetchScans(userContext) }()
	wg.Wait()
	for _, e := range []error{stErr, hErr, sErr} {
		if e != nil {
			log.Printf("IMUNIFYAV - %s: %v", userContext, e)
			if e == errNotRunning {
				data.NotRunning = true
			} else {
				data.Error = web.Tr(a, r, "Couldn't load your ImunifyAV results, try again in a minute.")
			}
			render(w, data)
			return
		}
	}
	fillPage(&data, userContext, st, hits, scans, time.Now())
	render(w, data)
}

func handleScan(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	injected, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, userContext, ok := userContextOf(injected)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !available() {
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "ImunifyAV is not running on this server."), summaryPath)
		return
	}
	if allowed, err := scanAllowed(); err != nil || !allowed {
		if err != nil {
			log.Printf("IMUNIFYAV - read scan permission: %v", err)
		}
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Starting scans is turned off on this server."), summaryPath)
		return
	}
	// a second put while one is queued makes imunify abort the first, so don't stack them
	if st, err := fetchStatus(userContext); err == nil && st.scanning() {
		web.FlashRedirect(a, w, r, "info", web.Tr(a, r, "A scan of your files is already running."), summaryPath)
		return
	}
	if err := queueScan(userContext); err != nil {
		log.Printf("IMUNIFYAV - queue scan for %s: %v", userContext, err)
		web.FlashRedirect(a, w, r, "error", web.Tr(a, r, "Couldn't start the scan, try again in a minute."), summaryPath)
		return
	}
	_ = logger.RecordUserAction(a.Config, username, "started an ImunifyAV scan", reqip.ClientIP(r))
	web.FlashRedirect(a, w, r, "success", web.Tr(a, r, "Scan started. Results show up here when it finishes."), summaryPath)
}
