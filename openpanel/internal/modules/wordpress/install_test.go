package wordpress

import "testing"

func TestValidateInstallInputs(t *testing.T) {
	if msg := validateInstallInputs("latest", "u_wpdb", "u_wpuser", "wp_"); msg != "" {
		t.Fatalf("defaults should pass, got %q", msg)
	}
	for _, v := range []string{"6.8.1", "6.8", "6.9-RC1"} {
		if msg := validateInstallInputs(v, "db", "user", "wp_"); msg != "" {
			t.Errorf("version %q should pass, got %q", v, msg)
		}
	}
	bad := []struct{ version, db, user, prefix string }{
		// the version lands in /etc/openpanel/wordpress/archives/wordpress-<v>.tar.gz
		{"/../../../../../home/victim/docker-data/volumes/victim_html_data/_data/backup", "db", "user", "wp_"},
		{"6.8/../../x", "db", "user", "wp_"},
		{"", "db", "user", "wp_"},
		{"latest", "db`; DROP DATABASE x; --", "user", "wp_"},
		{"latest", "db", "user'@'%", "wp_"},
		{"latest", "db", "user", "$1_"},
		{"latest", "db", "user", "wp';phpinfo();//_"},
	}
	for _, c := range bad {
		if msg := validateInstallInputs(c.version, c.db, c.user, c.prefix); msg == "" {
			t.Errorf("validateInstallInputs(%q, %q, %q, %q) should fail", c.version, c.db, c.user, c.prefix)
		}
	}
}
