package ipblocker

import "testing"

func TestNormalizeIP(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOK bool
	}{
		{"192.168.1.1", "192.168.1.1", true},
		{"10.0.0.5/24", "10.0.0.0/24", true},
		{"8.8.8.8/32", "8.8.8.8/32", true},
		{"::1", "::1", true},
		{"2001:db8::/32", "2001:db8::/32", true},
		{"not-an-ip", "", false},
		{"999.999.999.999", "", false},
		{"10.0.0.0/99", "", false},
	}
	for _, c := range cases {
		got, ok := normalizeIP(c.in)
		if ok != c.wantOK {
			t.Errorf("normalizeIP(%q) ok = %v, want %v", c.in, ok, c.wantOK)
			continue
		}
		if ok && got != c.want {
			t.Errorf("normalizeIP(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
