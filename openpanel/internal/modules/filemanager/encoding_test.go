package filemanager

import (
	"bytes"
	"testing"
)

func TestNormalizeEncoding(t *testing.T) {
	for in, want := range map[string]string{"": "utf-8", "UTF8": "utf-8", " Windows-1251 ": "windows-1251", "koi8-r": "koi8-r"} {
		if got, err := normalizeEncoding(in); err != nil || got != want {
			t.Errorf("normalizeEncoding(%q) = %q, %v, want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"latin9000", "../etc", "utf-7"} {
		if _, err := normalizeEncoding(bad); err == nil {
			t.Errorf("normalizeEncoding(%q) should fail", bad)
		}
	}
}

func TestEditorEncodingRoundTrip(t *testing.T) {
	// "Привет" in Windows-1251
	cp1251 := []byte{0xcf, 0xf0, 0xe8, 0xe2, 0xe5, 0xf2}
	text, err := decodeForEditor(cp1251, "windows-1251")
	if err != nil || text != "Привет" {
		t.Fatalf("decode windows-1251 = %q, %v", text, err)
	}
	back, err := encodeFromEditor(text+"!", "windows-1251")
	if err != nil || !bytes.Equal(back, append(cp1251, '!')) {
		t.Fatalf("encode windows-1251 = %x, %v", back, err)
	}

	// the fallback must give back every byte it was given, that's what makes saving a non-UTF-8 file safe
	all := make([]byte, 256)
	for i := range all {
		all[i] = byte(i)
	}
	text, err = decodeForEditor(all, fallbackEncoding)
	if err != nil {
		t.Fatal(err)
	}
	back, err = encodeFromEditor(text, fallbackEncoding)
	if err != nil || !bytes.Equal(back, all) {
		t.Fatalf("%s round trip lost bytes: %v", fallbackEncoding, err)
	}

	// a character the encoding can't hold is an error, not a silent "?"
	if _, err := encodeFromEditor("Привет ქართული", "windows-1251"); err == nil {
		t.Error("Georgian text should not encode to windows-1251")
	}

	// utf-8 is passed through untouched
	if text, _ := decodeForEditor([]byte("ქართული"), "utf-8"); text != "ქართული" {
		t.Errorf("utf-8 decode changed the text: %q", text)
	}
}

func TestEveryEditorEncodingRoundTripsASCII(t *testing.T) {
	for name := range editorEncodings {
		b, err := encodeFromEditor("hello <?php echo 1; ?>\n", name)
		if err != nil {
			t.Errorf("%s: encode: %v", name, err)
			continue
		}
		text, err := decodeForEditor(b, name)
		if err != nil || text != "hello <?php echo 1; ?>\n" {
			t.Errorf("%s: round trip = %q, %v", name, text, err)
		}
	}
}
