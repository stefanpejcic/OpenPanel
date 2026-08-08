package ftp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withFTPServerConf points ftpServerConfPath at a temp file for the
// duration of the test.
func withFTPServerConf(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "server.conf")
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	orig := ftpServerConfPath
	ftpServerConfPath = path
	t.Cleanup(func() { ftpServerConfPath = orig })
}

func TestResolveFTPHostPortNoOverride(t *testing.T) {
	withFTPServerConf(t, "")
	host, port := resolveFTPHostPort("1.2.3.4")
	if host != "1.2.3.4" {
		t.Errorf("host = %q, want default 1.2.3.4", host)
	}
	if port != "21" {
		t.Errorf("port = %q, want default 21", port)
	}
}

func TestResolveFTPHostPortWithOverride(t *testing.T) {
	withFTPServerConf(t, "hostname=ftp.example.com\nport=2121\n")
	host, port := resolveFTPHostPort("1.2.3.4")
	if host != "ftp.example.com" {
		t.Errorf("host = %q, want ftp.example.com", host)
	}
	if port != "2121" {
		t.Errorf("port = %q, want 2121", port)
	}
}

func TestResolveFTPHostPortInvalidPortIgnored(t *testing.T) {
	withFTPServerConf(t, "port=notanumber\n")
	_, port := resolveFTPHostPort("1.2.3.4")
	if port != "21" {
		t.Errorf("port = %q, want fallback 21 for invalid override", port)
	}

	withFTPServerConf(t, "port=99999\n")
	_, port = resolveFTPHostPort("1.2.3.4")
	if port != "21" {
		t.Errorf("port = %q, want fallback 21 for out-of-range override", port)
	}
}

func TestRenderFTPClientTemplateMissingFile(t *testing.T) {
	dir := t.TempDir()
	_, ok := renderFTPClientTemplate(filepath.Join(dir, "nope.conf"), "1.2.3.4", "21", "bob@example.com", "/var/www/html/")
	if ok {
		t.Error("expected ok=false for a missing template file")
	}
}

func TestRenderFTPClientTemplateEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.conf")
	if err := os.WriteFile(path, []byte("   \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, ok := renderFTPClientTemplate(path, "1.2.3.4", "21", "bob@example.com", "/var/www/html/")
	if ok {
		t.Error("expected ok=false for an empty template file")
	}
}

func TestRenderFTPClientTemplateInvalidXML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.conf")
	if err := os.WriteFile(path, []byte("<bookmark><hostname>{host}</bookmark>"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, ok := renderFTPClientTemplate(path, "1.2.3.4", "21", "bob@example.com", "/var/www/html/")
	if ok {
		t.Error("expected ok=false for malformed XML (mismatched tags)")
	}
}

func TestRenderFTPClientTemplateValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "good.conf")
	tmpl := `<?xml version="1.0" encoding="UTF-8"?>
<bookmark>
    <hostname>{host}</hostname>
    <port>{port}</port>
    <username>{username}</username>
    <path>{path}</path>
</bookmark>`
	if err := os.WriteFile(path, []byte(tmpl), 0o644); err != nil {
		t.Fatal(err)
	}
	content, ok := renderFTPClientTemplate(path, "ftp.example.com", "2121", "bob@example.com", "/var/www/html/")
	if !ok {
		t.Fatal("expected ok=true for a valid, well-formed template")
	}
	for _, want := range []string{"<hostname>ftp.example.com</hostname>", "<port>2121</port>", "<username>bob@example.com</username>", "<path>/var/www/html/</path>"} {
		if !strings.Contains(content, want) {
			t.Errorf("rendered content missing %q, got: %s", want, content)
		}
	}
}
