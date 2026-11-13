package phpapp

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestWrapperStrip(t *testing.T) {
	cases := []struct {
		name  string
		names []string
		want  int
	}{
		{"github wrapper", []string{"log-3.0.0/", "log-3.0.0/composer.json", "log-3.0.0/src/Log.php"}, 1},
		{"dot prefix", []string{"./", "./app/", "./app/index.php"}, 1},
		{"no dir entry", []string{"app/index.php"}, 1},
		{"flat files", []string{"index.php", "composer.json"}, 0},
		{"two top folders", []string{"app/a.php", "lib/b.php"}, 0},
		{"folder next to a file", []string{"app/a.php", "index.php"}, 0},
		{"single file", []string{"index.php"}, 0},
		{"empty folder only", []string{"app/"}, 0},
		{"empty", nil, 0},
		{"climbs out", []string{"app/a.php", "../b.php"}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := wrapperStrip(c.names); got != c.want {
				t.Errorf("wrapperStrip(%v) = %d, want %d", c.names, got, c.want)
			}
		})
	}
}

func makeZip(t *testing.T, names ...string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, n := range names {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// same shape as a github tarball, global header first
func makeTarGz(t *testing.T, names ...string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{Typeflag: tar.TypeXGlobalHeader, Name: "pax_global_header", PAXRecords: map[string]string{"comment": "abc"}}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	for _, n := range names {
		if err := tw.WriteHeader(&tar.Header{Name: n, Typeflag: tar.TypeReg, Mode: 0o644, Size: 1}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDownloadAndExtractInitialProject(t *testing.T) {
	cases := []struct {
		name string
		file string
		body []byte
		want []string
		gone []string
	}{
		{"zip wrapper is dropped", "p.zip", makeZip(t, "proj-1.0/composer.json", "proj-1.0/src/a.php"), []string{"composer.json", "src/a.php"}, []string{"proj-1.0"}},
		{"flat zip stays as is", "p.zip", makeZip(t, "composer.json", "src/a.php"), []string{"composer.json", "src/a.php"}, nil},
		{"two folders stay as is", "p.zip", makeZip(t, "app/a.php", "lib/b.php"), []string{"app/a.php", "lib/b.php"}, nil},
		{"tarball wrapper is dropped", "p.tar.gz", makeTarGz(t, "proj-1.0/composer.json", "proj-1.0/src/a.php"), []string{"composer.json", "src/a.php"}, []string{"proj-1.0", "pax_global_header"}},
	}
	// the real downloadTransport refuses loopback (SSRF guard); the test server is on 127.0.0.1
	oldTransport := downloadTransport
	downloadTransport = http.DefaultTransport
	defer func() { downloadTransport = oldTransport }()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(c.body) }))
			defer srv.Close()
			dest := t.TempDir()
			if err := downloadAndExtractInitialProject(context.Background(), srv.URL+"/"+c.file, dest); err != nil {
				t.Fatalf("extract: %v", err)
			}
			for _, f := range c.want {
				if _, err := os.Stat(filepath.Join(dest, f)); err != nil {
					t.Errorf("missing %s: %v", f, err)
				}
			}
			for _, f := range c.gone {
				if _, err := os.Stat(filepath.Join(dest, f)); err == nil {
					t.Errorf("%s should not exist", f)
				}
			}
		})
	}
}

func TestDownloadAndExtractRejectsZipSlip(t *testing.T) {
	body := makeZip(t, "app/a.php", "../evil.php")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(body) }))
	defer srv.Close()
	parent := t.TempDir()
	dest := filepath.Join(parent, "site")
	if err := os.Mkdir(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := downloadAndExtractInitialProject(context.Background(), srv.URL+"/p.zip", dest); err == nil {
		t.Fatal("expected an error")
	}
	if _, err := os.Stat(filepath.Join(parent, "evil.php")); err == nil {
		t.Error("evil.php got written outside dest")
	}
}

func TestFirstMissingDir(t *testing.T) {
	base := t.TempDir()
	if got := firstMissingDir(base); got != base {
		t.Errorf("existing dir: got %s", got)
	}
	want := filepath.Join(base, "a")
	if got := firstMissingDir(filepath.Join(base, "a", "b", "c")); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
