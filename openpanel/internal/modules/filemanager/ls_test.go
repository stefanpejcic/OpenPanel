package filemanager

import "testing"

func TestParseLsOutput(t *testing.T) {
	output := `total 12
drwxr-xr-x  3 user group 4096 Jan  1 12:00 documents
-rw-r--r--  1 user group  123 Jan  1 12:00 report.txt
lrwxrwxrwx  1 user group    7 Jan  1 12:00 link -> target
drwxr-xr-x  2 user group 4096 Jan  1 12:00 .
drwxr-xr-x  3 user group 4096 Jan  1 12:00 ..
-rw-r--r--  1 user group   10 Jan  1 12:00 'quoted name.txt'
`
	entries := parseLsOutput(output)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries (total/./.. excluded), got %d: %+v", len(entries), entries)
	}

	if entries[0].Name != "documents" || entries[0].Type != "directory" {
		t.Errorf("entries[0] = %+v, want directory named documents", entries[0])
	}
	if entries[1].Name != "report.txt" || entries[1].Type != "file" || entries[1].Size != "123" {
		t.Errorf("entries[1] = %+v, want file report.txt size 123", entries[1])
	}
	if entries[2].Name != "link" || entries[2].Type != "symlink" || entries[2].LinkTarget != "target" {
		t.Errorf("entries[2] = %+v, want symlink link -> target", entries[2])
	}
	if entries[3].Name != "quoted name.txt" {
		t.Errorf("entries[3].Name = %q, want unquoted %q", entries[3].Name, "quoted name.txt")
	}
}
