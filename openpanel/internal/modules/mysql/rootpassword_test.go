package mysql

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestRootPwProblem(t *testing.T) {
	cases := map[string]string{
		"":                        "Enter a new password.",
		"Ab1":                     "Password must be between 8 and 30 characters.",
		strings.Repeat("Ab1", 11): "Password must be between 8 and 30 characters.",
		"Abcdef12 xyz":            rootPwBadCharsMsg,
		"Abcdef12\nuser=x":        rootPwBadCharsMsg,
		`Abcdef12"x`:              rootPwBadCharsMsg,
		`Abcdef12\x`:              rootPwBadCharsMsg,
		"Abcdef12$x":              rootPwBadCharsMsg,
		"Abcdef12`x":              rootPwBadCharsMsg,
		"Abcdef12#x":              rootPwBadCharsMsg,
		"abcdefgh":                "Password does not meet the required strength.",
		"Abcdefgh12":              "",
		"Abcdef12!@%^&*()'":       "",
	}
	for pw, want := range cases {
		if got := rootPwProblem(pw, 50); got != want {
			t.Errorf("rootPwProblem(%q) = %q, want %q", pw, got, want)
		}
	}
}

// fakeRootPw swaps the exec/save hooks; failOn makes any query containing that text fail with err
func fakeRootPw(t *testing.T, failOn string, err error, socket bool, envErr, cnfErr error) *[]string {
	t.Helper()
	var queries []string
	oldExec, oldSock, oldEnv, oldCnf := rootPwExec, rootPwSocketExists, rootPwSetEnv, rootPwSetMyCnf
	t.Cleanup(func() {
		rootPwExec, rootPwSocketExists, rootPwSetEnv, rootPwSetMyCnf = oldExec, oldSock, oldEnv, oldCnf
	})
	rootPwExec = func(_ context.Context, _, q string) error {
		queries = append(queries, q)
		if failOn != "" && strings.Contains(q, failOn) {
			return err
		}
		return nil
	}
	rootPwSocketExists = func(string) bool { return socket }
	rootPwSetEnv = func(string, string) error { return envErr }
	rootPwSetMyCnf = func(string, string) error { return cnfErr }
	return &queries
}

func TestChangeRootPassword(t *testing.T) {
	ctx := context.Background()
	const pw = "Abc'def12"

	t.Run("ok", func(t *testing.T) {
		q := fakeRootPw(t, "", nil, true, nil, nil)
		res := changeRootPassword(ctx, "rp_test_ctx", pw, "%")
		if res.Kind != "ok" || res.Step != -1 || !res.Changed || !res.EnvSaved || !res.MyCnfSaved {
			t.Fatalf("got %+v", res)
		}
		want := []string{`ALTER USER 'root'@'%' IDENTIFIED BY 'Abc\'def12'`, `ALTER USER 'root'@'localhost' IDENTIFIED BY 'Abc\'def12'`, "FLUSH PRIVILEGES"}
		if strings.Join(*q, "|") != strings.Join(want, "|") {
			t.Errorf("queries = %q", *q)
		}
	})

	t.Run("localhost only once", func(t *testing.T) {
		q := fakeRootPw(t, "", nil, true, nil, nil)
		changeRootPassword(ctx, "rp_test_ctx", pw, "localhost")
		if len(*q) != 2 {
			t.Errorf("queries = %q", *q)
		}
	})

	t.Run("socket missing is down", func(t *testing.T) {
		fakeRootPw(t, "ALTER", errors.New("MySQL socket not found: /home/x/sockets/mysqld/mysqld.sock"), false, nil, nil)
		res := changeRootPassword(ctx, "rp_test_ctx", pw, "%")
		if res.Kind != "down" || res.Step != 0 || res.Changed || !strings.Contains(res.Error, "socket not found") {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("alter fails", func(t *testing.T) {
		fakeRootPw(t, "'root'@'%'", errors.New("Error 1396 (HY000): Operation ALTER USER failed for 'root'@'%'"), true, nil, nil)
		res := changeRootPassword(ctx, "rp_test_ctx", pw, "%")
		if res.Kind != "fail" || res.Step != 0 || res.Changed || res.EnvSaved || res.MyCnfSaved {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("second alter fails after first changed", func(t *testing.T) {
		fakeRootPw(t, "'root'@'localhost'", errors.New("Error 1396"), true, nil, nil)
		res := changeRootPassword(ctx, "rp_test_ctx", pw, "%")
		if res.Kind != "fail" || !res.Changed {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("error never carries the password", func(t *testing.T) {
		fakeRootPw(t, "ALTER", errors.New(`syntax error near 'Abc\'def12'' and Abc'def12`), true, nil, nil)
		res := changeRootPassword(ctx, "rp_test_ctx", pw, "%")
		if strings.Contains(res.Error, "def12") {
			t.Fatalf("password leaked: %q", res.Error)
		}
	})

	t.Run("flush failure still saves", func(t *testing.T) {
		fakeRootPw(t, "FLUSH", errors.New("denied"), true, nil, nil)
		res := changeRootPassword(ctx, "rp_test_ctx", pw, "%")
		if res.Kind != "ok" || !res.MyCnfSaved || !res.EnvSaved {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run("my.cnf not saved", func(t *testing.T) {
		fakeRootPw(t, "", nil, true, nil, errors.New("read-only"))
		res := changeRootPassword(ctx, "rp_test_ctx", pw, "%")
		if res.Kind != "cfg" || res.Step != 2 || !res.EnvSaved || res.MyCnfSaved || !res.Changed {
			t.Fatalf("got %+v", res)
		}
	})

	t.Run(".env not saved", func(t *testing.T) {
		fakeRootPw(t, "", nil, true, errors.New("read-only"), nil)
		res := changeRootPassword(ctx, "rp_test_ctx", pw, "%")
		if res.Kind != "cfg" || res.EnvSaved || !res.MyCnfSaved {
			t.Fatalf("got %+v", res)
		}
	})
}

func TestRootPwStatus(t *testing.T) {
	for kind, want := range map[string]int{"ok": 200, "cfg": 200, "invalid": 400, "down": 503, "fail": 500} {
		if got := rootPwStatus(kind); got != want {
			t.Errorf("rootPwStatus(%q) = %d, want %d", kind, got, want)
		}
	}
}

func TestRenderRootPasswordPort(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	render := func(t *testing.T, data RootPasswordPageData) string {
		t.Helper()
		data.LayoutData = baseLayout(mgr, "/mysql/root-password")
		w := httptest.NewRecorder()
		if err := rootPasswordPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		return w.Body.String()
	}

	t.Run("mariadb running, remote off", func(t *testing.T) {
		body := render(t, RootPasswordPageData{Service: "mariadb", ServerVersion: "13.0.2", Running: true, RemotePort: "32770"})
		for _, want := range []string{
			"MariaDB 13.0.2", "The database keeps running.", `name="csrf_token" value="test-csrf-token"`, `name="password"`,
			"Generate", "rp-crit", "Who signs in as root", "my.cnf", "MYSQL_ROOT_PASSWORD", `Off on this account. When on, root can also sign in from outside on port <span class="op-mono">32770</span>`,
			"Update by hand", "Change the root password?", "MariaDB keeps running. Nothing restarts", "rp-steps", "New root password", "Shown this once",
			`href="/mysql/phpmyadmin"`, `'X-CSRF-Token': csrf_token`, "const running = true",
		} {
			if !strings.Contains(body, want) {
				t.Errorf("missing %q", want)
			}
		}
		if strings.Contains(strings.ToLower(body), "restart mysql") || strings.Contains(body, "The service restarts") {
			t.Error("page still says the service restarts")
		}
	})

	t.Run("mysql stopped, remote on", func(t *testing.T) {
		body := render(t, RootPasswordPageData{Service: "mysql", RemoteOn: true, RemotePort: "3307"})
		for _, want := range []string{`On for this account. Root can sign in from outside on port <span class="op-mono">3307</span>`, "const running = false", "MySQL isn&#39;t running, so nothing changed."} {
			if !strings.Contains(body, want) {
				t.Errorf("missing %q", want)
			}
		}
	})

	t.Run("no phpmyadmin feature", func(t *testing.T) {
		data := RootPasswordPageData{Service: "mariadb", Running: true}
		data.LayoutData = baseLayout(mgr, "/mysql/root-password")
		data.UserAllowed = map[string]bool{"mysql": true}
		w := httptest.NewRecorder()
		if err := rootPasswordPage.Render(w, 200, data); err != nil {
			t.Fatalf("Render: %v", err)
		}
		body := w.Body.String()
		if strings.Contains(body, "Open phpMyAdmin") || strings.Contains(body, "Signs you in as root") {
			t.Error("phpMyAdmin shown without the feature")
		}
		if !strings.Contains(body, "Off on this account.") {
			t.Error("expected remote off without a port")
		}
	})
}
