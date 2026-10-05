package auth

import (
	"net/http"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// Injected returns the logged-in user's id, username and container context
func Injected(a *appctx.App, r *http.Request) (userID int, username, userContext string, err error) {
	userID, _ = UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		return userID, "", "", err
	}
	username, _ = data["current_username"].(string)
	userContext, _ = data["context"].(string)
	return userID, username, userContext, nil
}
