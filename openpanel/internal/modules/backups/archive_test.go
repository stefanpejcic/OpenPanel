package backups

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"io"
	"reflect"
	"testing"
)

type tarEntry struct {
	name string
	body []byte
}

// makeArchive builds a .tar.gz the way docker-volume-backup does, leading slash and directory entries included
func makeArchive(t *testing.T, entries []tarEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, e := range entries {
		h := &tar.Header{Name: e.name, Mode: 0o644, Size: int64(len(e.body)), Typeflag: tar.TypeReg}
		if e.body == nil {
			h = &tar.Header{Name: e.name, Mode: 0o755, Typeflag: tar.TypeDir}
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(e.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return b
}

type countingReader struct {
	r io.Reader
	n int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

func newFormatArchive(t *testing.T, withMySQL bool) []byte {
	entries := []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/00-meta", body: nil},
	}
	if withMySQL {
		entries = append(entries, tarEntry{name: "/backup/00-meta/mysql.json", body: []byte(`{"format":1,"engine":"mysql","databases":["wp_db","shop"],"sizes":{"wp_db":25,"shop":25}}`)})
	}
	entries = append(entries,
		tarEntry{name: "/backup/00-meta/postgres.json", body: []byte(`{"format":1,"engine":"postgres","databases":["analytics"]}`)},
		tarEntry{name: "/backup/crons.ini", body: []byte("[cron]\n")},
		tarEntry{name: "/backup/html", body: nil},
		tarEntry{name: "/backup/html/big.bin", body: randomBytes(t, 4<<20)},
		tarEntry{name: "/backup/mysql", body: nil},
		tarEntry{name: "/backup/mysql/wp_db.sql", body: []byte("CREATE TABLE x (id int);\n")},
		tarEntry{name: "/backup/mysql/shop.sql", body: []byte("CREATE TABLE y (id int);\n")},
		tarEntry{name: "/backup/postgres", body: nil},
		tarEntry{name: "/backup/postgres/analytics.sql", body: []byte("CREATE TABLE z (id int);\n")},
	)
	return makeArchive(t, entries)
}

func TestClassifyArchiveStreamUsesManifestAndStopsEarly(t *testing.T) {
	archive := newFormatArchive(t, true)
	cr := &countingReader{r: bytes.NewReader(archive)}

	info, found, err := classifyArchiveStream("u-2026.tar.gz", cr, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Fatal("expected the manifest to be used")
	}
	if want := []string{"wp_db", "shop", "analytics"}; !reflect.DeepEqual(info.Databases, want) {
		t.Errorf("databases = %v, want %v", info.Databases, want)
	}
	if want := []string{"crons", "html", "mail", "mysql", "postgres", "vhosts"}; !reflect.DeepEqual(info.Types, want) {
		t.Errorf("types = %v, want %v", info.Types, want)
	}
	if !info.HasFiles || !info.HasCrons {
		t.Errorf("expected has_files and has_crons, got %+v", info)
	}
	if want := map[string]int64{"wp_db": 25, "shop": 25}; !reflect.DeepEqual(info.DatabaseSizes, want) {
		t.Errorf("database sizes = %v, want %v (postgres.json has none)", info.DatabaseSizes, want)
	}
	if cr.n > len(archive)/4 {
		t.Errorf("read %d of %d bytes, expected to stop right after the manifest", cr.n, len(archive))
	}
}

func TestClassifyArchiveStreamWithoutMySQLMetaListsEverything(t *testing.T) {
	archive := newFormatArchive(t, false)

	info, found, err := classifyArchiveStream("u-2026.tar.gz", bytes.NewReader(archive), true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("manifest without mysql.json must not be trusted")
	}
	want := map[string]bool{"wp_db": true, "shop": true, "analytics": true}
	if len(info.Databases) != len(want) {
		t.Fatalf("databases = %v, want all three from the full listing", info.Databases)
	}

	_, found, err = classifyArchiveStream("u-2026.tar.gz", bytes.NewReader(archive), false, nil)
	if err != nil || found {
		t.Fatalf("listAll=false should report not found, got found=%v err=%v", found, err)
	}
}

func TestClassifyArchiveStreamOldArchive(t *testing.T) {
	archive := makeArchive(t, []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/crons.ini", body: []byte("")},
		{name: "/backup/html", body: nil},
		{name: "/backup/html/index.php", body: []byte("<?php")},
		{name: "/backup/mysql", body: nil},
		{name: "/backup/mysql/wp_db.sql", body: []byte("x")},
	})

	info, found, err := classifyArchiveStream("old.tar.gz", bytes.NewReader(archive), true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if found {
		t.Fatal("old archive has no manifest")
	}
	if !reflect.DeepEqual(info.Databases, []string{"wp_db"}) {
		t.Errorf("databases = %v", info.Databases)
	}
	if info.DatabaseSizes["wp_db"] != 1 {
		t.Errorf("database sizes from tar headers = %v", info.DatabaseSizes)
	}
}

type fakeStore struct {
	archives map[string][]byte
}

func (f *fakeStore) List(ctx context.Context) ([]remoteFile, error) { return nil, nil }

func (f *fakeStore) Open(ctx context.Context, name string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.archives[name])), nil
}

type fakeListerStore struct {
	fakeStore
	listed int
}

func (f *fakeListerStore) ListMembers(ctx context.Context, name string) (string, error) {
	f.listed++
	return "backup/mysql/remote_db.sql\n", nil
}

func TestClassifyArchiveFallsBackToRemoteListing(t *testing.T) {
	old := makeArchive(t, []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/html", body: nil},
		{name: "/backup/mysql/wp_db.sql", body: []byte("x")},
	})
	store := &fakeListerStore{fakeStore: fakeStore{archives: map[string][]byte{"old.tar.gz": old, "new.tar.gz": newFormatArchive(t, true)}}}

	info, err := classifyArchive(context.Background(), store, "old.tar.gz", nil)
	if err != nil {
		t.Fatal(err)
	}
	if store.listed != 1 || !reflect.DeepEqual(info.Databases, []string{"remote_db"}) {
		t.Errorf("expected the remote listing to be used, listed=%d info=%+v", store.listed, info)
	}

	if _, err := classifyArchive(context.Background(), store, "new.tar.gz", nil); err != nil {
		t.Fatal(err)
	}
	if store.listed != 1 {
		t.Error("archive with a manifest shouldn't need a remote listing")
	}
}

func TestClassifyArchiveSkipsNonArchives(t *testing.T) {
	info, err := classifyArchive(context.Background(), &fakeStore{}, "notes.txt", nil)
	if err != nil || info.HasFiles || len(info.Types) != 0 {
		t.Errorf("expected an empty entry, got %+v err=%v", info, err)
	}
}

func TestDumpDBName(t *testing.T) {
	cases := map[string]string{
		"backup/mysql/wp_db.sql":                     "wp_db",
		"backup/mysql/shop.sql.gz":                   "shop",
		"backup/mysql/wp_db_2026-01-01_00-00-00.sql": "wp_db",
		"backup/postgres/analytics.sql":              "analytics",
		"backup/mysql/readme.txt":                    "",
		"backup/mysql/.sql":                          "",
	}
	for member, want := range cases {
		got, ok := dumpDBName(member)
		if got != want || ok != (want != "") {
			t.Errorf("dumpDBName(%q) = %q, %v, want %q", member, got, ok, want)
		}
	}
}
