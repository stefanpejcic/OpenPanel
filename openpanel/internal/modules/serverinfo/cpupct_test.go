package serverinfo

import "testing"

func TestCPUPct(t *testing.T) {
	for _, c := range []struct {
		usage, total float64
		want         int
	}{
		{101, 200, 51}, // 1.01 of 2 cores
		{79, 200, 40},
		{50, 100, 50},
		{30, 0, 30}, // no limit known, keep the raw figure
	} {
		var l ResourceUsageLine
		l.CPU.Usage.Pct, l.CPU.Total.Pct = c.usage, c.total
		if got := l.CPUPct(); got != c.want {
			t.Errorf("CPUPct(%v of %v) = %d, want %d", c.usage, c.total, got, c.want)
		}
	}
}
