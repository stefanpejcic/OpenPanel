package wordpress

import "testing"

func TestWPManagerRuleRE(t *testing.T) {
	content := "some caddy directive\nwp_manager_block_xmlrpc\nother line wp_manager_disable_file_edit more\n"
	matches := wpManagerRuleRE.FindAllString(content, -1)
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %v", matches)
	}
}

func TestEscapeMySQLString(t *testing.T) {
	if got := escapeMySQLString(`o'brien\`); got != `o\'brien\\` {
		t.Errorf("escapeMySQLString = %q", got)
	}
}

func TestEscapePHPSingleQuoted(t *testing.T) {
	if got := escapePHPSingleQuoted(`pass'word\`); got != `pass\'word\\` {
		t.Errorf("escapePHPSingleQuoted = %q", got)
	}
}

func TestCloneWPConfigREs(t *testing.T) {
	content := `define('DB_NAME', 'old_db');
define('DB_USER', 'old_user');
define('DB_PASSWORD', 'old_pass');`
	if !cloneDBNameRE.MatchString(content) {
		t.Error("cloneDBNameRE should match")
	}
	if !cloneDBUserRE.MatchString(content) {
		t.Error("cloneDBUserRE should match")
	}
	if !cloneDBPasswordRE.MatchString(content) {
		t.Error("cloneDBPasswordRE should match")
	}
}

func TestRemoveDBNameAndUserRE(t *testing.T) {
	content := `define('DB_NAME', 'my_db');
define('DB_USER', 'my_user');`
	m := removeDBNameRE.FindStringSubmatch(content)
	if m == nil || m[1] != "my_db" {
		t.Fatalf("removeDBNameRE match = %v", m)
	}
	m2 := removeDBUserRE.FindStringSubmatch(content)
	if m2 == nil || m2[1] != "my_user" {
		t.Fatalf("removeDBUserRE match = %v", m2)
	}
}
