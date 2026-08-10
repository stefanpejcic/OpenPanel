package account

import (
	"encoding/json"
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// RegisterFavoritesAPI wires the /api/account/favorites routes onto mux,
// gated behind the "favorites" feature flag. Reuses the same
// read/add/deleteFavorite helpers as the /json/favorites page endpoint
// (favorites.go) - this just gives it a stable, documented /api/ home with
// REST-conventional verbs (POST to add, DELETE to remove) instead of
// /json/favorites's PUT/DELETE pair.
func RegisterFavoritesAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "favorites", "GET /api/account/favorites", func(w http.ResponseWriter, r *http.Request) { apiFavoritesList(a, w, r) })
	apiregistry.Handle(mux, a, "favorites", "POST /api/account/favorites", func(w http.ResponseWriter, r *http.Request) { apiFavoritesAdd(a, w, r) })
	apiregistry.Handle(mux, a, "favorites", "DELETE /api/account/favorites", func(w http.ResponseWriter, r *http.Request) { apiFavoritesRemove(a, w, r) })
}

func apiFavoritesUserPath(a *appctx.App, r *http.Request) (username, path string, ok bool) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return "", "", false
	}
	username, _ = data["current_username"].(string)
	path = favoritesPath(username)
	return username, path, ensureFavoritesFile(path) == nil
}

// apiFavoritesList returns the caller's favorite links.
func apiFavoritesList(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, path, ok := apiFavoritesUserPath(a, r)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	favorites, readErr := readFavoritesCached(r.Context(), a, path, username)
	if readErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"favorites": favorites})
}

// apiFavoritesAdd adds a new favorite link for the caller.
func apiFavoritesAdd(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, path, ok := apiFavoritesUserPath(a, r)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var body struct {
		Link  string `json:"link"`
		Title string `json:"title"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	favorites, favErr := addFavorite(a, r.Context(), path, username, body.Link, body.Title, reqip.ClientIP(r))
	if favErr != nil {
		writeAPIAccountJSON(w, favErr.Status, map[string]string{"error": favErr.Message})
		return
	}
	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"favorites": favorites})
}

// apiFavoritesRemove removes a favorite link for the caller.
func apiFavoritesRemove(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	username, path, ok := apiFavoritesUserPath(a, r)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	var body struct {
		Link string `json:"link"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	favorites, favErr := deleteFavorite(a, r.Context(), path, username, body.Link, reqip.ClientIP(r))
	if favErr != nil {
		writeAPIAccountJSON(w, favErr.Status, map[string]string{"error": favErr.Message})
		return
	}
	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"favorites": favorites})
}
