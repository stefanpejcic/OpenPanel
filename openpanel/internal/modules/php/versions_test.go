package php

import "testing"

func TestCheckDefaultVersion(t *testing.T) {
	installed := []string{"8.5", "8.4"}
	ok := []struct {
		v  string
		ls bool
	}{{"8.4", false}, {"8.5", false}, {"8.5", true}, {"8.2", true}}
	for _, c := range ok {
		if msg := checkDefaultVersion(c.v, c.ls, installed); msg != "" {
			t.Errorf("checkDefaultVersion(%q, litespeed=%v) = %q, want ok", c.v, c.ls, msg)
		}
	}
	bad := []struct {
		v  string
		ls bool
	}{
		{"7.4", false},      // not installed on this account
		{"8.50", false},     // looks like a version but isn't one we have
		{"", false},         // empty
		{"8.3\nX=1", false}, // only a plain major.minor is written to .env
		{"8.50", true},      // parsed as 8.5 before, giving tag lsphp850
		{"08.5", true},
		{"8.5e0", true},
		{"8.1", true}, // not a litespeed tag
	}
	for _, c := range bad {
		if msg := checkDefaultVersion(c.v, c.ls, installed); msg == "" {
			t.Errorf("checkDefaultVersion(%q, litespeed=%v) should be refused", c.v, c.ls)
		}
	}
}

func TestSetEnvLine(t *testing.T) {
	env := "WEB_SERVER=apache\nDEFAULT_PHP_VERSION=8.4\nMYSQL_TYPE=mariadb\n"
	got, found := setEnvLine(env, "DEFAULT_PHP_VERSION", "8.5")
	if !found || got != "WEB_SERVER=apache\nDEFAULT_PHP_VERSION=8.5\nMYSQL_TYPE=mariadb\n" {
		t.Errorf("setEnvLine = %q, %v", got, found)
	}
	// no line to update has to be reported, not saved as a success
	if got, found := setEnvLine("WEB_SERVER=apache\n", "DEFAULT_PHP_VERSION", "8.5"); found || got != "WEB_SERVER=apache\n" {
		t.Errorf("missing key: setEnvLine = %q, %v", got, found)
	}
}

func TestVhostVersionSwitch(t *testing.T) {
	vhost := "<VirtualHost *:80>\n  ProxyPassMatch ^/(.*\\.php)$ fcgi://php-fpm-8.4:9000/var/www/html/$1\n</VirtualHost>\n"
	if v := currentVhostVersion(vhost); v != "8.4" {
		t.Fatalf("currentVhostVersion = %q, want 8.4", v)
	}
	updated, ok := switchVhostVersion(vhost, "8.4", "8.5")
	if !ok || currentVhostVersion(updated) != "8.5" {
		t.Errorf("switch 8.4 -> 8.5: ok=%v version=%q", ok, currentVhostVersion(updated))
	}
	// a stale old version from the page used to change nothing yet report success
	if _, ok := switchVhostVersion(vhost, "8.2", "8.5"); ok {
		t.Error("switching from a version the vhost doesn't use should not be ok")
	}
	if v := currentVhostVersion("<VirtualHost *:80>\n</VirtualHost>\n"); v != "" {
		t.Errorf("vhost without php-fpm = %q, want empty", v)
	}
}
