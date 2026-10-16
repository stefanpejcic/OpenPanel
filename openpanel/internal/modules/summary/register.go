package summary

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var summaryPage = web.LoadPage("system/summary.html")

// PageData is system/summary.html's template context
type PageData struct {
	web.LayoutData
	Summary    Summary
	Found      bool
	Dates      []string
	Prev, Next string
}

// Register wires the summary pages onto mux.
func Register(mux *http.ServeMux, a *appctx.App) {
	h := auth.RequireLogin(a, "summary")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handlePage(a, w, r) }))
	mux.Handle("GET /summary", h)
	mux.Handle("GET /summary/{date}", h)
	mux.Handle("GET /summary/latest.json", auth.RequireLogin(a, "summary")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleLatest(a, w, r) })))
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
	_, username, _, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	dates := Dates(username)
	date := r.PathValue("date")
	if date == "" && len(dates) > 0 {
		date = dates[0]
	}

	status := http.StatusOK
	s, found := Load(username, date)
	if !found && r.PathValue("date") != "" {
		status = http.StatusNotFound
	}

	layout, _, err := web.BuildLayoutData(a, w, r, "Daily Summary")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data := PageData{LayoutData: layout, Summary: s, Found: found, Dates: dates}
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
