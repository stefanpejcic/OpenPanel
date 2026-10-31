package account

import (
	"os"
	"path/filepath"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

// the old check looked for "Unrecognized" and "password" in the translated text, which misses in any other language
func TestLoginErrorFieldsTranslated(t *testing.T) {
	dir := t.TempDir()
	po := `msgid ""
msgstr ""
"Content-Type: text/plain; charset=UTF-8\n"

msgid "Unrecognized account. Please check username."
msgstr "Unbekanntes Konto. Bitte Benutzernamen prüfen."

msgid "Invalid password. Please try again."
msgstr "Falsches Kennwort. Bitte erneut versuchen."
`
	if err := os.MkdirAll(filepath.Join(dir, "de", "LC_MESSAGES"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "de", "LC_MESSAGES", "messages.po"), []byte(po), 0o644); err != nil {
		t.Fatal(err)
	}
	mgr := i18n.NewManager(dir, nil)

	for _, locale := range []string{"en", "de"} {
		tr := mgr.Translator(locale)
		if u, p := loginErrorFields(tr, tr.Get("Unrecognized account. Please check username.")); !u || p {
			t.Errorf("%s: unknown user gave unrecognized=%v password=%v", locale, u, p)
		}
		if u, p := loginErrorFields(tr, tr.Get("Invalid password. Please try again.")); u || !p {
			t.Errorf("%s: wrong password gave unrecognized=%v password=%v", locale, u, p)
		}
		// other messages mark no field, even when they mention the password
		if u, p := loginErrorFields(tr, tr.Get("Username and password are required.")); u || p {
			t.Errorf("%s: a general message marked a field", locale)
		}
	}
}
