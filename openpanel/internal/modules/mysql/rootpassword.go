package mysql

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/mysqlmanager"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/core/validators"
	"gist.github.com/stefanpejcic/openpanel/internal/core/webserver"
	"gist.github.com/stefanpejcic/openpanel/internal/modules/docker"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// RootPasswordPageData is mysql/root_password.html's template context.
type RootPasswordPageData struct {
	web.LayoutData
	Service       string // MYSQL_TYPE: mysql or mariadb, also the host apps reach it on
	ServerVersion string // number from VERSION(), "" when the service isn't answering
	Running       bool   // the mysqld socket is there
	RemoteOn      bool   // MYSQL_PORT is published on all interfaces
	RemotePort    string // host port from MYSQL_PORT, "" when it isn't set the usual way
}

// rootPwResult is what a root password change did, the page reads it as JSON
type rootPwResult struct {
	Kind       string `json:"kind"`  // ok, cfg (changed but my.cnf/.env not saved), down (MySQL not answering), fail (ALTER USER failed), invalid
	Step       int    `json:"step"`  // step that failed: 0 ALTER USER, 2 saving my.cnf/.env; -1 when none did
	Error      string `json:"error"` // never holds the password
	Changed    bool   `json:"changed"`
	EnvSaved   bool   `json:"env_saved"`
	MyCnfSaved bool   `json:"mycnf_saved"`
}

// swapped by tests
var (
	rootPwExec = func(ctx context.Context, userContext, query string) error {
		_, err := mysqlmanager.Exec(ctx, userContext, query, "")
		return err
	}
	rootPwSocketExists = mysqlmanager.SocketExists
	rootPwSetEnv       = func(userContext, password string) error {
		if msg := docker.SetEnvValue(userContext, "MYSQL_ROOT_PASSWORD", password); msg != "" {
			return errors.New(msg)
		}
		return nil
	}
	rootPwSetMyCnf = setMyCnfPassword
)

// rootPwBadChars can't be in the password: my.cnf holds it unquoted (# starts a comment there) and .env double-quoted (" \ $ ` are compose syntax)
const (
	rootPwBadChars    = "\"\\$`#"
	rootPwBadCharsMsg = "Spaces and the characters \" \\ $ ` # aren't allowed."
)

// rootPwHasBadChars reports whitespace, control characters or rootPwBadChars, any of which would break the my.cnf or .env line
func rootPwHasBadChars(password string) bool {
	return strings.ContainsAny(password, rootPwBadChars) || strings.IndexFunc(password, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0
}

// rootPwProblem is why password can't be used, "" when it can; the page runs the same checks before posting
func rootPwProblem(password string, threshold int) string {
	n := utf8.RuneCountInString(password)
	switch {
	case password == "":
		return "Enter a new password."
	case n < 8 || n > 30:
		return "Password must be between 8 and 30 characters."
	case rootPwHasBadChars(password):
		return rootPwBadCharsMsg
	case !validators.IsPasswordStrongEnough(password, threshold):
		return "Password does not meet the required strength."
	}
	return ""
}

// changeRootPassword sets root@dbHost and root@localhost, then saves the password to .env (phpMyAdmin) and my.cnf (the panel), MySQL keeps running
func changeRootPassword(ctx context.Context, userContext, password, dbHost string) rootPwResult {
	res := rootPwResult{Kind: "ok", Step: -1}
	hide := func(err error) string {
		msg := err.Error()
		if password != "" {
			msg = strings.ReplaceAll(strings.ReplaceAll(msg, escapeMySQLString(password), "***"), password, "***")
		}
		return msg
	}
	hosts := []string{dbHost, "localhost"}
	if dbHost == "localhost" {
		hosts = hosts[:1]
	}
	escaped := escapeMySQLString(password)
	for i, host := range hosts {
		if err := rootPwExec(ctx, userContext, "ALTER USER 'root'@'"+host+"' IDENTIFIED BY '"+escaped+"'"); err != nil {
			res.Kind, res.Step, res.Error, res.Changed = "fail", 0, hide(err), i > 0
			if i == 0 && (strings.Contains(err.Error(), "socket not found") || !rootPwSocketExists(userContext)) {
				res.Kind = "down"
			}
			return res
		}
	}
	res.Changed = true
	// ALTER USER already took effect, so a failed reload mustn't stop the save below or the panel loses access
	if err := rootPwExec(ctx, userContext, "FLUSH PRIVILEGES"); err != nil {
		log.Printf("MYSQL - root password FLUSH PRIVILEGES for %s: %s", userContext, hide(err))
	}
	res.EnvSaved = rootPwSetEnv(userContext, password) == nil
	res.MyCnfSaved = rootPwSetMyCnf(userContext, password) == nil
	if !res.EnvSaved || !res.MyCnfSaved {
		res.Kind, res.Step, res.Error = "cfg", 2, "Password changed but failed to update stored configuration."
	}
	if res.MyCnfSaved {
		mysqlmanager.InvalidatePool(userContext)
	}
	return res
}

func rootPwStatus(kind string) int {
	switch kind {
	case "invalid":
		return http.StatusBadRequest
	case "down":
		return http.StatusServiceUnavailable
	case "fail":
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

// handleRootPasswordMySQL changes the MySQL root password; the page posts with Accept: application/json and gets a rootPwResult, a plain form post gets a flash
func handleRootPasswordMySQL(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, currentUsername, userContext, err := auth.Injected(a, r)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		_ = r.ParseForm()
		password := r.Form.Get("password")
		dbHost := strings.TrimSpace(r.Form.Get("db_host"))
		if dbHost == "" {
			dbHost = "%"
		}
		threshold := validators.ClampPasswordStrength(a.Config.Get("password_strength", ""), 50)
		var res rootPwResult
		if problem := rootPwProblem(password, threshold); problem != "" {
			res = rootPwResult{Kind: "invalid", Step: -1, Error: web.Tr(a, r, problem)}
		} else if !validators.IsValidHost(dbHost) {
			res = rootPwResult{Kind: "invalid", Step: -1, Error: web.Tr(a, r, "Invalid host format.")}
		} else {
			res = changeRootPassword(ctx, userContext, password, dbHost)
			if res.Changed {
				_ = logger.RecordUserAction(a.Config, currentUsername, "changed MySQL root user password", reqip.ClientIP(r))
			}
			if res.Kind == "cfg" {
				res.Error = web.Tr(a, r, res.Error)
			}
		}

		if strings.Contains(r.Header.Get("Accept"), "application/json") {
			web.WriteJSON(w, rootPwStatus(res.Kind), res)
			return
		}
		switch res.Kind {
		case "ok":
			web.Flash(a, w, r, "success", "Successfully changed root password.")
		case "down", "fail":
			web.Flash(a, w, r, "error", web.Tr(a, r, "Error changing MySQL root password: %(error)s", "error", res.Error))
		default:
			web.Flash(a, w, r, "error", res.Error)
		}
	}

	renderRootPasswordPage(a, w, r, rootPasswordPageData(ctx, a, userContext))
}

// rootPasswordPageData only asks VERSION() when the socket is there, so a stopped service doesn't hold the page for the socket wait
func rootPasswordPageData(ctx context.Context, a *appctx.App, userContext string) RootPasswordPageData {
	data := RootPasswordPageData{Service: GetMySQLVersion(ctx, a, userContext), Running: rootPwSocketExists(userContext)}
	if data.Running {
		if rows, err := mysqlmanager.Exec(ctx, userContext, "SELECT VERSION()", ""); err == nil && len(rows) > 0 && len(rows[0]) > 0 {
			data.ServerVersion = shortServerVersion(toStringCell(rows[0][0]))
		}
	}
	if m := mysqlPortRE.FindStringSubmatch(webserver.GetEnvFileValue(userContext, "MYSQL_PORT")); m != nil {
		data.RemotePort, data.RemoteOn = m[2], m[1] == ""
	}
	return data
}

func renderRootPasswordPage(a *appctx.App, w http.ResponseWriter, r *http.Request, data RootPasswordPageData) {
	layout, _, err := web.BuildLayoutData(a, w, r, "Change Root Password")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	data.LayoutData = layout
	if err := rootPasswordPage.Render(w, http.StatusOK, data); err != nil {
		log.Printf("MYSQL - root password template render error: %v", err)
	}
}

// setMyCnfPassword rewrites the password= line in place, a sed -i swaps the inode and the container's single-file bind mount keeps showing the old file until a restart
func setMyCnfPassword(userContext, password string) error {
	path := "/home/" + userContext + "/my.cnf"
	content, err := tenantfs.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "password=") {
			lines[i] = "password=" + password
		}
	}
	return tenantfs.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600)
}
