package reqip

import (
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	cases := []struct {
		name, remote, xff, cf, want string
	}{
		{"direct client, header ignored", "203.0.113.7:5000", "", "198.51.100.9", "203.0.113.7"},
		{"direct from cloudflare", "172.70.1.2:443", "", "198.51.100.9", "198.51.100.9"},
		{"cloudflare over ipv6", "[2606:4700::1]:443", "", "2001:db8::5", "2001:db8::5"},
		{"behind caddy, not via cloudflare", "127.0.0.1:40000", "203.0.113.7", "198.51.100.9", "203.0.113.7"},
		{"behind caddy via cloudflare", "127.0.0.1:40000", "198.51.100.9, 162.158.1.1", "198.51.100.9", "198.51.100.9"},
		{"behind podman gateway", "10.88.0.1:40000", "203.0.113.7", "", "203.0.113.7"},
		{"forwarded header from a public peer ignored", "203.0.113.7:5000", "162.158.1.1", "198.51.100.9", "203.0.113.7"},
		{"bad cloudflare header falls back", "162.158.1.1:443", "", "not-an-ip", "162.158.1.1"},
		{"local proxy without forwarded header", "127.0.0.1:40000", "", "", "127.0.0.1"},
		{"bad forwarded header keeps the peer", "127.0.0.1:40000", "junk", "198.51.100.9", "127.0.0.1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/login", nil)
			r.RemoteAddr = c.remote
			if c.xff != "" {
				r.Header.Set("X-Forwarded-For", c.xff)
			}
			if c.cf != "" {
				r.Header.Set("CF-Connecting-IP", c.cf)
			}
			if got := ClientIP(r); got != c.want {
				t.Errorf("ClientIP = %q, want %q", got, c.want)
			}
		})
	}
}
