package mongodb

import (
	"testing"
)

func TestShouldCountUsers(t *testing.T) {
	for _, c := range []struct {
		state string
		dbs   int
		want  bool
	}{
		{"running", 0, true},
		{"running", 2, true},
		{"exited", 0, false},
		{"", 0, false},
		{"exited", 1, true},
	} {
		if got := shouldCountUsers(c.state, c.dbs); got != c.want {
			t.Errorf("shouldCountUsers(%q, %d) = %v, want %v", c.state, c.dbs, got, c.want)
		}
	}
}
