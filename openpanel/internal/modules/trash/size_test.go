package trash

import "testing"

func TestParseDuKB(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"1258291\t.\n", 1258291 * 1024, true},
		{"0\t.\n", 0, true},
		{"", 0, false},
		{"du: cannot read directory\n", 0, false},
	}
	for _, c := range cases {
		got, ok := parseDuKB(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("parseDuKB(%q) = %d, %v, want %d, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}
