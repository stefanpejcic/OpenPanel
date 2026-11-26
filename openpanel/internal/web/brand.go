package web

import (
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// brandFileRe matches the only names OpenAdmin's logo upload writes, anything else in the dir is never served
var brandFileRe = regexp.MustCompile(`^(logo|logo_dark)\.(png|jpg|webp|gif|svg)$`)

var brandTypes = map[string]string{
	".png": "image/png", ".jpg": "image/jpeg", ".webp": "image/webp", ".gif": "image/gif", ".svg": "image/svg+xml",
}

// ServeBrandFile serves /brand/{file}, the logos uploaded in OpenAdmin > Settings > OpenPanel. It's public because the login page shows them.
func ServeBrandFile(dir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("file")
		if !brandFileRe.MatchString(name) {
			http.NotFound(w, r)
			return
		}
		path := filepath.Join(dir, name)
		// a symlink could point anywhere on the host, only plain files count
		if info, err := os.Lstat(path); err != nil || !info.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}
		h := w.Header()
		h.Set("Content-Type", brandTypes[filepath.Ext(name)])
		h.Set("X-Content-Type-Options", "nosniff")
		// an svg opened on its own would run its scripts on the panel's origin, sandbox stops that (an <img> never runs them)
		h.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; sandbox")
		// OpenAdmin saves the url with ?v=<content hash>, so a new upload gets a new url
		if r.URL.Query().Get("v") != "" {
			h.Set("Cache-Control", "public, max-age=31536000, immutable")
		} else {
			h.Set("Cache-Control", "no-cache")
		}
		http.ServeContent(w, r, name, info.ModTime(), f)
	}
}
