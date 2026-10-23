package backups

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"sort"
	"testing"
)

func downloadArchive(t *testing.T) []byte {
	return makeArchive(t, []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/emails/a.com/info/new/1", body: []byte("hello info")},
		{name: "/backup/emails/a.com/sales/new/1", body: []byte("hello sales")},
		{name: "/backup/emails/b.com/x/new/1", body: []byte("hello x")},
		{name: "/backup/html/big.bin", body: randomBytes(t, 2<<20)},
		// sorted like docker-volume-backup writes them, the early stop relies on it
		{name: "/backup/mysql/shop.sql", body: []byte("CREATE TABLE shop;")},
		{name: "/backup/mysql/wp.sql", body: []byte("CREATE TABLE wp;")},
	})
}

func TestStreamItemsSingleDump(t *testing.T) {
	targets, err := downloadTargets("databases", []string{"mysql/shop"})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	sent, err := streamItems(w, bytes.NewReader(downloadArchive(t)), "databases", "u-2026.tar.gz", targets)
	if err != nil || sent != 1 {
		t.Fatalf("sent=%d err=%v", sent, err)
	}
	if w.Body.String() != "CREATE TABLE shop;" || w.Header().Get("Content-Disposition") != `attachment; filename="shop.sql"` {
		t.Errorf("got %q, %s", w.Body.String(), w.Header().Get("Content-Disposition"))
	}
}

func TestStreamItemsPacksMailboxes(t *testing.T) {
	targets, err := downloadTargets("mailboxes", []string{"a.com/info", "a.com/sales"})
	if err != nil {
		t.Fatal(err)
	}
	archive := downloadArchive(t)
	cr := &countingReader{r: bytes.NewReader(archive)}
	w := httptest.NewRecorder()
	if _, err := streamItems(w, cr, "mailboxes", "u-2026.tar.gz", targets); err != nil {
		t.Fatal(err)
	}
	gz, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatal(err)
	}
	var files []string
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if h.Typeflag == tar.TypeReg {
			files = append(files, h.Name)
		}
	}
	sort.Strings(files)
	if len(files) != 2 || files[0] != "emails/a.com/info/new/1" || files[1] != "emails/a.com/sales/new/1" {
		t.Errorf("files = %v", files)
	}
	if cr.n > len(archive)/2 {
		t.Errorf("read %d of %d bytes, should stop after the mailboxes", cr.n, len(archive))
	}
}

func TestDownloadTargetsRejectsBadItems(t *testing.T) {
	for kind, items := range map[string][]string{
		"databases": {"mysql/../etc"},
		"mailboxes": {"a.com/.."},
		"paths":     {"../x"},
		"other":     {"x"},
	} {
		if _, err := downloadTargets(kind, items); err == nil {
			t.Errorf("%s %v should be refused", kind, items)
		}
	}
}

func TestStreamItemsSingleDumpEarlyLayoutStopsBeforeHTML(t *testing.T) {
	archive := makeArchive(t, []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/00-meta/mysql.json", body: []byte(`{"format":1,"engine":"mysql","databases":["shop"]}`)},
		{name: "/backup/01-mysql/shop.sql", body: []byte("CREATE TABLE shop;")},
		{name: "/backup/html/big.bin", body: randomBytes(t, 2<<20)},
	})
	targets, err := downloadTargets("databases", []string{"mysql/shop"})
	if err != nil {
		t.Fatal(err)
	}
	cr := &countingReader{r: bytes.NewReader(archive)}
	w := httptest.NewRecorder()
	sent, err := streamItems(w, cr, "databases", "u-2026.tar.gz", targets)
	if err != nil || sent != 1 || w.Body.String() != "CREATE TABLE shop;" {
		t.Fatalf("sent=%d err=%v body=%q", sent, err, w.Body.String())
	}
	if cr.n > len(archive)/2 {
		t.Errorf("read %d of %d bytes, should stop at the dump", cr.n, len(archive))
	}
}

func TestStreamItemsDomainEarlyLayoutStopsBeforeHTML(t *testing.T) {
	archive := earlyLayoutArchive(t)
	targets, err := downloadTargets("domains", []string{"a.com"})
	if err != nil {
		t.Fatal(err)
	}
	cr := &countingReader{r: bytes.NewReader(archive)}
	w := httptest.NewRecorder()
	sent, err := streamItems(w, cr, "domains", "u-2026.tar.gz", targets)
	if err != nil || sent != 1 {
		t.Fatalf("sent=%d err=%v", sent, err)
	}
	gz, _ := gzip.NewReader(w.Body)
	h, err := tar.NewReader(gz).Next()
	if err != nil || h.Name != "vhosts/a.com.conf" {
		t.Errorf("got %v %v, want vhosts/a.com.conf", h, err)
	}
	if cr.n > len(archive)/2 {
		t.Errorf("read %d of %d bytes, should stop before html/", cr.n, len(archive))
	}
}

// a destination that drops mid-download used to end as a clean, complete looking archive
func TestStreamItemsReportsCutOffArchive(t *testing.T) {
	archive := makeArchive(t, []tarEntry{
		{name: "/backup", body: nil},
		{name: "/backup/html/a.bin", body: randomBytes(t, 1<<20)},
		{name: "/backup/html/b.bin", body: randomBytes(t, 1<<20)},
	})
	targets, err := downloadTargets("paths", []string{"a.bin", "b.bin"})
	if err != nil {
		t.Fatal(err)
	}
	for _, cut := range []int{len(archive) / 4, len(archive) * 3 / 4} {
		w := httptest.NewRecorder()
		sent, err := streamItems(w, bytes.NewReader(archive[:cut]), "paths", "u-2026.tar.gz", targets)
		if err == nil || sent == 0 {
			t.Errorf("cut at %d of %d: sent=%d err=%v, want an error after output started", cut, len(archive), sent, err)
		}
	}

	w := httptest.NewRecorder()
	if sent, err := streamItems(w, bytes.NewReader(archive), "paths", "u-2026.tar.gz", targets); err != nil || sent != 2 {
		t.Fatalf("whole archive: sent=%d err=%v", sent, err)
	}
}
