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
	orig := opencliBin
	t.Cleanup(func() { opencliBin = orig })

	bin := filepath.Join(t.TempDir(), "opencli")
	os.WriteFile(bin, []byte("#!/bin/sh\n[ \"$1\" = weakpass ] || exit 2\nread -r p; [ \"$p\" = password ] && exit 1; exit 0\n"), 0755)
	opencliBin = bin

	if !isCommonPassword(context.Background(), "password") {
		t.Error("expected exit 1 to mean a common password")
	}
	if isCommonPassword(context.Background(), "Xk9#pLq2!vRt") {
		t.Error("expected exit 0 to mean the password is fine")
	}

	opencliBin = filepath.Join(t.TempDir(), "missing")
	if isCommonPassword(context.Background(), "password") {
		t.Error("expected a missing opencli to let the password through")
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
