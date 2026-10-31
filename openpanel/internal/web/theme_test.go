package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func withTheme(t *testing.T, name string) {
	t.Helper()
	prev := Theme()
	SetTheme(name)
	t.Cleanup(func() { SetTheme(prev) })
}

func TestNormalizeTheme(t *testing.T) {
	cases := map[string]string{"": ThemeModern, "modern": ThemeModern, "classic": ThemeClassic, " Classic ": ThemeClassic, `"classic"`: ThemeClassic, "retro": ThemeModern}
	for in, want := range cases {
		if got := NormalizeTheme(in); got != want {
			t.Errorf("NormalizeTheme(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClassicPaths(t *testing.T) {
	got := classicPaths([]string{"templates/base.html", "templates/" + opIconsPartial, "templates/dashboard/dashboard.html"})
	want := []string{"templates_classic/base.html", "templates/" + opIconsPartial, "templates_classic/dashboard/dashboard.html"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("classicPaths = %v, want %v", got, want)
	}
	// php/wizard.html is new in the redesign, so the whole page stays modern
	if got := classicPaths([]string{"templates/base.html", "templates/php/wizard.html"}); got != nil {
		t.Fatalf("modern-only page got classic paths %v", got)
	}
}

func TestPageReparsesOnThemeSwitch(t *testing.T) {
	withTheme(t, ThemeModern)
	p := MustLoadPage("base.html")
	if p.classic == nil {
		t.Fatal("base.html has no classic copy")
	}
	modern := p.template()
	if p.tmplClassic {
		t.Fatal("modern theme parsed the classic copy")
	}
	SetTheme(ThemeClassic)
	classic := p.template()
	if classic == modern || !p.tmplClassic {
		t.Fatal("theme switch didn't parse the classic copy")
	}
	if p.template() != classic {
		t.Fatal("classic page wasn't cached")
	}
	SetTheme(ThemeModern)
	if p.template() == classic || p.tmplClassic {
		t.Fatal("switching back kept the classic template")
	}
}

func TestModernOnlyPageIgnoresClassicTheme(t *testing.T) {
	withTheme(t, ThemeClassic)
	p := LoadPage("php/wizard.html")
	if p.usesClassic() {
		t.Fatal("modern-only page claims a classic version")
	}
	if _, err := p.parse(p.usesClassic()); err != nil {
		t.Fatal(err)
	}
}

func TestClassicStaticURL(t *testing.T) {
	prev := assetFS
	t.Cleanup(func() {
		assetFS = prev
		assetHashes.Clear()
	})
	assetFS = fstest.MapFS{
		"js/table-sort.js":         {Data: []byte("modern")},
		"classic/js/table-sort.js": {Data: []byte("classic")},
		"js/other.js":              {Data: []byte("same")},
	}
	assetHashes.Clear()
	if got := classicStaticURL("js/table-sort.js"); !strings.HasPrefix(got, "/static/classic/js/table-sort.js?v=") {
		t.Errorf("classic copy not used: %s", got)
	}
	if got := classicStaticURL("js/other.js"); !strings.HasPrefix(got, "/static/js/other.js?v=") {
		t.Errorf("file without a classic copy: %s", got)
	}
}

func TestDashboardIconViewDefaultFollowsTheme(t *testing.T) {
	cases := []struct{ theme, cookie, want string }{
		{ThemeModern, "", "list"},
		{ThemeClassic, "", "icon"},
		{ThemeModern, "icon", "icon"},
		{ThemeClassic, "list", "list"},
		{ThemeClassic, "bogus", "icon"},
	}
	for _, c := range cases {
		withTheme(t, c.theme)
		r := httptest.NewRequest("GET", "/dashboard", nil)
		if c.cookie != "" {
			r.AddCookie(&http.Cookie{Name: DashboardIconViewCookie, Value: c.cookie})
		}
		if got := ReadDashboardIconView(r); got != c.want {
			t.Errorf("theme %s cookie %q: got %q, want %q", c.theme, c.cookie, got, c.want)
		}
	}
}
