package web

import (
	"io"
	"strings"
	"testing"
)

func TestBaseHTMLLangFollowsLocale(t *testing.T) {
	f, err := pageFS.Open("templates/base.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, _ := io.ReadAll(f)
	if !strings.Contains(string(b), `<html lang="{{if .Lang}}{{.Lang}}{{else}}en{{end}}"`) {
		t.Fatal("base.html should take <html lang> from the resolved locale")
	}
}
