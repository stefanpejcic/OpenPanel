package backups

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func testPlan(t *testing.T) (*restorePlan, string) {
	t.Helper()
	tmp := t.TempDir()
	html, vhosts, mail, zones := filepath.Join(tmp, "html"), filepath.Join(tmp, "vhosts"), filepath.Join(tmp, "mailstore"), filepath.Join(tmp, "zones")
	for _, d := range []string{html, vhosts, filepath.Join(mail, "example.com"), zones} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return &restorePlan{
		sections:  map[string]string{"backup/html": html, "backup/vhosts": vhosts, "backup/crons.ini": filepath.Join(tmp, "crons.ini")},
		domains:   map[string]bool{"example.com": true},
		mailStore: func(d string) string { return filepath.Join(mail, d) },
		zonesDir:  zones,
	}, tmp
}

func TestRestorePlanDest(t *testing.T) {
	plan, tmp := testPlan(t)
	cases := []struct {
		name string
		dest string
		ok   bool
	}{
		{"backup/html/index.php", filepath.Join(tmp, "html/index.php"), true},
		{"backup/crons.ini", filepath.Join(tmp, "crons.ini"), true},
		{"backup/emails/example.com/info/new/1", filepath.Join(tmp, "mailstore/example.com/info/new/1"), true},
		{"backup/emails/other.com/info/new/1", "", false},
		{"backup/dns/example.com.zone", filepath.Join(tmp, "zones/example.com.zone"), true},
		{"backup/dns/other.com.zone", "", false},
		{"backup/mysql/db.sql", "", false},
	}
	for _, c := range cases {
		dest, _, _, ok := plan.dest(c.name)
		if ok != c.ok || dest != c.dest {
			t.Errorf("dest(%q) = %q, %v, want %q, %v", c.name, dest, ok, c.dest, c.ok)
		}
	}
}

func TestWriteTarEntryRefusesSymlinkEscape(t *testing.T) {
	plan, tmp := testPlan(t)
	outside := filepath.Join(tmp, "outside")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		t.Fatal(err)
	}
	// a link the account could plant in its own volume
	if err := os.Symlink(outside, filepath.Join(tmp, "html", "evil")); err != nil {
		t.Fatal(err)
	}
	if _, err := plan.extractEntry(strings.NewReader("x"), "backup/html/evil/passwd", 0); err == nil {
		t.Error("write through a symlink out of the volume should be refused")
	}
	if _, err := os.Stat(filepath.Join(outside, "passwd")); err == nil {
		t.Error("file was written outside the volume")
	}
	if dest, err := plan.extractEntry(strings.NewReader("ok"), "backup/html/sub/ok.txt", 0); err != nil || dest == "" {
		t.Error("normal write should work")
	}
}

func TestSelectiveTargets(t *testing.T) {
	plan, _ := testPlan(t)
	got, err := selectiveTargets("mailboxes", []string{"example.com/info"}, plan)
	if err != nil || !reflect.DeepEqual(got, []string{"backup/emails/example.com/info"}) {
		t.Errorf("mailboxes = %v, %v", got, err)
	}
	for target, values := range map[string][]string{
		"mailboxes": {"other.com/info"},
		"dns_zones": {"other.com"},
		"paths":     {"../etc"},
		"domains":   {"../x"},
	} {
		if _, err := selectiveTargets(target, values, plan); err == nil {
			t.Errorf("%s %v should be refused", target, values)
		}
	}
	// a deleted domain can still be picked, it gets added back before its vhost is restored
	if got, err := selectiveTargets("domains", []string{"gone.com"}, plan); err != nil || !reflect.DeepEqual(got, []string{"backup/dns/gone.com.zone", "backup/vhosts/gone.com.conf"}) {
		t.Errorf("domains = %v, %v", got, err)
	}
}

func TestReadItemsAndDocroot(t *testing.T) {
	tmp := t.TempDir()
	vhosts := filepath.Join(tmp, "v.tsv")
	if err := os.WriteFile(vhosts, []byte("100\t1700000000\texample.com.conf\n50\t1700000000\tgone.com.conf\n10\t1\tREADME\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err := readItems(vhosts, "vhosts", map[string]bool{"example.com": true})
	if err != nil || len(items) != 2 || items[0].Name != "example.com" || !items[0].OnServer || items[1].OnServer {
		t.Errorf("vhost items = %+v, %v", items, err)
	}

	mail := filepath.Join(tmp, "m.tsv")
	if err := os.WriteFile(mail, []byte("2048\t3\texample.com/info\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	items, err = readItems(mail, "mail", map[string]bool{"example.com": true})
	if err != nil || len(items) != 1 || items[0].Count != 3 || items[0].Size != 2048 {
		t.Errorf("mail items = %+v, %v", items, err)
	}

	if got := vhostDocroot([]byte("<VirtualHost *:80>\n    DocumentRoot /var/www/html/example.com\n"), "example.com"); got != "/var/www/html/example.com" {
		t.Errorf("apache docroot = %s", got)
	}
	if got := vhostDocroot([]byte("server {\n    root /var/www/html/site;\n"), "x.com"); got != "/var/www/html/site" {
		t.Errorf("nginx docroot = %s", got)
	}
	if got := vhostDocroot([]byte("root /etc;"), "x.com"); got != "/var/www/html/x.com" {
		t.Errorf("docroot outside /var/www/html should fall back, got %s", got)
	}
}

func TestListKindAndPrune(t *testing.T) {
	if listKind("backup/00-meta/mail.tsv") != "mail" || listKind("backup/00-meta/mysql.json") != "" || listKind("backup/00-meta/evil.tsv") != "" {
		t.Error("unexpected listKind")
	}
	home := t.TempDir()
	for _, n := range []string{"keep.tar.gz.files.tsv", "keep.tar.gz.mail.tsv", "gone.tar.gz.files.tsv"} {
		if err := saveList(home, strings.Split(n, ".tar.gz")[0]+".tar.gz", strings.TrimSuffix(strings.Split(n, ".tar.gz.")[1], ".tsv"), bytes.NewReader(nil)); err != nil {
			t.Fatal(err)
		}
	}
	pruneFileLists(home, []string{"keep.tar.gz"})
	if !listsExist(home, "keep.tar.gz", []string{"files", "mail"}) || listsExist(home, "gone.tar.gz", []string{"files"}) {
		t.Error("prune kept or dropped the wrong lists")
	}
}

func TestSuspendedDomainVhostIsBlocked(t *testing.T) {
	plan, _ := testPlan(t)
	plan.suspended = func(d string) bool { return d == "example.com" }
	if _, _, _, ok := plan.dest("backup/vhosts/example.com.conf"); ok {
		t.Error("vhost of a suspended domain should not be restored")
	}
	if _, err := selectiveTargets("domains", []string{"example.com"}, plan); err == nil {
		t.Error("picking a suspended domain should be refused")
	}
	if _, _, _, ok := plan.dest("backup/vhosts/other.com.conf"); !ok {
		t.Error("vhosts of other domains still restore")
	}
}

func TestVhostPHPVersion(t *testing.T) {
	if v := vhostPHPVersion([]byte(`SetHandler "proxy:fcgi://php-fpm-8.5:9000"`)); v != "8.5" {
		t.Errorf("apache = %q", v)
	}
	if v := vhostPHPVersion([]byte("fastcgi_pass php-fpm-7.4:9000;")); v != "7.4" {
		t.Errorf("nginx = %q", v)
	}
	if v := vhostPHPVersion([]byte("extprocessor lsphp {}")); v != "" {
		t.Errorf("litespeed = %q", v)
	}
}
