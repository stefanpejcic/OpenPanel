package web

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// overridablePaths maps a served static path to its file name in the override dir, everything else is embedded only
var overridablePaths = map[string]string{"css/custom.css": "custom.css", "js/custom.js": "custom.js", "robots.txt": "robots.txt", "security.txt": "security.txt"}

// customCodePaths are only overridden when custom code is enabled (Enterprise)
var customCodePaths = map[string]bool{"css/custom.css": true, "js/custom.js": true}

// StaticAssets serves static/ from the embedded binary, preferring an on-disk copy for the paths in overridablePaths when one exists there - checked once at startup and cached, not per-request.
type StaticAssets struct {
	handler   http.Handler
	CustomCSS bool
	CustomJS  bool
}

// NewStaticAssets builds a StaticAssets serving fsys (already rooted at the static content root, e.g. fs.Sub(assets.Static, "static")) with disk overrides read from overrideDir - overrideDir may be "" to disable overrides entirely.
func NewStaticAssets(fsys fs.FS, overrideDir string, customCode bool) (*StaticAssets, error) {
	overrides := map[string]string{}
	if overrideDir != "" {
		for rel, name := range overridablePaths {
			if customCodePaths[rel] && !customCode {
				continue
			}
			diskPath := filepath.Join(overrideDir, name)
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

// ServeRootFile serves one static/ file at a root-level path (used for /robots.txt and /security.txt) - override-aware the same way Handler() is.
func (s *StaticAssets) ServeRootFile(rel string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := r.Clone(r.Context())
		req.URL.Path = "/" + rel
		s.handler.ServeHTTP(w, req)
	}
}
