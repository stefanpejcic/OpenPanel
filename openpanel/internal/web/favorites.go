package web

import (
	"context"
	"encoding/json"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/cache"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// Favorite is one entry of a user's favorites.json.
type Favorite struct {
	Link  string `json:"link"`
	Title string `json:"title"`
}

func FavoritesPath(username string) string {
	return "/etc/openpanel/openpanel/core/users/" + username + "/favorites.json"
}

func FavoritesCacheKey(username string) string {
	return "favorites:" + username
}

func ReadFavoritesFile(path string) ([]Favorite, error) {
	content, err := tenantfs.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var favs []Favorite
	if err := json.Unmarshal(content, &favs); err != nil {
		return nil, err
	}
	return favs, nil
}

// ReadFavoritesCached reads a user's favorites file, cached 300s.
func ReadFavoritesCached(ctx context.Context, a *appctx.App, username string) ([]Favorite, error) {
	return cache.Memoize(ctx, a.Cache, FavoritesCacheKey(username), 300*time.Second, func() ([]Favorite, error) {
		return ReadFavoritesFile(FavoritesPath(username))
	})
}

// SidebarFavorites is what the sidebar renders on first paint, empty when the module is off or there's no file yet
func SidebarFavorites(ctx context.Context, a *appctx.App, allowed map[string]bool, username string) []Favorite {
	if !allowed["favorites"] || username == "" {
		return nil
	}
	favs, _ := ReadFavoritesCached(ctx, a, username)
	return favs
}
