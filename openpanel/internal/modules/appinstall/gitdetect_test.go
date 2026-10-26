package appinstall

import (
	"net"
	"testing"
)

func TestGitIPIsPublic(t *testing.T) {
	cases := []struct {
		ip   string
		want bool
	}{
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"127.0.0.1", false},
		{"10.0.0.5", false},
		{"192.168.1.10", false},
		{"172.16.3.4", false},
		{"169.254.169.254", false}, // cloud metadata
		{"100.64.0.1", false},      // CGNAT
		{"::1", false},
		{"fe80::1", false},
	}
	for _, c := range cases {
		if got := gitIPIsPublic(net.ParseIP(c.ip)); got != c.want {
			t.Errorf("gitIPIsPublic(%s) = %v, want %v", c.ip, got, c.want)
		}
	}
}

func TestResolvePublicGitHostRejectsInternal(t *testing.T) {
	// literal IPs need no DNS, so these are deterministic
	for _, u := range []string{
		"https://127.0.0.1/x.git",
		"https://169.254.169.254/x.git",
		"https://10.1.2.3/x.git",
		"https://[::1]/x.git",
	} {
		if err := resolvePublicGitHost(u); err == nil {
			t.Errorf("resolvePublicGitHost(%q) allowed an internal address", u)
		}
	}
}
