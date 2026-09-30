package web

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// read once at startup, OpenAdmin flags a restart when these change
var customHeaderHTML, customFooterHTML template.HTML

// LoadCustomCode reads in_header.html/in_footer.html from dir, "" clears them
func LoadCustomCode(dir string) {
	customHeaderHTML, customFooterHTML = "", ""
	if dir == "" {
		return
	}
	customHeaderHTML = readCustomHTML(filepath.Join(dir, "in_header.html"))
	customFooterHTML = readCustomHTML(filepath.Join(dir, "in_footer.html"))
}

func readCustomHTML(path string) template.HTML {
	raw, err := os.ReadFile(path)
	if err != nil || strings.TrimSpace(string(raw)) == "" {
		return ""
	}
	return template.HTML(raw) //nolint:gosec // admin-authored html, editing is Enterprise + admin only
}
