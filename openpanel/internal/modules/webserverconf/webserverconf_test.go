package webserverconf

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestLookupWebserverConf(t *testing.T) {
	cases := []struct {
		webServer   string
		wantFile    string
		wantService string
	}{
		{"nginx", "nginx.conf", "nginx"},
		{"apache", "httpd.conf", "apache"},
		{"openresty", "openresty.conf", "openresty"},
		{"openlitespeed", "openlitespeed.conf", "openlitespeed"},
		{"litespeed", "openlitespeed.conf", "litespeed"},
	}
	for _, c := range cases {
		entry := lookupWebserverConf(c.webServer)
		if entry.ConfFile != c.wantFile || entry.ServiceName != c.wantService {
			t.Errorf("lookupWebserverConf(%q) = %+v, want file=%q service=%q", c.webServer, entry, c.wantFile, c.wantService)
		}
	}
}

func TestLookupWebserverConfUnknown(t *testing.T) {
	entry := lookupWebserverConf("something-unsupported")
	if entry.ConfFile != "" || entry.ServiceName != "" {
		t.Errorf("expected empty ConfFile/ServiceName for unknown web server, got %+v", entry)
	}
	if entry.PageTitle != "Web Server Configuration Editor" {
		t.Errorf("PageTitle = %q", entry.PageTitle)
	}
}

func TestFileExists(t *testing.T) {
	if fileExists("/nonexistent/path/does/not/exist") {
		t.Error("expected false for a nonexistent path")
	}
	if !fileExists("webserverconf_test.go") {
		t.Error("expected true for this test file")
	}
}

func TestEditorContentNormalizesCRLF(t *testing.T) {
	r := httptest.NewRequest("POST", "/server/webserver_conf", strings.NewReader("editor_content="+url.QueryEscape("Listen 80\r\n<IfModule x>\r\n</IfModule>\r\n")))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_ = r.ParseForm()
	if got := editorContent(r); got != "Listen 80\n<IfModule x>\n</IfModule>\n" {
		t.Fatalf("got %q", got)
	}
}
