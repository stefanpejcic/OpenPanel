package trash

import "testing"

func TestTrashLogLines(t *testing.T) {
	if got, want := restoreLogLine([]string{"/d/a"}, "File Manager"), "restored /d/a from Trash using File Manager"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := restoreLogLine([]string{"/d/a", "/d/b"}, "Restore All"), "restored 2 items from Trash using Restore All: /d/a, /d/b"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := trashDeleteLogLine("directory", []string{"/t/a"}), "permanently deleted directory /t/a using File Manager"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, want := trashDeleteLogLine("file", []string{"/t/a", "/t/b"}), "permanently deleted 2 items from Trash using File Manager: /t/a, /t/b"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestCleanTrashName(t *testing.T) {
	if got, err := cleanTrashName("../../etc/passwd"); err != nil || got != "passwd" {
		t.Errorf("cleanTrashName = %q, %v", got, err)
	}
	if _, err := cleanTrashName("/"); err == nil {
		t.Error("expected error for /")
	}
}
