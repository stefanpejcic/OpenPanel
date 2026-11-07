package postgresql

import (
	"testing"
)

func TestShortPGVersion(t *testing.T) {
	cases := map[string]string{"17.4": "17.4", "16.2 (Debian 16.2-1.pgdg120+2)": "16.2", "": ""}
	for in, want := range cases {
		if got := shortPGVersion(in); got != want {
			t.Errorf("shortPGVersion(%q) = %q, want %q", in, got, want)
		}
	}
}
