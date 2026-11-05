package dns

import (
	"path/filepath"
	"testing"
)

func TestZoneDirectiveError(t *testing.T) {
	ok := "$TTL 1h\n$ORIGIN example.com.\n@ IN A 192.0.2.1\n"
	if msg := ZoneDirectiveError(ok); msg != "" {
		t.Fatalf("rejected a normal zone: %s", msg)
	}
	for _, bad := range []string{
		"$GENERATE 0-4294967294 h$ A 192.0.2.1 ;\n",
		"@ IN A 192.0.2.1\n$INCLUDE /etc/bind/rndc.key\n",
		"$include /etc/passwd\n",
	} {
		if ZoneDirectiveError(bad) == "" {
			t.Errorf("accepted %q", bad)
		}
	}
	if err := WriteZoneFile(filepath.Join(t.TempDir(), "x.zone"), []byte("$GENERATE 1-9 h$ A 192.0.2.1\n")); err == nil {
		t.Fatal("WriteZoneFile wrote a $GENERATE zone")
	}
}
