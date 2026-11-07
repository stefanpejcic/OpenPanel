package php

import "testing"

func TestIniVersionRejectsTraversal(t *testing.T) {
	for _, v := range []string{"../../bob/php.ini/8.3", "8.3/../../x", "..", "8.3.ini", ""} {
		if phpVersionFormRE.MatchString(v) {
			t.Errorf("accepted %q", v)
		}
	}
	if err := updatePHPConfigFile("alice", "../../bob/php.ini/8.3", nil, nil); err == nil {
		t.Fatal("updatePHPConfigFile took a traversal version")
	}
}
