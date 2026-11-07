package mysql

import (
	"context"
	"errors"
	"strings"
	"testing"
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
