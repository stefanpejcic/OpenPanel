//go:build gallery

package web

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// go test -tags gallery ./internal/... with OPENPANEL_GALLERY_DIR set writes every page the render tests produce, for ui-tests
var galleryDir = os.Getenv("OPENPANEL_GALLERY_DIR")

var galleryEnabled = galleryDir != ""

func captureGallery(paths []string, data any, body string) {
	page, partial := "", ""
	for _, p := range paths {
		f := strings.TrimPrefix(p, "templates/")
		if f == "base.html" || strings.HasPrefix(f, "partials/") {
			continue
		}
		// modules prepend their own _shared.html, the page is the first file that isn't one of those
		if strings.HasPrefix(filepath.Base(f), "_") {
			if partial == "" {
				partial = f
			}
			continue
		}
		page = f
		break
	}
	if page == "" {
		page = partial
	}
	if page == "" {
		page = "base.html"
	}
	sum := sha256.Sum256([]byte(body))
	name := strings.NewReplacer("/", "__", ".html", "").Replace(page) + "--" + hex.EncodeToString(sum[:4]) + ".html"
	if err := os.MkdirAll(galleryDir, 0o755); err != nil {
		return
	}
	if err := os.WriteFile(filepath.Join(galleryDir, name), []byte(body), 0o644); err != nil {
		return
	}
	entry := map[string]string{"file": name, "template": page, "path": fieldString(data, "RequestPath"), "title": fieldString(data, "Title")}
	line, _ := json.Marshal(entry)
	if f, err := os.OpenFile(filepath.Join(galleryDir, "index.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		f.Write(append(line, '\n'))
		f.Close()
	}
}

func fieldString(data any, name string) string {
	v := reflect.ValueOf(data)
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	f := v.FieldByName(name)
	if f.IsValid() && f.Kind() == reflect.String {
		return f.String()
	}
	return ""
}
