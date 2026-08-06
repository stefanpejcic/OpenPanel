package web

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// overridablePaths is the small set of static files admins can edit
// in place on disk (custom CSS/JS), plus robots.txt/security.txt which
// admins commonly replace on other panels. Everything else under static/
// is served from the embedded binary only - it's vendored third-party
// JS/CSS, not meant to be edited.
var overridablePaths = []string{"css/custom.css", "js/custom.js", "robots.txt", "security.txt"}

// StaticAssets serves static/ from the embedded binary, preferring an
// on-disk copy for the paths in overridablePaths when one exists there -
// checked once at startup and cached, not per-request.
type StaticAssets struct {
	handler   http.Handler
	CustomCSS bool
	CustomJS  bool
}

// NewStaticAssets builds a StaticAssets serving fsys (already rooted at
// the static content root, e.g. fs.Sub(assets.Static, "static")) with
// disk overrides read from overrideDir. overrideDir may be "" to disable
// overrides entirely.
func NewStaticAssets(fsys fs.FS, overrideDir string) (*StaticAssets, error) {
	overrides := map[string]string{}
	if overrideDir != "" {
		for _, rel := range overridablePaths {
			diskPath := filepath.Join(overrideDir, rel)
			if info, err := os.Stat(diskPath); err == nil && info.Size() > 1 {
				overrides[rel] = diskPath
			}
		}
	}

	embeddedHandler := http.FileServerFS(fsys)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/")
		if diskPath, ok := overrides[rel]; ok {
			http.ServeFile(w, r, diskPath)
			return
		}
		embeddedHandler.ServeHTTP(w, r)
	})

	return &StaticAssets{
		handler:   handler,
		CustomCSS: overrides["css/custom.css"] != "",
		CustomJS:  overrides["js/custom.js"] != "",
	}, nil
}

func (s *StaticAssets) Handler() http.Handler { return s.handler }

// ServeRootFile serves one static/ file at a root-level path (used for
// /robots.txt and /security.txt) - override-aware the same way Handler()
// is.
func (s *StaticAssets) ServeRootFile(rel string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := r.Clone(r.Context())
		req.URL.Path = "/" + rel
		s.handler.ServeHTTP(w, req)
	}
}
