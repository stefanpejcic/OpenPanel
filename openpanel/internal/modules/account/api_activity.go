package account

import (
	"net/http"
	"strconv"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
)

// apiActivity serves the same search/pagination view as
// handleViewActivityPage, as JSON instead of HTML.
func apiActivity(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	username, _ := data["current_username"].(string)

	searchTerm := r.URL.Query().Get("search")
	showAll := r.URL.Query().Get("show_all") == "true"
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	logContent := readActivityLog(username)
	result := paginateActivityLog(a, logContent, searchTerm, showAll, page)

	writeAPIJSON(w, http.StatusOK, map[string]any{
		"rows":        result.Rows,
		"page":        result.Page,
		"per_page":    result.ItemsPerPage,
		"total_pages": result.TotalPages,
		"total_lines": result.TotalLines,
		"show_all":    result.ShowAll,
		"search":      result.SearchTerm,
	})
}

// RegisterActivityAPI wires the /api/account/activity route onto mux.
func RegisterActivityAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "activity", "GET /api/account/activity", func(w http.ResponseWriter, r *http.Request) {
		apiActivity(a, w, r)
	})
}
