package web

import (
	"bufio"
	"compress/gzip"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
)

// level 5 gets most of the size win of 9 at a fraction of the CPU
var gzipWriters = sync.Pool{New: func() any {
	w, _ := gzip.NewWriterLevel(nil, 5)
	return w
}}

// gzipTypes are the responses worth compressing, downloads and event streams are left alone
var gzipTypes = []string{"text/html", "application/json", "text/plain", "text/css", "application/javascript", "text/javascript", "image/svg+xml"}

// Gzip compresses HTML, JSON, CSS and JS responses for browsers that accept gzip
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead || r.Header.Get("Upgrade") != "" || !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gw := &gzipResponseWriter{ResponseWriter: w}
		defer gw.close()
		next.ServeHTTP(gw, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	decided bool
	zw      *gzip.Writer
}

// decide looks at the headers once, right before the first byte goes out
func (g *gzipResponseWriter) decide(status int, first []byte) {
	if g.decided {
		return
	}
	g.decided = true
	h := g.Header()
	h.Add("Vary", "Accept-Encoding")
	// downloads keep their Content-Length so the browser can show progress
	if h.Get("Content-Encoding") != "" || h.Get("Content-Disposition") != "" || status == http.StatusNoContent || status == http.StatusNotModified || status == http.StatusPartialContent || (status >= 100 && status < 200) {
		return
	}
	ct := h.Get("Content-Type")
	if ct == "" && first != nil {
		ct = http.DetectContentType(first)
		h.Set("Content-Type", ct)
	}
	for _, t := range gzipTypes {
		if strings.HasPrefix(ct, t) {
			h.Set("Content-Encoding", "gzip")
			h.Del("Content-Length")
			g.zw = gzipWriters.Get().(*gzip.Writer)
			g.zw.Reset(g.ResponseWriter)
			return
		}
	}
}

func (g *gzipResponseWriter) WriteHeader(status int) {
	g.decide(status, nil)
	g.ResponseWriter.WriteHeader(status)
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	g.decide(http.StatusOK, b)
	if g.zw != nil {
		return g.zw.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipResponseWriter) Flush() {
	if g.zw != nil {
		_ = g.zw.Flush()
	}
	if f, ok := g.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack is only reached by a handler that upgrades without an Upgrade header, which can't happen with gzip already chosen
func (g *gzipResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if g.zw != nil {
		return nil, nil, errors.New("gzip: response already compressed")
	}
	return http.NewResponseController(g.ResponseWriter).Hijack()
}

func (g *gzipResponseWriter) Unwrap() http.ResponseWriter { return g.ResponseWriter }

func (g *gzipResponseWriter) close() {
	if g.zw == nil {
		return
	}
	_ = g.zw.Close()
	g.zw.Reset(nil)
	gzipWriters.Put(g.zw)
	g.zw = nil
}
