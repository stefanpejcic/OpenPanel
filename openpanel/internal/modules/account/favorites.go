package account

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// Favorite is one entry of a user's favorites.json.
type Favorite struct {
	Link  string `json:"link"`
	Title string `json:"title"`
}

func favoritesPath(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/favorites.json"
}

func favoritesCacheKey(username string) string {
	return "favorites:" + username
}

func favoritesMax(a *appctx.App) int {
	n, err := strconv.Atoi(a.Config.Get("favorites_items", "10"))
	if err != nil {
		return 10
	}
	return n
}

func ensureFavoritesFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, []byte("[]"), 0o644)
	}
	return nil
}

func readFavoritesFile(path string) ([]Favorite, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var favs []Favorite
	if err := json.Unmarshal(content, &favs); err != nil {
		return nil, err
	}
	return favs, nil
}

// readFavoritesCached reads a user's favorites file, cached 300s.
func readFavoritesCached(ctx context.Context, a *appctx.App, path, username string) ([]Favorite, error) {
	return cache.Memoize(ctx, a.Cache, favoritesCacheKey(username), 300*time.Second, func() ([]Favorite, error) {
		return readFavoritesFile(path)
	})
}

// favoriteErr pairs an error message with the HTTP status to report it with.
type favoriteErr struct {
	Message string
	Status  int
}

// addFavorite appends a new favorite link for a user.
func addFavorite(a *appctx.App, ctx context.Context, path, username, link, title, ip string) ([]Favorite, *favoriteErr) {
	newLink := strings.TrimLeft(link, "/")
	if newLink == "" || title == "" {
		return nil, &favoriteErr{"Link and title must be provided", http.StatusBadRequest}
	}

	var favorites []Favorite
	if _, err := os.Stat(path); err == nil {
		favorites, _ = readFavoritesFile(path)
	}

	for _, fav := range favorites {
		if fav.Link == newLink {
			return favorites, nil
		}
	}
	if len(favorites) >= favoritesMax(a) {
		return nil, &favoriteErr{"Favorite list is full", http.StatusBadRequest}
	}

	favorites = append(favorites, Favorite{Link: newLink, Title: title})
	content, _ := json.Marshal(favorites)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return nil, &favoriteErr{"Failed to save favorite", http.StatusInternalServerError}
	}

	_ = a.Cache.Delete(ctx, favoritesCacheKey(username))
	_ = logger.RecordUserAction(a.Config, username, "Added "+newLink+" to Favorites", ip)

	return favorites, nil
}

// deleteFavorite removes a favorite link for a user.
func deleteFavorite(a *appctx.App, ctx context.Context, path, username, link, ip string) ([]Favorite, *favoriteErr) {
	linkToRemove := strings.TrimLeft(link, "/")
	if linkToRemove == "" {
		return nil, &favoriteErr{"No link provided", http.StatusBadRequest}
	}

	var favorites []Favorite
	if _, err := os.Stat(path); err == nil {
		favorites, _ = readFavoritesFile(path)
	}

	kept := make([]Favorite, 0, len(favorites))
	for _, fav := range favorites {
		if fav.Link != linkToRemove {
			kept = append(kept, fav)
		}
	}

	content, _ := json.Marshal(kept)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return nil, &favoriteErr{"Failed to save favorite", http.StatusInternalServerError}
	}

	_ = a.Cache.Delete(ctx, favoritesCacheKey(username))
	_ = logger.RecordUserAction(a.Config, username, "removed "+linkToRemove+" from Favorites", ip)

	return kept, nil
}

// handleJSONFavorites is the JSON API for reading/adding/removing favorites.
func handleJSONFavorites(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	path := favoritesPath(username)
	if ensureErr := ensureFavoritesFile(path); ensureErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		favorites, readErr := readFavoritesCached(ctx, a, path, username)
		if readErr != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeJSONFavorites(w, http.StatusOK, favorites)

	case http.MethodPut:
		var body struct {
			Link  string `json:"link"`
			Title string `json:"title"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		favorites, favErr := addFavorite(a, ctx, path, username, body.Link, body.Title, reqip.ClientIP(r))
		if favErr != nil {
			writeJSONFavorites(w, favErr.Status, map[string]string{"error": favErr.Message})
			return
		}
		writeJSONFavorites(w, http.StatusOK, favorites)

	case http.MethodDelete:
		var body struct {
			Link string `json:"link"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		favorites, favErr := deleteFavorite(a, ctx, path, username, body.Link, reqip.ClientIP(r))
		if favErr != nil {
			writeJSONFavorites(w, favErr.Status, map[string]string{"error": favErr.Message})
			return
		}
		writeJSONFavorites(w, http.StatusOK, favorites)
	}
}

func writeJSONFavorites(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// FavoriteRow is user/favorites.html's per-row template shape (title
// renamed to "name", matching edit_favorites_user()'s in-place rename).
type FavoriteRow struct {
	Name string
	Link string
}

// handleFavoritesPage renders the favorites management page.
func handleFavoritesPage(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)
	path := favoritesPath(username)
	if ensureErr := ensureFavoritesFile(path); ensureErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	favorites, readErr := readFavoritesCached(ctx, a, path, username)
	if readErr != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rows := make([]FavoriteRow, len(favorites))
	for i, f := range favorites {
		rows[i] = FavoriteRow{Name: f.Title, Link: f.Link}
	}

	renderFavoritesPage(a, w, r, rows)
}

// RegisterFavorites wires the favorites routes onto mux, gated behind the
// "favorites" feature flag.
func RegisterFavorites(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "favorites")(h)
	}
	mux.Handle("/json/favorites", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleJSONFavorites(a, w, r) }))
	mux.Handle("/account/favorites", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleFavoritesPage(a, w, r) }))
}
