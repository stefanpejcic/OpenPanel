package filemanager

import (
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSizeFolderDecodesOnce(t *testing.T) {
	// a second decode used to turn c++ into "c  " and a%41 into aA
	for folder, want := range map[string]string{
		"c++":                     "c++",
		"100%/a%41":               "100%/a%41",
		"my site/a&b#1":           "my site/a&b#1",
		"/var/www/html/site/blog": "site/blog",
		"/site":                   "site",
	} {
		r := httptest.NewRequest("GET", "/json/directory-size?folder="+url.QueryEscape(folder), nil)
		if got := sizeFolder(r); got != want {
			t.Errorf("sizeFolder(%q) = %q, want %q", folder, got, want)
		}
	}
}
