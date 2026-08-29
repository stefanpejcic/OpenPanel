package tinyphotogallery

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

// apiInstallTinyPhotoGallery delegates straight to handleInstallPage (which
// itself calls handleInstallStream on POST): same site-limit check, same
// NDJSON progress stream written directly to the response - just fed from
// the API's JSON body instead of a UI form post.
func apiInstallTinyPhotoGallery(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	var body struct {
		DomainID     string `json:"domain_id"`
		Subdirectory string `json:"subdirectory"`
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
	}
	handleInstallPage(a, w, withTinyPhotoGalleryForm(r, form))
}

// apiRemoveTinyPhotoGallery delegates to handleRemoveTinyPhotoGallery with
// the path's {site_id} translated into the "id" form field it expects,
// and output=json forced so it returns JSON instead of a flash-and-redirect.
func apiRemoveTinyPhotoGallery(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	siteID := r.PathValue("site_id")
	cloned := withTinyPhotoGalleryForm(r, url.Values{"id": {siteID}})
	q := cloned.URL.Query()
	q.Set("output", "json")
	cloned.URL.RawQuery = q.Encode()
	handleRemoveTinyPhotoGallery(a, w, cloned)
}
