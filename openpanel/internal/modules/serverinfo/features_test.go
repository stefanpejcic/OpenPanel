package serverinfo

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFeatures(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	// trimmed from the VM's /etc/openpanel/openadmin/config/features.json
	good := write("features.json", `[
  {
    "name": "notifications",
    "title": "Email Notifications",
    "description": "If enabled, users can configure notification preferences from their account and they will receive emails.",
    "link": "/account/notifications",
    "type": "community",
    "help_link": "https://openpanel.com/docs/admin/settings/modules/#notifications"
  }
]`)
	if raw, ok := readFeatures(good); !ok || len(raw) == 0 {
		t.Errorf("readFeatures(good) = %d bytes, %v", len(raw), ok)
	}
	for name, path := range map[string]string{
		"missing":    filepath.Join(dir, "nope.json"),
		"object":     write("obj.json", `{"name": "x"}`),
		"truncated":  write("bad.json", `[{"name": "x"`),
		"empty file": write("empty.json", ``),
	} {
		if _, ok := readFeatures(path); ok {
			t.Errorf("readFeatures(%s) should fail", name)
		}
	}
}

func TestHandleSystemFeaturesWithoutFile(t *testing.T) {
	if _, err := os.Stat(featuresFile); err == nil {
		t.Skip("features file exists on this machine")
	}
	w := httptest.NewRecorder()
	handleSystemFeatures(w, httptest.NewRequest("GET", "/json/system/features", nil))
	if w.Code != 404 {
		t.Errorf("status = %d, want 404 so the page falls back to GitHub", w.Code)
	}
}
