package account

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterLocaleAPI wires the /api/account/language routes onto mux,
// gated behind the "locale" feature flag.
func RegisterLocaleAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "locale", "GET /api/account/language", func(w http.ResponseWriter, r *http.Request) { apiLocaleGet(a, w, r) })
	apiregistry.Handle(mux, a, "locale", "PUT /api/account/language", func(w http.ResponseWriter, r *http.Request) { apiLocaleUpdate(a, w, r) })
}

// apiLocaleGet returns the caller's current UI language and the available
// options.
func apiLocaleGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	userContext, _ := data["context"].(string)

	locales := a.I18n.AvailableLocales(ctx)
	current := i18n.UserLocale(userContext)
	if current != "" && !contains(locales, current) {
		current = ""
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"locales": locales, "current": current})
}

// apiLocaleUpdate sets the caller's preferred UI language.
func apiLocaleUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	userContext, _ := data["context"].(string)

	var body struct {
		Locale string `json:"locale"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Locale = r.Form.Get("locale")
	}
	if body.Locale == "" {
		writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "locale is required"})
		return
	}
	if !contains(a.I18n.AvailableLocales(ctx), body.Locale) {
		writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "Unknown locale"})
		return
	}

	userLocaleFile := filepath.Join("/home", userContext, "locale")
	if mkdirErr := os.MkdirAll(filepath.Dir(userLocaleFile), 0o755); mkdirErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if writeErr := os.WriteFile(userLocaleFile, []byte(body.Locale), 0o644); writeErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	_ = logger.RecordUserAction(a.Config, username, "changed language to "+body.Locale+" via API", reqip.ClientIP(r))
	writeAPIAccountJSON(w, http.StatusOK, map[string]string{"locale": body.Locale})
}
