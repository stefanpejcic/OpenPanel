package account

import (
	"net/http/httptest"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

func TestRenderResetPasswordPages(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := loginPageData{
		Title: "Reset Password", CSRFToken: "test-csrf-token", PasswordReset: "yes",
		Email: "user@example.com", PasswordStrength: 50, T: mgr.Translator("en"),
	}
	for name, page := range map[string]*web.Page{"request": requestEmailPage, "sent": linkSentPage, "set": setNewPasswordPage} {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			if err := page.Render(w, 200, data); err != nil {
				t.Fatalf("Render: %v", err)
			}
			if name != "sent" && !strings.Contains(w.Body.String(), "test-csrf-token") {
				t.Error("expected the csrf token in the form")
			}
		})
	}
}

func TestRenderAPISwaggerPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	w := httptest.NewRecorder()
	if err := apiSwaggerPage.Render(w, 200, APISwaggerPageData{LayoutData: baseLayout(mgr, "/api/docs"), Token: "tok"}); err != nil {
		t.Fatalf("Render: %v", err)
	}
}
