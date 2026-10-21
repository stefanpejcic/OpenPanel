package backups

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// fakeRestores swaps restoreDumpInto for one that records what would be restored
func fakeRestores(t *testing.T) map[string]string {
	got := map[string]string{}
	orig := restoreDumpInto
	restoreDumpInto = func(_ context.Context, _, engine, db string, dump io.Reader) error {
		b, err := io.ReadAll(dump)
		got[engine+"/"+db] = string(b)
		return err
	}
	t.Cleanup(func() { restoreDumpInto = orig })
	return got
}

func earlyLayoutArchive(t *testing.T) []byte {
	return makeArchive(t, []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/00-meta/mysql.json", body: []byte(`{"format":1,"engine":"mysql","databases":["wp_db","shop"]}`)},
		{name: "/backup/01-mail/a.com", body: nil},
		{name: "/backup/01-mysql", body: nil},
		{name: "/backup/01-mysql/shop.sql", body: []byte("shop dump")},
		{name: "/backup/01-mysql/wp_db.sql", body: []byte("wp dump")},
		{name: "/backup/01-postgres/analytics.sql", body: []byte("pg dump")},
		{name: "/backup/01-vhosts/a.com.conf", body: []byte("vhost a")},
		{name: "/backup/crons.ini", body: []byte("[cron]\n")},
		{name: "/backup/html/big.bin", body: randomBytes(t, 4<<20)},
	})
}

func TestRestoreDatabaseEarlyLayoutStopsAtDump(t *testing.T) {
	got := fakeRestores(t)
	archive := earlyLayoutArchive(t)
	cr := &countingReader{r: bytes.NewReader(archive)}
	restored, err := restoreDatabasesFromTar(context.Background(), "u", cr, "wp_db")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, []string{"wp_db"}) || got["mysql/wp_db"] != "wp dump" || len(got) != 1 {
		t.Errorf("restored=%v got=%v", restored, got)
	}
	if cr.n > len(archive)/2 {
		t.Errorf("read %d of %d bytes, should stop at the dump", cr.n, len(archive))
	}
}

func TestRestoreDatabaseEarlyLayoutMissingStopsBeforeHTML(t *testing.T) {
	fakeRestores(t)
	archive := earlyLayoutArchive(t)
	cr := &countingReader{r: bytes.NewReader(archive)}
	restored, err := restoreDatabasesFromTar(context.Background(), "u", cr, "nope")
	if err != nil || len(restored) != 0 {
		t.Fatalf("restored=%v err=%v", restored, err)
	}
	if cr.n > len(archive)/2 {
		t.Errorf("read %d of %d bytes, should stop past the dump dirs", cr.n, len(archive))
	}
}

func TestRestoreDatabasesAllEngines(t *testing.T) {
	got := fakeRestores(t)
	restored, err := restoreDatabasesFromTar(context.Background(), "u", bytes.NewReader(earlyLayoutArchive(t)), "")
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"shop", "wp_db", "analytics"}; !reflect.DeepEqual(restored, want) || got["postgres/analytics"] != "pg dump" {
		t.Errorf("restored=%v got=%v", restored, got)
	}
}

func TestRestoreDatabaseLegacyLayout(t *testing.T) {
	got := fakeRestores(t)
	restored, err := restoreDatabasesFromTar(context.Background(), "u", bytes.NewReader(newFormatArchive(t, true)), "analytics")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored, []string{"analytics"}) || got["postgres/analytics"] != "CREATE TABLE z (id int);\n" {
		t.Errorf("restored=%v got=%v", restored, got)
	}
}

func TestClassifyTarListingEarlyLayout(t *testing.T) {
	info := classifyTarListing("u.tar.gz", "backup/01-mysql/wp_db.sql\nbackup/01-postgres/analytics.sql\nbackup/html/index.php\n")
	if want := []string{"wp_db", "analytics"}; !reflect.DeepEqual(info.Databases, want) {
		t.Errorf("databases = %v, want %v", info.Databases, want)
	}
	if info.DatabaseEngines["analytics"] != "postgres" || info.DatabaseEngines["wp_db"] != "mysql" {
		t.Errorf("engines = %v", info.DatabaseEngines)
	}
}

func layoutTestPlan(t *testing.T) (*restorePlan, string) {
	dir := t.TempDir()
	return &restorePlan{
		sections: map[string]string{
			"backup/html":      filepath.Join(dir, "html"),
			"backup/vhosts":    filepath.Join(dir, "vhosts"),
			"backup/mail":      filepath.Join(dir, "mail"),
			"backup/crons.ini": filepath.Join(dir, "crons.ini"),
		},
		domains:   map[string]bool{"a.com": true},
		mailStore: func(d string) string { return filepath.Join(dir, "mailstore", d) },
		zonesDir:  filepath.Join(dir, "zones"),
	}, dir
}

func TestRestoreSectionEarlyLayoutStopsBeforeHTML(t *testing.T) {
	archive := earlyLayoutArchive(t)
	for _, target := range []string{"vhosts", "crons"} {
		plan, dir := layoutTestPlan(t)
		cr := &countingReader{r: bytes.NewReader(archive)}
		extracted, err := restoreFilesFromTar(cr, plan, 0, restoreSections[target].prefixes)
		if err != nil || len(extracted) != 1 {
			t.Fatalf("%s: extracted=%v err=%v", target, extracted, err)
		}
		if target == "vhosts" {
			if b, _ := os.ReadFile(filepath.Join(dir, "vhosts", "a.com.conf")); string(b) != "vhost a" {
				t.Errorf("vhost = %q", b)
			}
		}
		if cr.n > len(archive)/2 {
			t.Errorf("%s: read %d of %d bytes, should stop before html/", target, cr.n, len(archive))
		}
	}
}

func TestRestoreSectionLegacyLayout(t *testing.T) {
	archive := makeArchive(t, []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/crons.ini", body: []byte("[cron]\n")},
		{name: "/backup/html/big.bin", body: randomBytes(t, 1<<20)},
		{name: "/backup/vhosts/a.com.conf", body: []byte("old vhost")},
	})
	plan, dir := layoutTestPlan(t)
	extracted, err := restoreFilesFromTar(bytes.NewReader(archive), plan, 0, restoreSections["vhosts"].prefixes)
	if err != nil || len(extracted) != 1 {
		t.Fatalf("extracted=%v err=%v", extracted, err)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "vhosts", "a.com.conf")); string(b) != "old vhost" {
		t.Errorf("vhost = %q", b)
	}
}

func TestFullRestoreEarlyLayoutMapsDirs(t *testing.T) {
	plan, dir := layoutTestPlan(t)
	if _, err := restoreFilesFromTar(bytes.NewReader(earlyLayoutArchive(t)), plan, 0, nil); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"vhosts/a.com.conf", "crons.ini", "html/big.bin"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Errorf("%s not restored: %v", f, err)
		}
	}
	// dumps go to the db, never into a volume
	if _, err := os.Stat(filepath.Join(dir, "mysql")); err == nil {
		t.Error("dumps were written as files")
	}
}

func TestCanonicalName(t *testing.T) {
	for in, want := range map[string]string{
		"backup/01-vhosts/a.com.conf": "backup/vhosts/a.com.conf",
		"backup/01-mysql":             "backup/mysql",
		"backup/01-other/x":           "backup/01-other/x",
		"backup/html/01-mysql/x":      "backup/html/01-mysql/x",
		"backup/vhosts/a.com.conf":    "backup/vhosts/a.com.conf",
	} {
		if got, _ := canonicalName(in); got != want {
			t.Errorf("canonicalName(%q) = %q, want %q", in, got, want)
		}
	}
}
