package filemanager

import "testing"

func TestHasBareName(t *testing.T) {
	bare := []string{"error_log", "env", "gitconfig", "cfg", "htaccess", "gemfile"}
	for name, want := range map[string]bool{
		".env": true, ".env.local": true, "error_log": true, "php_error_log": true, ".gitconfig": true,
		"app.cfg": true, ".htaccess": true, "gemfile": true, "gemfile.lock": true, "my-env": true,
		// substring matches used to skip the binary check, so a pdf opened in the editor
		"envelope.pdf": false, "environment": false, "config.cfgx": false, "nogemfiles": false,
	} {
		if got := hasBareName(name, bare); got != want {
			t.Errorf("hasBareName(%q) = %v, want %v", name, got, want)
		}
	}
}
