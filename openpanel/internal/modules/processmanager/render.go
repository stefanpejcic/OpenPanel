package processmanager

import (
	"log"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var processManagerPage = web.LoadPage("system/process_manager.html")

// ProcessManagerPageData is system/process_manager.html's template context.
type ProcessManagerPageData struct {
	web.LayoutData
	ProcessData []Process
}

// Containers is the sorted list of containers that have processes, for the filter and the summary
func (d ProcessManagerPageData) Containers() []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range d.ProcessData {
		if !seen[p.Container] {
			seen[p.Container] = true
			out = append(out, p.Container)
		}
	}
	sort.Strings(out)
	return out
}

// Busiest is the process using the most CPU, nil when there are none
func (d ProcessManagerPageData) Busiest() *Process {
	var top *Process
	best := -1.0
	for i := range d.ProcessData {
		if c, err := strconv.ParseFloat(d.ProcessData[i].C, 64); err == nil && c > best {
			best, top = c, &d.ProcessData[i]
		}
	}
	return top
}

// ProgramName is the executable's base name from a command line, e.g. "php-fpm: master process" -> "php-fpm"
func (d ProcessManagerPageData) ProgramName(cmd string) string {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return cmd
	}
	name := path.Base(fields[0])
	return strings.TrimSuffix(name, ":")
}

func renderProcessManagerPage(a *appctx.App, w http.ResponseWriter, r *http.Request, processes []Process) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Process Manager")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	layout.BulkActions = processBulkActions(layout.T)
	data := ProcessManagerPageData{LayoutData: layout, ProcessData: processes}
	if err := processManagerPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("PROCESSMANAGER - template render error: %v", err)
	}
}
