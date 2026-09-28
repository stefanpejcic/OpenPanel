package account

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/sessions"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestIsCommonPasswordExitCodes(t *testing.T) {
	orig := weakpassScript
	t.Cleanup(func() { weakpassScript = orig })

	script := filepath.Join(t.TempDir(), "weakpass.sh")
	os.WriteFile(script, []byte("read -r p; [ \"$p\" = password ] && exit 1; exit 0\n"), 0644)
	weakpassScript = script

	if !isCommonPassword(context.Background(), "password") {
		t.Error("expected exit 1 to mean a common password")
	}
	if isCommonPassword(context.Background(), "Xk9#pLq2!vRt") {
		t.Error("expected exit 0 to mean the password is fine")
	}

	weakpassScript = filepath.Join(t.TempDir(), "missing.sh")
	if isCommonPassword(context.Background(), "password") {
		t.Error("expected a missing script to let the password through")
	}
}

func TestUpdatePasswordByIDRejectsCommonPassword(t *testing.T) {
	orig := isCommonPassword
	t.Cleanup(func() { isCommonPassword = orig })
	isCommonPassword = func(context.Context, string) bool { return true }

	// no DB on the App, so this would panic if the check didn't stop it first
	a := &appctx.App{Config: config.Config{"password_strength": "50"}}
	sess := &sessions.Session{Values: map[interface{}]interface{}{}}
	if got := updatePasswordByID(context.Background(), a, sess, 1, "Password123"); got != errPasswordCommon {
		t.Errorf("expected %q, got %q", errPasswordCommon, got)
	}
	if got := updatePasswordByID(context.Background(), a, sess, 1, "abc"); got != errPasswordWeak {
		t.Errorf("expected the strength check to run first, got %q", got)
	}
}
