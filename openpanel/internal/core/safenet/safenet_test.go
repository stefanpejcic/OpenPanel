package safenet

import (
	"net"
	"testing"
)

func TestIsPublicIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1", "10.0.0.1", "169.254.169.254", "100.64.0.1", "0.0.0.0", "::1", "::",
		"::ffff:169.254.169.254", "fe80::1", "fc00::1",
		"64:ff9b::a9fe:a9fe", "64:ff9b::7f00:1", "64:ff9b:1::1",
		"2002:a9fe:a9fe::", "::7f00:1", "::a9fe:a9fe", "fec0::1", "2001:0:4136:e378::1",
	}
	for _, s := range blocked {
		if IsPublicIP(net.ParseIP(s)) {
			t.Errorf("%s should be blocked", s)
		}
	}
	for _, s := range []string{"8.8.8.8", "1.1.1.1", "2001:4860:4860::8888", "2606:4700:4700::1111"} {
		if !IsPublicIP(net.ParseIP(s)) {
			t.Errorf("%s should be allowed", s)
		}
	}
}
