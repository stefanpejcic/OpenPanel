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

func TestAppendZoneRecordRefusesDirectives(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.zone")
	for _, line := range []string{"$GENERATE 1-4294967294 h$ IN A 192.0.2.1", "$INCLUDE /etc/bind/rndc.key ; IN A x", "a 3600 IN A 1.2.3.4\n$INCLUDE /x"} {
		if err := AppendZoneRecord(p, line); err == nil {
			t.Errorf("appended %q", line)
		}
	}
	if err := AppendZoneRecord(p, "www 3600 IN A 192.0.2.1"); err != nil {
		t.Fatalf("normal record refused: %v", err)
	}
}

func TestNormalizeZoneLineEndings(t *testing.T) {
	got := NormalizeZoneLineEndings("$TTL 1h\r\nwww IN A 1.1.1.1\r\n")
	if got != "$TTL 1h\nwww IN A 1.1.1.1\n" {
		t.Errorf("got %q", got)
	}
	if msg := ZoneDirectiveError(got); msg != "" {
		t.Errorf("a CRLF zone should save once normalized: %s", msg)
	}
	if msg := ZoneDirectiveError(NormalizeZoneLineEndings("www IN A 1.1.1.1\r$INCLUDE /etc/passwd\n")); msg == "" {
		t.Error("a lone carriage return has to stay refused")
	}
}
