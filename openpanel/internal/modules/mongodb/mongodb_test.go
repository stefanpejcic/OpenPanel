package mongodb

import (
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/mongomanager"
)

func TestOpKillable(t *testing.T) {
	cases := []struct {
		name string
		op   mongomanager.Op
		want bool
	}{
		{"running client query", mongomanager.Op{IsClient: true, Active: true, Op: "query"}, true},
		{"background thread", mongomanager.Op{IsClient: false, Active: true, Op: "none"}, false},
		{"background thread with op", mongomanager.Op{IsClient: false, Active: true, Op: "command"}, false},
		{"inactive client", mongomanager.Op{IsClient: true, Active: false, Op: "query"}, false},
		{"client with no op", mongomanager.Op{IsClient: true, Active: true, Op: "none"}, false},
	}
	for _, c := range cases {
		if got := opKillable(c.op); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestKillOpRejectsBadIDs(t *testing.T) {
	for _, id := range []string{"", "0", "-3", "abc", "1; db.dropDatabase()", "1e5"} {
		if err := killOp(t.Context(), "nobody", id); err != errInvalidOpID {
			t.Errorf("id %q: got %v, want errInvalidOpID", id, err)
		}
	}
}
