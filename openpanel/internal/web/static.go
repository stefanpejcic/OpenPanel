package web

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
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
	assetFS = fsys

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/")
		if diskPath, ok := overrides[rel]; ok {
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFile(w, r, diskPath)
			return
		}
		// embedded files have no mod time, so without an ETag the browser downloads them again on every page
		if h := assetHash(rel); h != "" {
			w.Header().Set("ETag", `"`+h+`"`)
		}
		w.Header().Set("Cache-Control", staticCacheControl(rel, r.URL.Query().Get("v") != ""))
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

var (
	assetFS     fs.FS
	assetHashes sync.Map
)

// StaticURL is the {{static}} template func, embedded files get ?v=<content hash> so browsers can keep them for good
func StaticURL(rel string) string {
	if h := assetHash(rel); h != "" {
		return "/static/" + rel + "?v=" + h
	}
	return "/static/" + rel
}

// assetHash is a short content hash of an embedded file, "" for overridable files, directories and missing files
func assetHash(rel string) string {
	if assetFS == nil {
		return ""
	}
	if _, ok := overridablePaths[rel]; ok {
		return ""
	}
	if v, ok := assetHashes.Load(rel); ok {
		return v.(string)
	}
	h := ""
	if b, err := fs.ReadFile(assetFS, rel); err == nil {
		sum := sha256.Sum256(b)
		h = hex.EncodeToString(sum[:6])
	}
	assetHashes.Store(rel, h)
	return h
}

// staticCacheControl: versioned URLs never change, vendor files only change with an upgrade, the rest gets checked each time
func staticCacheControl(rel string, versioned bool) string {
	switch {
	case versioned:
		return "public, max-age=31536000, immutable"
	case strings.HasPrefix(rel, "vendor/"), strings.HasPrefix(rel, "flags/"):
		return "public, max-age=86400"
	default:
		return "no-cache"
	}
}
