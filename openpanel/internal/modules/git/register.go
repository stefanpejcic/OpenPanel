package git

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// Register wires the Git page, the site manager tab's JSON, and the public push/webhook endpoints which do their own auth
func Register(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "git")(h)
	}
	mux.Handle("/git", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleGitPage(a, w, r) }))
	mux.Handle("GET /git/site.json", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleSiteJSON(a, w, r) }))
	mux.Handle("POST /git/action.json", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleGitJSON(a, w, r) }))
	mux.Handle("POST /git/bulk", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleGitBulk(a, mux, w, r) }))

	mux.HandleFunc(httpPrefix, func(w http.ResponseWriter, r *http.Request) { handleSmartHTTP(a, w, r) })
	mux.HandleFunc("POST "+webhookPrefix+"{user}/{token}", func(w http.ResponseWriter, r *http.Request) { handleWebhook(a, w, r) })
}

// RegisterAPI wires the /api/git twins
func RegisterAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "git", "GET /api/git", func(w http.ResponseWriter, r *http.Request) { apiList(a, w, r) })
	apiregistry.Handle(mux, a, "git", "POST /api/git", func(w http.ResponseWriter, r *http.Request) { apiAction(a, w, r) })
}

// IsPublicPath reports whether a path is a git client or webhook call, which carry their own credentials instead of a CSRF token
func IsPublicPath(p string) bool {
	return strings.HasPrefix(p, httpPrefix) || strings.HasPrefix(p, webhookPrefix)
}

func apiList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	_, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	repos, err := loadRepos(userContext)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	web.WriteJSON(w, http.StatusOK, map[string]any{"repos": buildViews(r.Context(), userContext, username, web.PublicBaseURL(r.Context(), a, r), repos, r.URL.Query().Get("domain"))})
}

// apiAction takes the same fields as the page form, as JSON or form values
func apiAction(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, username, userContext, err := auth.Injected(a, r)
	if err != nil {
		web.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
		return
	}
	values, err := requestValues(r)
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body."})
		return
	}
	res, err := doAction(r.Context(), a, r, userID, username, userContext, values)
	if err != nil {
		web.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	web.WriteJSON(w, http.StatusOK, res)
}

// requestValues accepts a flat JSON object or a regular form post
func requestValues(r *http.Request) (url.Values, error) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			return nil, err
		}
		v := url.Values{}
		for k, x := range body {
			switch t := x.(type) {
			case string:
				v.Set(k, t)
			case bool:
				if t {
					v.Set(k, "1")
				}
			case nil:
			default:
				v.Set(k, fmt.Sprint(t))
			}
		}
		return v, nil
	}
	parseForm(r)
	return r.Form, nil
}
