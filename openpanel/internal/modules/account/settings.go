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
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func updateEmailByID(ctx context.Context, a *appctx.App, userID int, newEmail string) error {
	_, err := a.DB.ExecContext(ctx, "UPDATE users SET email = ? WHERE id = ?", newEmail, userID)
	return err
}

// clearUserSessions deletes every "session:<id>:*" Redis key for the user, forcing all of that user's devices (including this one) to log in again
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

// updatePasswordByID rejects a weak or common password, otherwise hashes it, saves it, and clears every Redis session for the user, including the caller's own current one. Returns one of the errPassword* codes, empty on success.
func updatePasswordByID(ctx context.Context, a *appctx.App, sess *sessions.Session, userID int, newPassword string) string {
	threshold := validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)
	if !validators.IsPasswordStrongEnough(newPassword, threshold) {
		log.Printf("ACCOUNT - Rejected password update for user ID %d: does not meet the required strength", userID)
		return errPasswordWeak
	}
	if isCommonPassword(ctx, newPassword) {
		log.Printf("ACCOUNT - Rejected password update for user ID %d: password is in the common passwords list", userID)
		return errPasswordCommon
	}

	hashed, err := werkzeugpw.GeneratePasswordHash(newPassword)
	if err != nil {
		return errPasswordSave
	}

	if _, err := a.DB.ExecContext(ctx, "UPDATE users SET password = ? WHERE id = ?", hashed, userID); err != nil {
		log.Printf("ACCOUNT - Failed to update password for user ID %d: %v", userID, err)
		return errPasswordSave
	}

	clearUserSessions(ctx, a, userID)
	if uid, _ := sess.Values["user_id"].(int); uid == userID {
		clearSessionValues(sess)
	}

	return ""
}

const (
	errPasswordWeak   = "weak"
	errPasswordCommon = "common"
	errPasswordSave   = "save"
)

// passwordErrorText turns an updatePasswordByID code into a translated message, literals inline so the catalog sync finds them
func passwordErrorText(a *appctx.App, r *http.Request, code string) string {
	switch code {
	case errPasswordWeak:
		return web.Tr(a, r, "Password does not meet the required strength.")
	case errPasswordCommon:
		return web.Tr(a, r, "Password is too common, please choose a different one.")
	}
	return web.Tr(a, r, "Failed to update password.")
}

func notifySentinelPasswordChange(username string) {
	cmd := exec.Command("opencli", "sentinel", "--action=user_password",
		"--title", "User account password change",
		"--message", "Password for user account '"+username+"' has been changed.")
	if err := cmd.Start(); err == nil {
		go func() { _ = cmd.Wait() }() // reap, avoid a zombie process
	}
}

// handleAccountSettings implements email/password/username self-service, mounted at both /settings and /account
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
			if errMsg := updatePasswordByID(ctx, a, sess, userID, newPassword); errMsg != "" {
				flash.Add(sess, "error", passwordErrorText(a, r, errMsg))
			} else {
				message := securityEmail(a, r, "Password changed for account "+currentUsername,
					"The password for account "+currentUsername+" was changed from the OpenPanel interface. All other sessions stay logged in until they expire.",
					"If you didn't change it, reset your password right away or contact your hosting provider.")
				checkIfUserShouldBeNotified(a, ctx, userID, currentUsername, "notify_password_change", message)
				flash.Add(sess, "success", "Password has been changed successfully.")
				_ = logger.RecordUserAction(a.Config, currentUsername, "changed password", ip)
				notifySentinelPasswordChange(currentUsername)
			}
		}

		if newEmail != currentEmail {
			message := securityEmail(a, r, "Email address changed for account "+currentUsername,
				"The contact email address for account "+currentUsername+" was changed from "+currentEmail+" to "+newEmail+". This is the last notification sent to this address, new ones go to "+newEmail+".",
				"If you didn't change it, log in and set your email address back, then change your password.")
			checkIfUserShouldBeNotified(a, ctx, userID, currentUsername, "notify_contact_address_change", message)
			_ = updateEmailByID(ctx, a, userID, newEmail)
			a.Cache.Delete(ctx, "get_user_details_with_plan:"+strconv.Itoa(userID))

			flash.Add(sess, "success", web.Tr(a, r, "Email address has been changed successfully from %(old_email)s to %(new_email)s.", "old_email", currentEmail, "new_email", newEmail))
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
					flash.Add(sess, "success", web.Tr(a, r, "Username has been changed successfully from %(current_username)s to %(new_username)s.", "current_username", currentUsername, "new_username", newUsername))

					message := securityEmail(a, r, "Username changed for account "+newUsername,
						"The username was changed from "+currentUsername+" to "+newUsername+". Use "+newUsername+" to log in from now on.",
						"If you didn't change it, contact your hosting provider right away.")
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

// RegisterSettings wires the account self-service routes onto mux, gated behind "account" - unlike login/logout, which Register in login.go wires unconditionally
func RegisterSettings(mux *http.ServeMux, a *appctx.App) {
	requireLogin := func(h http.HandlerFunc) http.Handler {
		return auth.RequireLogin(a, "account")(h)
	}
	mux.Handle("/settings", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAccountSettings(a, w, r) }))
	mux.Handle("/account", requireLogin(func(w http.ResponseWriter, r *http.Request) { handleAccountSettings(a, w, r) }))
}
