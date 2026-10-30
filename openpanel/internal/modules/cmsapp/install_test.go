package cmsapp

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
	"gist.github.com/stefanpejcic/openpanel/internal/web/webtest"
)

func TestRenderInstallPageEnforcedPrefix(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	app := New(App{Slug: "joomla", Name: "Joomla"})
	allowed := map[string]bool{"dashboard": true, "joomla": true}
	data := InstallPageData{
		LayoutData: web.LayoutData{
			Title: "Test", CSRFToken: "t", PanelDir: "ltr", UserAllowed: allowed, UserAllowedJSON: web.UserAllowedList(allowed),
			NavGroups: web.BuildClassicSidebarNav(allowed, nil, "/joomla/install"), RequestPath: "/joomla/install",
			PasswordStrength: 50, MySQLPrefix: "john_", T: mgr.Translator("en"),
		},
		Domains: []appctx.Domain{{DomainID: 1, Docroot: "/var/www/html/example.com", DomainURL: "example.com"}},
		App:     web.AppNames{Slug: "joomla", Name: "Joomla"},
	}
	w := httptest.NewRecorder()
	if err := app.installPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if !strings.Contains(w.Body.String(), `mysqlPrefix: "john_"`) {
		t.Error("expected the john_ prefix next to the db fields")
	}
}

func TestRenderAllInstallPages(t *testing.T) {
	webtest.ModernOnly(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	// every app's wizard posts the form fields its install.go reads
	fields := map[string][]string{
		"joomla": {"'site_name'", "'joomla_version'", "'admin_name'", "K.db('joomla')"}, "drupal": {"'site_name'", "'drupal_version'", "K.db('drupal')"},
		"flarum": {"'site_name'", "'flarum_version'", "min: '8.1'"}, "matomo": {"'matomo_version'", "name: 'admin_login'"}, "mediawiki": {"'site_name'", "'mediawiki_version'"},
		"moodle": {"'moodle_version'"}, "nextcloud": {"'nextcloud_version'"}, "ojs": {"'ojs_version'"}, "opencart": {"'opencart_version'"},
		"phpbb": {"'board_name'", "'board_description'", "'phpbb_version'"}, "prestashop": {"'prestashop_version'", "'admin_firstname'", "'admin_lastname'"},
		"dokuwiki": {"'site_title'", "name: 'admin_user'", "min: '7.4'"}, "sofawiki": {"max: '7.4'"}, "tinyfilemanager": {"K.user()", "K.pass()"}, "tinyphotogallery": {"fields: []"},
	}
	for _, slug := range []string{"joomla", "dokuwiki", "drupal", "flarum", "matomo", "mediawiki", "moodle", "nextcloud", "ojs", "opencart", "phpbb", "prestashop", "sofawiki", "tinyfilemanager", "tinyphotogallery"} {
		t.Run(slug, func(t *testing.T) {
			app := New(App{Slug: slug, Name: slug})
			allowed := map[string]bool{"dashboard": true, slug: true}
			data := InstallPageData{
				LayoutData: web.LayoutData{
					Title: "Test", BrandName: "Test Panel", CSRFToken: "t", PanelDir: "ltr", UserAllowed: allowed, UserAllowedJSON: web.UserAllowedList(allowed),
					NavGroups: web.BuildClassicSidebarNav(allowed, nil, "/"+slug+"/install"), RequestPath: "/" + slug + "/install",
					PasswordStrength: 50, MySQLPrefix: "john_", CurrentUsername: "testuser", T: mgr.Translator("en"),
				},
				Domains: []appctx.Domain{{DomainID: 1, Docroot: "/var/www/html/example.com", DomainURL: "example.com"}},
				App:     web.AppNames{Slug: template.JS(slug), Name: template.JS(slug)},
			}
			w := httptest.NewRecorder()
			if err := app.installPage.Render(w, 200, data); err != nil {
				t.Fatalf("Render: %v", err)
			}
			body := w.Body.String()
			for _, want := range append([]string{`<div class="op-content">`, `x-data="opInstallWizard(window.opIWConfig)"`, `"DomainURL":"example.com"`, `endpoint: "/` + slug + `/install"`, `name="csrf_token"`, "images/apps/" + slug + ".svg"}, fields[slug]...) {
				if !strings.Contains(body, want) {
					t.Errorf("expected %s in the install page", want)
				}
			}
			if strings.Contains(body, "{{") {
				t.Error("unrendered template action in the page")
			}
		})
	}
}

func TestRenderInstallPageNoDomains(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	app := New(App{Slug: "nextcloud", Name: "Nextcloud"})
	allowed := map[string]bool{"dashboard": true, "nextcloud": true}
	data := InstallPageData{
		LayoutData: web.LayoutData{
			Title: "Test", CSRFToken: "t", PanelDir: "ltr", UserAllowed: allowed, UserAllowedJSON: web.UserAllowedList(allowed),
			NavGroups: web.BuildClassicSidebarNav(allowed, nil, "/nextcloud/install"), RequestPath: "/nextcloud/install", T: mgr.Translator("en"),
		},
		App: web.AppNames{Slug: "nextcloud", Name: "Nextcloud"},
	}
	w := httptest.NewRecorder()
	if err := app.installPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	if body := w.Body.String(); !strings.Contains(body, "/domains/new") || strings.Contains(body, "opInstallWizard(") {
		t.Error("expected the no-domains card instead of the wizard")
	}
}
