package processmanager

import "testing"

func TestServiceNamesFromCompose(t *testing.T) {
	compose := map[string]any{
		"services": map[string]any{"mysql": map[string]any{}, "nginx-proxy": map[string]any{}},
	}
	names := serviceNamesFromCompose(compose)
	set := map[string]bool{}
	for _, n := range names {
		set[n] = true
	}
	if len(names) != 2 || !set["mysql"] || !set["nginx-proxy"] {
		t.Errorf("names = %v", names)
	}
}

func TestServiceNamesFromComposeMissing(t *testing.T) {
	if names := serviceNamesFromCompose(map[string]any{}); names != nil {
		t.Errorf("expected nil, got %v", names)
	}
}

func TestIsDisplayableCmd(t *testing.T) {
	cases := []struct {
		cmd  string
		want bool
	}{
		{"mysqld", true},
		{"/etc/entrypoint.sh", false},
		{"sh -c ps -eo pid,%cpu,time,cmd", false},
		{"tail -f /dev/null", false},
	}
	for _, c := range cases {
		if got := isDisplayableCmd(c.cmd); got != c.want {
			t.Errorf("isDisplayableCmd(%q) = %v, want %v", c.cmd, got, c.want)
		}
	}
}

func TestTimeFieldRE(t *testing.T) {
	valid := []string{"7s", "23m13s", "1h2m3s", "12:41", "01:02:03"}
	for _, v := range valid {
		if !timeFieldRE.MatchString(v) {
			t.Errorf("expected %q to match", v)
		}
	}
	invalid := []string{"", "abc", "2026-07-31"}
	for _, v := range invalid {
		if timeFieldRE.MatchString(v) {
			t.Errorf("expected %q to not match", v)
		}
	}
}

func TestProcessKey(t *testing.T) {
	if processKey("mysql", "123") == processKey("nginx", "123") {
		t.Error("expected different containers with the same PID to produce different keys")
	}
	if processKey("mysql", "123") != processKey("mysql", "123") {
		t.Error("expected identical inputs to produce identical keys")
	}
}
