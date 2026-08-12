package account

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/sessions"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/apiregistry"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
)

// RegisterAccountAPI wires the /api/account routes onto mux.
func RegisterAccountAPI(mux *http.ServeMux, a *appctx.App) {
	apiregistry.Handle(mux, a, "account", "GET /api/account", func(w http.ResponseWriter, r *http.Request) { apiAccountGet(a, w, r) })
	apiregistry.Handle(mux, a, "account", "PATCH /api/account", func(w http.ResponseWriter, r *http.Request) { apiAccountUpdate(a, w, r) })
	apiregistry.Handle(mux, a, "account", "GET /api/account/sessions", func(w http.ResponseWriter, r *http.Request) { apiAccountSessions(a, w, r) })
	apiregistry.Handle(mux, a, "account", "DELETE /api/account/sessions/{session_token}", func(w http.ResponseWriter, r *http.Request) { apiAccountSessionTerminate(a, w, r) })
}

func writeAPIAccountJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// apiAccountGet returns the caller's account profile.
func apiAccountGet(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeAPIAccountJSON(w, http.StatusOK, map[string]any{
		"username":               data["current_username"],
		"email":                  data["current_email"],
		"plan_id":                data["hosting_plan"],
		"context":                data["context"],
		"permit_username_change": a.Config.Get("permit_username_change_by_user", "no") == "yes",
	})
}

// apiAccountUpdate applies any of email/password/username changes present
// in the request body.
func apiAccountUpdate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _ := auth.UserID(r)
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	currentUsername, _ := data["current_username"].(string)
	currentEmail, _ := data["current_email"].(string)

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Username string `json:"username"`
	}
	if jsonErr := json.NewDecoder(r.Body).Decode(&body); jsonErr != nil {
		_ = r.ParseForm()
		body.Email = r.Form.Get("email")
		body.Password = r.Form.Get("password")
		body.Username = r.Form.Get("username")
	}
	newEmail := strings.TrimSpace(body.Email)
	newPassword := strings.TrimSpace(body.Password)
	newUsername := strings.TrimSpace(body.Username)

	ip := reqip.ClientIP(r)
	var actions []string

	if newPassword != "" {
		sess := &sessions.Session{Values: map[interface{}]interface{}{}}
		if !updatePasswordByID(ctx, a, sess, userID, newPassword) {
			writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "Password does not meet the required strength"})
			return
		}
		_ = logger.RecordUserAction(a.Config, currentUsername, "changed password via API", ip)
		actions = append(actions, "password updated")
	}

	if newEmail != "" && newEmail != currentEmail {
		if updateErr := updateEmailByID(ctx, a, userID, newEmail); updateErr != nil {
			writeAPIAccountJSON(w, http.StatusInternalServerError, map[string]string{"error": updateErr.Error()})
			return
		}
		a.Cache.Delete(ctx, "get_user_details_with_plan:"+strconv.Itoa(userID))
		_ = logger.RecordUserAction(a.Config, currentUsername, "changed email to "+newEmail+" via API", ip)
		actions = append(actions, "email updated to "+newEmail)
	}

	if newUsername != "" && newUsername != currentUsername {
		if a.Config.Get("permit_username_change_by_user", "no") != "yes" {
			writeAPIAccountJSON(w, http.StatusForbidden, map[string]string{"error": "Username change is not permitted on this server"})
			return
		}
		if !validators.IsValidPanelUsername(newUsername) {
			writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "Username must be 3-20 characters, letters and numbers only"})
			return
		}
		out, runErr := exec.CommandContext(ctx, "opencli", "user-rename", currentUsername, newUsername).CombinedOutput()
		output := string(out)
		if runErr != nil {
			writeAPIAccountJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to rename: " + output})
			return
		}
		if strings.Contains(strings.ToLower(output), "successfully") {
			a.Cache.Delete(ctx, "get_user_details_with_plan:"+strconv.Itoa(userID))
			_ = logger.RecordUserAction(a.Config, newUsername, "renamed from "+currentUsername+" to "+newUsername, ip)
			actions = append(actions, "username changed to "+newUsername)
		} else {
			msg := strings.TrimSpace(output)
			if msg == "" {
				msg = "Username change failed"
			}
			writeAPIAccountJSON(w, http.StatusInternalServerError, map[string]string{"error": msg})
			return
		}
	}

	if len(actions) == 0 {
		writeAPIAccountJSON(w, http.StatusBadRequest, map[string]string{"error": "Provide at least one of: email, password, username"})
		return
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"message": "Account updated", "actions": actions})
}

// apiAccountSessions lists the caller's active sessions from Redis.
func apiAccountSessions(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()

	sessionsList := []map[string]string{}
	pattern := fmt.Sprintf("session:%d:*", userID)
	iter := a.Cache.Raw().Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		data, hgetErr := a.Cache.Raw().HGetAll(ctx, key).Result()
		if hgetErr != nil || len(data) == 0 {
			continue
		}

		ttl, _ := a.Cache.Raw().TTL(ctx, key).Result()
		parts := strings.Split(key, ":")
		token := parts[len(parts)-1]

		expiresIn := "Expiring..."
		if ttl > 0 {
			expiresIn = fmt.Sprintf("%dm", int64(ttl/time.Minute))
		}

		ip := data["ip_address"]
		if ip == "" {
			ip = "Unknown"
		}
		createdAt := data["created_at"]
		if createdAt == "" {
			createdAt = "N/A"
		}

		sessionsList = append(sessionsList, map[string]string{
			"session_token": token, "ip_address": ip, "created_at": createdAt, "expires_in": expiresIn,
		})
	}
	if err := iter.Err(); err != nil {
		writeAPIAccountJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list sessions: " + err.Error()})
		return
	}

	writeAPIAccountJSON(w, http.StatusOK, map[string]any{"sessions": sessionsList})
}

// apiAccountSessionTerminate deletes a single session record by token.
func apiAccountSessionTerminate(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	sessionToken := r.PathValue("session_token")

	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	currentUsername, _ := data["current_username"].(string)

	sessionKey := fmt.Sprintf("session:%d:%s", userID, sessionToken)
	n, delErr := a.Cache.Raw().Del(ctx, sessionKey).Result()
	if delErr != nil {
		writeAPIAccountJSON(w, http.StatusInternalServerError, map[string]string{"error": "Redis error: " + delErr.Error()})
		return
	}
	if n == 0 {
		writeAPIAccountJSON(w, http.StatusNotFound, map[string]string{"error": "Session not found or already expired"})
		return
	}

	_ = logger.RecordUserAction(a.Config, currentUsername, "terminated session "+sessionToken+" via API", reqip.ClientIP(r))
	writeAPIAccountJSON(w, http.StatusOK, map[string]string{"message": "Session terminated"})
}
