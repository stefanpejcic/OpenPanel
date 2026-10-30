package php

import (
	"testing"
)

func TestIniLineCount(t *testing.T) {
	cases := map[string]int{"": 0, "a": 1, "a\n": 1, "a\nb\n": 2, "a\n\nb": 3}
	for in, want := range cases {
		if got := (PHPIniEditorPageData{FileContent: in}).LineCount(); got != want {
			t.Errorf("LineCount(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestWebServerLabel(t *testing.T) {
	cases := map[string]string{"apache": "Apache", "openlitespeed": "OpenLiteSpeed", "nginx": "Nginx", "": "", "caddy": "caddy"}
	for in, want := range cases {
		if got := webServerLabel(in); got != want {
			t.Errorf("webServerLabel(%q) = %q, want %q", in, got, want)
		}
	}
}
