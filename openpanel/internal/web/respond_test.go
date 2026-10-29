package web

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestFormTextNormalizesCRLF(t *testing.T) {
	body := "content=" + url.QueryEscape("Listen 80\r\n<IfModule x>\r\n</IfModule>\r\n")
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if got := FormText(r, "content"); got != "Listen 80\n<IfModule x>\n</IfModule>\n" {
		t.Fatalf("got %q", got)
	}
	if got := FormText(r, "missing"); got != "" {
		t.Fatalf("missing field got %q", got)
	}
}
