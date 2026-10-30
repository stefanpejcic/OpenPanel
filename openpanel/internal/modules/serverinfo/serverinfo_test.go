package serverinfo

import "testing"

func TestUnameToString(t *testing.T) {
	var b [65]int8
	for i, c := range "Linux" {
		b[i] = int8(c)
	}
	if got := unameToString(b); got != "Linux" {
		t.Errorf("got %q, want %q", got, "Linux")
	}
	if got := unameToString([65]int8{}); got != "" {
		t.Errorf("empty array: got %q, want empty", got)
	}
}
