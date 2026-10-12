package filemanager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestDownloadToFileBlocksLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("secret")) }))
	defer srv.Close()
	dest := filepath.Join(t.TempDir(), "out")
	if err := downloadToFile(context.Background(), srv.URL, dest, 1<<20, nil); !errors.Is(err, errBlockedAddress) {
		t.Fatalf("err = %v, want errBlockedAddress", err)
	}
}
