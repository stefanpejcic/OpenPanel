package dns

import "testing"

func TestExtractComment(t *testing.T) {
	cases := map[string]string{
		`www 14400 IN A 1.2.3.4 ; comment here`:            "comment here",
		`www 14400 IN A 1.2.3.4`:                           "",
		`@ 14400 IN TXT "v=spf1 include:_spf; -all" ; spf`: "spf",
		`@ 14400 IN TXT "has ; a semicolon inside"`:        "",
	}
	for line, want := range cases {
		if got := extractComment(line); got != want {
			t.Errorf("extractComment(%q) = %q, want %q", line, got, want)
		}
	}
}

func TestSplitMaxN(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want []string
	}{
		{"a b c d e", 4, []string{"a", "b", "c", "d", "e"}},
		{"a  b   c d e f", 4, []string{"a", "b", "c", "d", "e f"}},
		{"a b", 4, []string{"a", "b"}},
		{"", 4, nil},
	}
	for _, c := range cases {
		got := splitMaxN(c.in, c.n)
		if len(got) != len(c.want) {
			t.Errorf("splitMaxN(%q, %d) = %v, want %v", c.in, c.n, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitMaxN(%q, %d)[%d] = %q, want %q", c.in, c.n, i, got[i], c.want[i])
			}
		}
	}
}

func TestReadSerialNumber(t *testing.T) {
	lines := []string{
		"$TTL 14400\n",
		"@ IN SOA ns1.example.com. admin.example.com. (\n",
		"  2026010101 ; Serial number\n",
		"  14400 )\n",
	}
	if got := readSerialNumber(lines); got != "2026010101" {
		t.Errorf("readSerialNumber = %q, want 2026010101", got)
	}
	if got := readSerialNumber([]string{"no serial here\n"}); got != "" {
		t.Errorf("readSerialNumber with no match = %q, want empty", got)
	}
}

func TestParseZoneWithLineNumbers(t *testing.T) {
	content := "$TTL 14400\n" +
		"@ IN SOA ns1.example.com. admin.example.com. (\n" +
		"  2026010101 ; Serial number\n" +
		"  14400 )\n" +
		"@ IN NS ns1.example.com.\n" +
		"www 14400 IN A 1.2.3.4\n" +
		"txt 14400 IN TXT ( \"first part\"\n" +
		"  \"second part\" ) ; multiline comment\n"

	entries := parseZoneWithLineNumbers(content)

	var names []string
	for _, e := range entries {
		names = append(names, e.Line)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries (A + multiline TXT), got %d: %+v", len(entries), names)
	}
	if entries[0].LineNumber != 6 || entries[0].Multiline {
		t.Errorf("entry[0] (A record) = %+v", entries[0])
	}
	if !entries[1].Multiline || entries[1].LineNumber != 7 || entries[1].EndLineNumber != 8 {
		t.Errorf("entry[1] (multiline TXT) = %+v", entries[1])
	}
	if entries[1].Comment != "multiline comment" {
		t.Errorf("entry[1].Comment = %q, want %q", entries[1].Comment, "multiline comment")
	}
}

func TestBuildZoneRows(t *testing.T) {
	entries := []ZoneLineEntry{
		{LineNumber: 1, EndLineNumber: 1, Line: "@ 14400 IN SOA ns1.example.com. admin.example.com."},
		{LineNumber: 2, EndLineNumber: 2, Line: "www 14400 IN A 1.2.3.4"},
		{LineNumber: 3, EndLineNumber: 3, Line: `@ 14400 IN TXT "v=spf1 -all" ; spf record`, Comment: "spf record"},
	}
	rows := buildZoneRows(entries)
	if len(rows) != 2 {
		t.Fatalf("expected SOA row excluded, got %d rows: %+v", len(rows), rows)
	}
	if rows[0].Name != "www" || rows[0].DisplayValue != "1.2.3.4" {
		t.Errorf("rows[0] = %+v", rows[0])
	}
	// Known quirk: the quote-strip check runs against the
	// pre-comment-stripped value, so a trailing " ; comment" (which makes
	// the raw field no longer end in a bare quote) means the quotes never
	// get stripped even though the comment itself does.
	if rows[1].DisplayValue != `"v=spf1 -all"` {
		t.Errorf("rows[1].DisplayValue = %q, want %q (comment removed, quotes NOT stripped)", rows[1].DisplayValue, `"v=spf1 -all"`)
	}
}

func TestHasSubdomainLabel(t *testing.T) {
	cases := map[string]bool{
		"example.com":       false,
		"blog.example.com":  true,
		"example.co.uk":     false,
		"www.example.co.uk": true,
	}
	for domain, want := range cases {
		if got := hasSubdomainLabel(domain); got != want {
			t.Errorf("hasSubdomainLabel(%q) = %v, want %v", domain, got, want)
		}
	}
}

func TestReadLinesKeepEnds(t *testing.T) {
	if got := readLinesKeepEnds(""); got != nil {
		t.Errorf("empty content should give nil, got %v", got)
	}
	got := readLinesKeepEnds("a\nb\nc\n")
	want := []string{"a\n", "b\n", "c\n"}
	if len(got) != len(want) {
		t.Fatalf("readLinesKeepEnds = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("readLinesKeepEnds[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	gotNoTrailing := readLinesKeepEnds("a\nb")
	wantNoTrailing := []string{"a\n", "b"}
	if len(gotNoTrailing) != len(wantNoTrailing) || gotNoTrailing[1] != "b" {
		t.Errorf("readLinesKeepEnds (no trailing newline) = %v, want %v", gotNoTrailing, wantNoTrailing)
	}
}
