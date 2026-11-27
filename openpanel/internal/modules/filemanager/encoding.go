package filemanager

import (
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// fallbackEncoding opens files that aren't valid UTF-8: every byte maps to one character and back, so saving keeps the bytes the user didn't touch
const fallbackEncoding = "iso-8859-1"

// editorEncodings are the encodings the file editor can open and save with, the names match the select in edit_file.html
var editorEncodings = map[string]encoding.Encoding{
	"iso-8859-1":   charmap.ISO8859_1,
	"iso-8859-2":   charmap.ISO8859_2,
	"iso-8859-5":   charmap.ISO8859_5,
	"iso-8859-7":   charmap.ISO8859_7,
	"iso-8859-9":   charmap.ISO8859_9,
	"iso-8859-15":  charmap.ISO8859_15,
	"windows-1250": charmap.Windows1250,
	"windows-1251": charmap.Windows1251,
	"windows-1252": charmap.Windows1252,
	"windows-1253": charmap.Windows1253,
	"windows-1254": charmap.Windows1254,
	"windows-1256": charmap.Windows1256,
	"koi8-r":       charmap.KOI8R,
	"koi8-u":       charmap.KOI8U,
	"shift_jis":    japanese.ShiftJIS,
	"euc-jp":       japanese.EUCJP,
	"euc-kr":       korean.EUCKR,
	"gbk":          simplifiedchinese.GBK,
	"gb18030":      simplifiedchinese.GB18030,
	"big5":         traditionalchinese.Big5,
	"utf-16le":     unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM),
	"utf-16be":     unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM),
}

var errUnknownEncoding = errors.New("unknown encoding")

// normalizeEncoding lowercases the name and maps "" and the utf8 spellings to "utf-8", anything else must be in editorEncodings
func normalizeEncoding(name string) (string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "", "utf-8", "utf8":
		return "utf-8", nil
	}
	if _, ok := editorEncodings[name]; !ok {
		return "", errUnknownEncoding
	}
	return name, nil
}

// decodeForEditor turns the file's bytes into the editor's UTF-8 text
func decodeForEditor(data []byte, name string) (string, error) {
	if name == "utf-8" {
		return string(data), nil
	}
	out, err := editorEncodings[name].NewDecoder().Bytes(data)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// encodeFromEditor turns the editor's text back into the file's encoding, it fails instead of writing "?" for characters the encoding can't store
func encodeFromEditor(text, name string) ([]byte, error) {
	if name == "utf-8" {
		return []byte(text), nil
	}
	if !utf8.ValidString(text) {
		return nil, errors.New("invalid text")
	}
	return editorEncodings[name].NewEncoder().Bytes([]byte(text))
}
