package websites

import "testing"

func TestParseDBVersion(t *testing.T) {
	for out, want := range map[string]string{
		"mariadb from 13.0.2-MariaDB, client 15.2 for debian-linux-gnu (x86_64) using  EditLine wrapper": "13.0.2",
		"mysql  Ver 9.4.0 for Linux on x86_64 (MySQL Community Server - GPL)":                            "9.4.0",
		"Error: no container with name or ID \"mariadb\" found":                                          "",
		"": "",
	} {
		if got := parseDBVersion(out); got != want {
			t.Errorf("parseDBVersion(%q) = %q, want %q", out, got, want)
		}
	}
}
