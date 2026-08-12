package account

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gorilla/sessions"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/flash"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/session"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/core/werkzeugpw"
)

func updateEmailByID(ctx context.Context, a *appctx.App, userID int, newEmail string) error {
	_, err := a.DB.ExecContext(ctx, "UPDATE users SET email = ? WHERE id = ?", newEmail, userID)
	return err
}

// clearUserSessions deletes every "session:<id>:*" Redis key for the user,
// forcing all of that user's devices (including, per clearCurrentSession
// below, this one) to need to log in again.
func clearUserSessions(ctx context.Context, a *appctx.App, userID int) int {
	pattern := fmt.Sprintf("session:%d:*", userID)
	var keys []string
	iter := a.Cache.Raw().Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf("ACCOUNT - Failed to scan sessions from Redis: %v", err)
		return 0
	}
	if len(keys) == 0 {
		return 0
	}
	if err := a.Cache.Raw().Del(ctx, keys...).Err(); err != nil {
		log.Printf("ACCOUNT - Failed to clear sessions from Redis: %v", err)
		return 0
	}
	return len(keys)
}

func clearSessionValues(sess *sessions.Session) {
	for k := range sess.Values {
		delete(sess.Values, k)
	}
}

// updatePasswordByID rejects a weak password, otherwise hashes it, saves
// it, and clears every Redis session for the user (including - if it's the
// caller's own session, as it always is on this self-service page - the
// current one).
func updatePasswordByID(ctx context.Context, a *appctx.App, sess *sessions.Session, userID int, newPassword string) bool {
	threshold := validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)
	if !validators.IsPasswordStrongEnough(newPassword, threshold) {
		log.Printf("ACCOUNT - Rejected password update for user ID %d: does not meet the required strength", userID)
		return false
	}

	hashed, err := werkzeugpw.GeneratePasswordHash(newPassword)
	if err != nil {
		return false
	}

	if _, err := a.DB.ExecContext(ctx, "UPDATE users SET password = ? WHERE id = ?", hashed, userID); err != nil {
		log.Printf("ACCOUNT - Failed to update password for user ID %d: %v", userID, err)
		return false
	}

	clearUserSessions(ctx, a, userID)
	if uid, _ := sess.Values["user_id"].(int); uid == userID {
		clearSessionValues(sess)
	}

	return true
}

func notifySentinelPasswordChange(username string) {
	cmd := exec.Command("opencli", "sentinel", "--action=user_password",
		"--title", "User account password change",
		"--message", "Password for user account '"+username+"' has been changed.")
	_ = cmd.Start()
}

// handleAccountSettings implements email/password/username self-service,
// mounted at both /settings and /account.
func handleAccountSettings(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	userID, _ := auth.UserID(r)
	ctx := r.Context()
	data, err := a.InjectData(ctx, userID)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	currentUsername, _ := data["current_username"].(string)
	currentEmail, _ := data["current_email"].(string)

	permitUsernameChange := a.Config.Get("permit_username_change_by_user", "no")

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		newEmail := r.Form.Get("email")
		newPassword := r.Form.Get("password")
		confirmPassword := r.Form.Get("confirm_password")

		sess, _ := a.Sessions.Get(r, session.CookieName)
		ip := reqip.ClientIP(r)

		if newPassword != "" && newPassword == confirmPassword {
			if !updatePasswordByID(ctx, a, sess, userID, newPassword) {
				flash.Add(sess, "error", "Password does not meet the required strength.")
			} else {
				message := "Password changed for account " + currentUsername + "\n" +
					"Password for account <b>" + currentUsername + "</b> has been changed."
				prettyMessage := "Password for account " + currentUsername + " has been changed successfully."
				checkIfUserShouldBeNotified(a, ctx, userID, currentUsername, "notify_password_change", message)
				flash.Add(sess, "success", prettyMessage)
				_ = logger.RecordUserAction(a.Config, currentUsername, "changed password", ip)
				notifySentinelPasswordChange(currentUsername)
			}
		}

		if newEmail != currentEmail {
			message := "Email address changed for account " + currentUsername + "\n" +
				"Email address for account <b>" + currentUsername + "</b> has been changed to: <b>" + newEmail + "</b>."
			checkIfUserShouldBeNotified(a, ctx, userID, currentUsername, "notify_contact_address_change", message)
			_ = updateEmailByID(ctx, a, userID, newEmail)
			a.Cache.Delete(ctx, "get_user_details_with_plan:"+strconv.Itoa(userID))

			prettyMessage := fmt.Sprintf("Email address for account %s has been changed successfully from %s to %s.", currentUsername, currentEmail, newEmail)
			flash.Add(sess, "success", prettyMessage)
			_ = logger.RecordUserAction(a.Config, currentUsername, "changed email address to "+newEmail, ip)
			_ = a.Sessions.Save(r, w, sess)
			http.Redirect(w, r, "/account", http.StatusFound)
			return
		}

		if permitUsernameChange == "yes" {
			newUsername := r.Form.Get("username")
			if newUsername != "" && newUsername != currentUsername {
				if !validators.IsValidPanelUsername(newUsername) {
					flash.Add(sess, "error", "Username must be 3-20 characters, letters and numbers only.")
					_ = a.Sessions.Save(r, w, sess)
					http.Redirect(w, r, "/account", http.StatusFound)
					return
				}
				out, runErr := exec.CommandContext(ctx, "opencli", "user-rename", currentUsername, newUsername).CombinedOutput()
				output := string(out)
				if runErr == nil && strings.Contains(strings.ToLower(output), "successfully") {
					flash.Add(sess, "success", fmt.Sprintf("Username has been changed successfully from %s to %s.", currentUsername, newUsername))

					message := "Username " + currentUsername + " changed\n" +
						"OpenPanel username changed from <b>" + currentUsername + "</b> to <b>" + newUsername + "</b>."
					checkIfUserShouldBeNotified(a, ctx, userID, newUsername, "notify_username_change", message)
					a.Cache.Delete(ctx, "get_user_details_with_plan:"+strconv.Itoa(userID))
					_ = logger.RecordUserAction(a.Config, newUsername, "changed username from "+currentUsername+" to "+newUsername, ip)

					_ = a.Sessions.Save(r, w, sess)
					http.Redirect(w, r, "/account", http.StatusFound)
					return
				}

				if runErr != nil {
					log.Printf("ACCOUNT - Error changing username: %v", runErr)
					flash.Add(sess, "error", "Failed to change username.")
				} else {
					flash.Add(sess, "error", output)
				}
			}
		}

		_ = a.Sessions.Save(r, w, sess)
	}

	renderAccountPage(a, w, r, permitUsernameChange)
}

// RegisterSettings wires the account self-service routes onto mux, gated
// behind the "account" feature flag (unlike login/logout, which Register
// in login.go wires unconditionally).
func RegisterSettings(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "account")(h)
	}
	mux.Handle("/settings", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAccountSettings(a, w, r) }))
	mux.Handle("/account", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAccountSettings(a, w, r) }))
}
