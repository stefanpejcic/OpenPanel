package i18n

import (
	"path/filepath"
	"testing"
)

func TestSentenceCaseKeyFindsTitleCaseTranslation(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "sr", "LC_MESSAGES", "messages.po"), `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"

msgid "Change Language"
msgstr "Promeni jezik"

msgid "Save"
msgstr "Sačuvaj"
`)
	mgr := NewManager(dir, nil)
	if got := mgr.Get("sr", "Change language"); got != "Promeni jezik" {
		t.Errorf("sentence-case key: got %q", got)
	}
	if got := mgr.Get("sr", "Save"); got != "Sačuvaj" {
		t.Errorf("exact key: got %q", got)
	}
	if got := mgr.Get("sr", "Not in the catalog"); got != "Not in the catalog" {
		t.Errorf("missing key: got %q", got)
	}
	if got := mgr.Get("en", "Change language"); got != "Change language" {
		t.Errorf("english keeps the key as written: got %q", got)
	}
}
