package trash

import "testing"

func TestParseLsOutputTrashSkipsBookkeeping(t *testing.T) {
	out := "total 8\n" +
		"-rw-r--r-- 1 1002 1002 433 Oct 31 22:57 .trash_restore\n" +
		"-rw-r--r-- 1 1002 1002  12 Oct 31 22:57 draft.html\n"
	meta := "draft.html=/var/www/html/a.com/draft.html|deletion_date=2026-10-31 22:57:00\n"
	entries := parseLsOutputTrash(out, meta)
	if len(entries) != 1 || entries[0].Name != "draft.html" {
		t.Fatalf("expected only draft.html, got %+v", entries)
	}
}
