package web

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// WriteNDJSON writes one line of a streamed install/progress response and flushes it straight away
func WriteNDJSON(w http.ResponseWriter, flusher http.Flusher, canFlush bool, v map[string]any) {
	b, _ := json.Marshal(v)
	_, _ = w.Write(b)
	_, _ = w.Write([]byte("\n"))
	if canFlush {
		flusher.Flush()
	}
}

func Flash(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message string) {
	sess, _ := a.Sessions.Get(r, session.CookieName)
	flash.Add(sess, category, message)
	_ = a.Sessions.Save(r, w, sess)
}

func FlashRedirect(a *appctx.App, w http.ResponseWriter, r *http.Request, category, message, path string) {
	Flash(a, w, r, category, message)
	http.Redirect(w, r, path, http.StatusFound)
}

func FormOr(r *http.Request, key, def string) string {
	if v := r.FormValue(key); v != "" {
		return v
	}
	return def
}

func AtoiDefault(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}

// WithForm clones r as a POST carrying values as both Form and PostForm, so a UI handler reading r.FormValue sees an API's JSON body fields
func WithForm(r *http.Request, values url.Values) *http.Request {
	clone := r.Clone(r.Context())
	clone.Method = http.MethodPost
	clone.Form = values
	clone.PostForm = values
	return clone
}
