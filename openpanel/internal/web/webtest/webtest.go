// Package webtest holds helpers shared by the render tests.
package webtest

import (
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// ModernOnly skips a test that asserts redesign markup when the suite runs with OPENPANEL_UI_THEME=classic
func ModernOnly(t testing.TB) {
	t.Helper()
	if web.Theme() == web.ThemeClassic {
		t.Skip("asserts modern markup, skipped under the classic theme")
	}
}

// Classic renders with the classic templates until the test ends, for tests that assert pre-redesign markup
func Classic(t testing.TB) {
	t.Helper()
	prev := web.Theme()
	web.SetTheme(web.ThemeClassic)
	t.Cleanup(func() { web.SetTheme(prev) })
}
