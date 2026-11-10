package serverinfo

import "testing"

func TestCPUModel(t *testing.T) {
	x86 := "processor\t: 0\nmodel name\t: Intel Xeon Processor (Cascadelake)\n\nprocessor\t: 1\nmodel name\t: Intel Xeon Processor (Cascadelake)\n"
	if got := cpuModel(x86); got != "Intel Xeon Processor (Cascadelake) · 2 cores" {
		t.Errorf("x86 = %q", got)
	}
	arm := "processor\t: 0\nBogoMIPS\t: 48.00\nHardware\t: BCM2835\n"
	if got := cpuModel(arm); got != "BCM2835" {
		t.Errorf("arm = %q", got)
	}
	if got := cpuModel("flags: sse\n"); got != "" {
		t.Errorf("no model = %q", got)
	}
}
