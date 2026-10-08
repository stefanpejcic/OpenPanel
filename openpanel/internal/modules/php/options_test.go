package php

import (
	"strings"
	"testing"
)

func TestBlockedStartupFunctions(t *testing.T) {
	cases := map[string]string{
		"":                                  "",
		"system,passthru,proc_open":         "",
		"exec,passthru":                     "exec",
		`"SHELL_EXEC, ini_get_all"`:         "shell_exec,ini_get_all",
		"exec shell_exec,ini_get_all,popen": "exec,shell_exec,ini_get_all",
		"execute,shell_exec_x,ini_get_all_values": "",
	}
	for in, want := range cases {
		if got := strings.Join(blockedStartupFunctions(in), ","); got != want {
			t.Errorf("blockedStartupFunctions(%q) = %q, want %q", in, got, want)
		}
	}
}
