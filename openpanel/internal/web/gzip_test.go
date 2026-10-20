package web

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func gzipGet(t *testing.T, h http.HandlerFunc, header map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	Gzip(h).ServeHTTP(rec, req)
	return rec
}

func TestGzipCompressesHTML(t *testing.T) {
	page := strings.Repeat("<p>hello</p>", 500)
	rec := gzipGet(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))
	}, nil)
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("html not compressed: %v", rec.Header())
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(zr)
	if string(got) != page {
		t.Error("body changed")
	}
}

func TestGzipLeavesOthersAlone(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"download": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/sql")
			w.Write([]byte("CREATE TABLE t;"))
		},
		"already gzipped": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Header().Set("Content-Encoding", "gzip")
			w.Write([]byte("x"))
		},
		"event stream": func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Write([]byte("data: 1\n\n"))
			w.(http.Flusher).Flush()
		},
	}
	for name, h := range cases {
		rec := gzipGet(t, h, nil)
		if rec.Header().Get("Content-Encoding") == "gzip" && name != "already gzipped" {
			t.Errorf("%s should not be compressed", name)
		}
		if name == "already gzipped" && rec.Body.String() != "x" {
			t.Errorf("%s was compressed twice", name)
		}
	}
	rec := gzipGet(t, func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("<html></html>")) }, map[string]string{"Upgrade": "websocket"})
	if rec.Header().Get("Content-Encoding") != "" {
		t.Error("websocket upgrades must pass through untouched")
	}
}

func TestGzipSkipsAttachments(t *testing.T) {
	rec := gzipGet(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Disposition", `attachment; filename="error.log"`)
		w.Write([]byte(strings.Repeat("line\n", 500)))
	}, nil)
	if rec.Header().Get("Content-Encoding") != "" {
		t.Error("attachment was compressed")
	}
}
