package imunifyav

import (
	"log"
	"net/http"
	"strings"
	"time"

	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var page = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/imunifyav.html",
)

var guiPage = web.MustLoadPage(
	"base.html",
	"partials/_header.html",
	"partials/_footer.html",
	"partials/_service.html",
	"partials/_search.html",
	"partials/_impersonate.html",
	"partials/_service_js.html",
	"partials/punnycode.html",
	"partials/theme_switcher.html",
	"files/imunifyav_gui.html",
)

// GUIPageData is files/imunifyav_gui.html's template context
type GUIPageData struct {
	web.LayoutData
	Token string
}

func renderGUI(w http.ResponseWriter, data GUIPageData) {
	if err := guiPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("IMUNIFYAV - gui template render error: %v", err)
	}
}

// PageData is files/imunifyav.html's template context
type PageData struct {
	web.LayoutData
	NotRunning bool
	// GUIReady links to imunify's own ui
	GUIReady bool
	Error    string
	Scanning bool
	LastScan string
	Infected int
	Hits     []HitRow
	Scans    []ScanRow
}

type HitRow struct {
	Path      string
	Signature string
	Detected  string
	Status    string
	Found     bool
}

type ScanRow struct {
	Started  string
	Duration string
	Files    int
	Found    int
	// Status is running, done, failed or aborted
	Status  string
	Message string
}

const timeLayout = "2006-01-02 15:04"

func unix(sec float64) time.Time { return time.Unix(int64(sec), 0) }

// displayPath shows site files the way the user sees them inside their containers
func displayPath(userContext, file string) string {
	docroot := home(userContext) + "/docker-data/volumes/" + userContext + "_html_data/_data"
	if rest, ok := strings.CutPrefix(file, docroot); ok && (rest == "" || strings.HasPrefix(rest, "/")) {
		return "/var/www/html" + rest
	}
	return file
}

func fillPage(data *PageData, userContext string, st userStatus, hits []hit, scans []scan, now time.Time) {
	data.Scanning = st.scanning()
	data.Infected = st.Infected + st.InfectedDB
	if st.ScanDate != nil && *st.ScanDate > 0 {
		data.LastScan = unix(*st.ScanDate).Format(timeLayout)
	}
	for _, h := range hits {
		row := HitRow{Path: displayPath(userContext, h.File), Signature: h.Type, Status: strings.ReplaceAll(h.Status, "_", " "), Found: h.Status == "found"}
		if h.Detected > 0 {
			row.Detected = unix(h.Detected).Format(timeLayout)
		}
		data.Hits = append(data.Hits, row)
	}
	for _, s := range scans {
		row := ScanRow{Started: unix(s.Started).Format(timeLayout), Files: s.TotalResources, Found: s.TotalMalicious, Status: "done"}
		switch {
		case s.Error != nil && *s.Error == "aborted":
			row.Status = "aborted"
		case s.Error != nil && *s.Error != "":
			row.Status, row.Message = "failed", *s.Error
		case s.Completed == nil:
			row.Status = "running"
		}
		end := now
		if s.Completed != nil {
			end = unix(*s.Completed)
		}
		if d := end.Sub(unix(s.Started)).Round(time.Second); d > 0 && row.Status != "running" {
			row.Duration = d.String()
		}
		data.Scans = append(data.Scans, row)
	}
}

func render(w http.ResponseWriter, data PageData) {
	if err := page.Render(w, http.StatusOK, data); err != nil {
		log.Printf("IMUNIFYAV - template render error: %v", err)
	}
}
