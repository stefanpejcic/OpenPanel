package summary

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var summaryPage = web.LoadPage("system/summary.html")

// PageData is system/summary.html's template context, Tiles compare the day with the saved day before
type PageData struct {
	web.LayoutData
	Summary    Summary
	Found      bool
	Dates      []string
	Prev, Next string
	Date       string
	Tiles      Tiles
	EmailOn    bool
	Email      string
}

// Register wires the summary pages onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	h := auth.RequireLogin(a, "summary")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handlePage(a, w, r) }))
	mux.Handle("GET /summary", h)
	mux.Handle("GET /summary/{date}", h)
	mux.Handle("GET /summary/latest.json", auth.RequireLogin(a, "summary")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleLatest(a, w, r) })))
	mux.Handle("POST /summary/email", auth.RequireLogin(a, "summary")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleEmail(a, w, r) })))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// handleEmail turns the morning summary email on or off, the same key Account > Notifications saves
func handleEmail(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	username, _ := data["current_username"].(string)
	allowed, _ := data["user_allowed"].([]string)
	if username == "" || !slices.Contains(allowed, "notifications") {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Email notifications are not available on this account."})
		return
	}
	on := r.FormValue("on")
	if on != "0" && on != "1" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "on must be 0 or 1"})
		return
	}
	path := notificationsPath(username)
	content, _ := tenantfs.ReadFile(path)
	if err := tenantfs.MkdirAll(usersDir+username, 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save the setting."})
		return
	}
	if err := tenantfs.WriteFile(path, []byte(setPref(string(content), emailKey, on)), 0o644); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Could not save the setting."})
		return
	}
	// the notifier caches each key for 5 minutes
	_ = a.Cache.Delete(r.Context(), "get_from_file_value:"+username+":"+emailKey)
	action := "turned off daily summary emails"
	if on == "1" {
		action = "turned on daily summary emails"
	}
	_ = logger.RecordUserAction(a.Config, username, action, reqip.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]bool{"on": on == "1"})
}

// handleLatest tells the dashboard how many problems yesterday's summary found, so it can show a toast
func handleLatest(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, username, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	out := map[string]any{"found": false}
	date := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if s, found := Load(username, date); found {
		out = map[string]any{"found": true, "date": date, "attention": len(s.Attention), "look": len(s.Look), "url": "/summary/" + date}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(out)
}

func handlePage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, username, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dates := Dates(username)
	// the day picker uses ?date=, emails and the dashboard link to /summary/<date>
	asked := r.PathValue("date")
	if asked == "" {
		asked = r.URL.Query().Get("date")
	}
	date := asked
	if date == "" && len(dates) > 0 {
		date = dates[0]
	}

	status := http.StatusOK
	s, found := Load(username, date)
	if !found && asked != "" {
		status = http.StatusNotFound
	}
	prev, hasPrev := Load(username, dayBefore(date))

	layout, _, err := web.BuildLayoutData(a, w, r, "Daily Summary")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PageData{LayoutData: layout, Summary: s, Found: found, Dates: dates, Date: date, Tiles: tiles(s, prev, hasPrev), EmailOn: emailOn(username)}
	if injected, err := a.InjectData(r.Context(), userID); err == nil {
		data.Email, _ = injected["current_email"].(string)
	}
	// dates are newest first, so the previous day is the next index
	for i, d := range dates {
		if d == date {
			if i+1 < len(dates) {
				data.Prev = dates[i+1]
			}
			if i > 0 {
				data.Next = dates[i-1]
			}
		}
	}
	if err := summaryPage.Render(w, status, data); err != nil {
		log.Printf("SUMMARY - template render error: %v", err)
	}
}
