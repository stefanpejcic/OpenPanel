package phpbb

import (
	"encoding/json"
	"net/http"
	"net/url"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

func writeAPIJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func apiInstallPhpbb(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID         string `json:"domain_id"`
		Subdirectory     string `json:"subdirectory"`
		BoardName        string `json:"board_name"`
		BoardDescription string `json:"board_description"`
		AdminUsername    string `json:"admin_username"`
		AdminPassword    string `json:"admin_password"`
		AdminEmail       string `json:"admin_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}
	if body.DomainID == "" {
		writeAPIJSON(w, http.StatusBadRequest, map[string]string{"error": "domain_id is required"})
		return
	}

	form := url.Values{
		"domain_id": {body.DomainID}, "subdirectory": {body.Subdirectory},
		"board_name": {body.BoardName}, "board_description": {body.BoardDescription},
		"admin_username": {body.AdminUsername}, "admin_password": {body.AdminPassword}, "admin_email": {body.AdminEmail},
	}
	handleInstallPage(a, w, withPhpbbForm(r, form))
}

func apiRemovePhpbb(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withPhpbbForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemovePhpbb(a, w, cloned)
}
