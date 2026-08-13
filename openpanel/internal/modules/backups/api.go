package backups

import (
	"encoding/json"
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// RegisterAPI wires the /api/backups/* routes onto mux. Every handler
// below delegates to the existing UI handler (handleBackupSettings,
// handleBackupTarget, handleBackupsPage, handleListBackupsFromDestination,
// handleRestoreFromBackup, handleDownloadBackup) rather than
// reimplementing their logic: a cloned request is built with an
// "output=json" query param (so the handler's own ?output=json branch
// returns JSON instead of rendering HTML) and, where the UI handler reads
// a POST form, with Form/PostForm pre-populated from the API's JSON body
// so the same code path runs unchanged.
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "backups", "GET /api/backups", func(w http.ResponseWriter, r *http.Request) {
		handleBackupsPage(a, w, withJSONOutput(r))
	})
	apiregistry.Handle(mux, a, "backups", "GET /api/backups/settings", func(w http.ResponseWriter, r *http.Request) {
		handleBackupSettings(a, w, withJSONOutput(r))
	})
	apiregistry.Handle(mux, a, "backups", "PUT /api/backups/settings", func(w http.ResponseWriter, r *http.Request) {
		apiUpdateBackupSettings(a, w, r)
	})
	apiregistry.Handle(mux, a, "backups", "GET /api/backups/destination", func(w http.ResponseWriter, r *http.Request) {
		handleBackupTarget(a, w, withJSONOutput(r))
	})
	apiregistry.Handle(mux, a, "backups", "PUT /api/backups/destination", func(w http.ResponseWriter, r *http.Request) {
		apiSwitchBackupDestination(a, w, r)
	})
	apiregistry.Handle(mux, a, "backups", "GET /api/backups/list", func(w http.ResponseWriter, r *http.Request) {
		handleListBackupsFromDestination(a, w, withJSONOutput(r))
	})
	apiregistry.Handle(mux, a, "backups", "POST /api/backups/restore", func(w http.ResponseWriter, r *http.Request) {
		apiRestoreFromBackup(a, w, r)
	})
	apiregistry.Handle(mux, a, "backups", "POST /api/backups/download", func(w http.ResponseWriter, r *http.Request) {
		apiDownloadBackup(a, w, r)
	})
}

// withJSONOutput clones r with "output=json" set on the query string, so a
// UI handler that branches on that param returns JSON instead of HTML.
func withJSONOutput(r *http.Request) *http.Request {
	clone := r.Clone(r.Context())
	q := clone.URL.Query()
	q.Set("output", "json")
	clone.URL.RawQuery = q.Encode()
	return clone
}

// withForm clones r as a POST carrying the given values as both Form and
// PostForm, so a UI handler that reads r.Form.Get(...)/r.PostForm after
// its own ParseForm/ParseMultipartForm call (which is a no-op once r.Form
// is already set) sees exactly the fields the API's JSON body supplied.
func withForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}

func decodeJSONBody(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func writeAPIError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// apiUpdateBackupSettings translates the API's JSON body into the
// values[KEY]/settings[KEY] form fields handleBackupSettings expects from
// the UI's POST, then delegates to it.
func apiUpdateBackupSettings(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Values   map[string]string `json:"values"`
		Settings map[string]string `json:"settings"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	form := url.Values{}
	for k, v := range body.Values {
		form.Set("values["+k+"]", v)
	}
	for k, v := range body.Settings {
		form.Set("settings["+k+"]", v)
	}

	cloned := withJSONOutput(withForm(r, form))
	handleBackupSettings(a, w, cloned)
}

// apiSwitchBackupDestination translates the API's JSON body into the
// "target" form field handleBackupTarget expects from the UI's POST, then
// delegates to it.
func apiSwitchBackupDestination(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Target string `json:"target"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	form := url.Values{"target": {body.Target}}
	cloned := withJSONOutput(withForm(r, form))
	handleBackupTarget(a, w, cloned)
}

// apiRestoreFromBackup translates the API's JSON body into the
// backup_file/restore_target/database form fields handleRestoreFromBackup
// expects from the UI's multipart POST, then delegates to it.
func apiRestoreFromBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		BackupFile    string `json:"backup_file"`
		RestoreTarget string `json:"restore_target"`
		Database      string `json:"database"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	form := url.Values{
		"backup_file":    {body.BackupFile},
		"restore_target": {body.RestoreTarget},
		"database":       {body.Database},
	}
	handleRestoreFromBackup(a, w, withForm(r, form))
}

// apiDownloadBackup translates the API's JSON body into the backup_file
// form field handleDownloadBackup expects from the UI's multipart POST,
// then delegates to it (it streams the archive back as the response body).
func apiDownloadBackup(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		BackupFile string `json:"backup_file"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	form := url.Values{"backup_file": {body.BackupFile}}
	handleDownloadBackup(a, w, withForm(r, form))
}
