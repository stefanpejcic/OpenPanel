package websites

import (
	"strings"
	"testing"
)

const testVCL = `vcl 4.1;

sub vcl_backend_response {
    if (!beresp.http.Cache-Control) {
        set beresp.ttl = 1h;
    }

    set beresp.grace = 6h;
    set beresp.keep = 24h;
}
`

func TestSetVCLTTLRoundTrip(t *testing.T) {
	vcl, err := setVCLTTL(testVCL, "example.com", 300)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(vcl, `bereq.http.host == "www.example.com") { set beresp.ttl = 300s; }`) {
		t.Fatalf("ttl line missing:\n%s", vcl)
	}
	if strings.Index(vcl, vclTTLEnd) > strings.Index(vcl, "set beresp.grace") {
		t.Fatalf("block must come before grace:\n%s", vcl)
	}

	vcl, _ = setVCLTTL(vcl, "other.org", 60)
	vcl, _ = setVCLTTL(vcl, "example.com", 900)
	got := parseVCLTTLs(vcl)
	if got["example.com"] != 900 || got["other.org"] != 60 || len(got) != 2 {
		t.Fatalf("unexpected ttls %v", got)
	}
	if strings.Count(vcl, vclTTLBegin) != 1 {
		t.Fatalf("block duplicated:\n%s", vcl)
	}

	vcl, _ = setVCLTTL(vcl, "example.com", 0)
	vcl, _ = setVCLTTL(vcl, "other.org", 0)
	if vcl != testVCL {
		t.Fatalf("removing all overrides should restore the original, got:\n%s", vcl)
	}
}

func TestSetVCLTTLRejectsBadInput(t *testing.T) {
	if _, err := setVCLTTL(testVCL, `x.com" || true`, 60); err == nil {
		t.Fatal("expected invalid domain error")
	}
	if _, err := setVCLTTL("vcl 4.1;\n", "example.com", 60); err == nil {
		t.Fatal("expected error for vcl without grace anchor")
	}
}

func TestSetUserIniOpcache(t *testing.T) {
	off := setUserIniOpcache("upload_max_filesize=64M\n", false)
	if !readUserIniOpcacheOffString(off) || !strings.Contains(off, "upload_max_filesize=64M") {
		t.Fatalf("unexpected: %q", off)
	}
	if again := setUserIniOpcache(off, false); again != off {
		t.Fatalf("disabling twice should be a no-op: %q", again)
	}
	on := setUserIniOpcache(off, true)
	if on != "upload_max_filesize=64M\n" {
		t.Fatalf("unexpected: %q", on)
	}
	if setUserIniOpcache(setUserIniOpcache("", false), true) != "" {
		t.Fatal("expected empty file content")
	}
}

func readUserIniOpcacheOffString(content string) bool {
	return strings.Contains(content, "\nopcache.enable=0") || strings.HasPrefix(content, "opcache.enable=0")
}

func TestParseMeasureOutput(t *testing.T) {
	out := "#cached\n0.300 200\n0.020 200\n0.030 200\n0.025 200\n#direct\n0.9 200\n0.5 200\n0.6 200\n"
	res := parseMeasureOutput(out)
	if res["cached"][0] != 25 || res["cached"][1] != 200 {
		t.Fatalf("cached = %v", res["cached"])
	}
	if res["direct"][0] != 550 {
		t.Fatalf("direct = %v", res["direct"])
	}
}

func TestParseSizeToBytes(t *testing.T) {
	cases := map[string]int64{`"0.1G"`: 107374182, "512M": 512 << 20, "64k": 64 << 10, "junk": 0}
	for in, want := range cases {
		if got := parseSizeToBytes(in); got != want {
			t.Errorf("%s: got %d want %d", in, got, want)
		}
	}
}

func TestRedisPrefix(t *testing.T) {
	s := &wpCacheSite{Site: "Example.com/blog"}
	if got := s.redisPrefix(); got != "op_example_com_blog:" {
		t.Fatalf("got %q", got)
	}
}

func TestVarnishSiteQuery(t *testing.T) {
	root := varnishSiteQuery(&wpCacheSite{Domain: "domain.rs"}, []string{"wp"})
	if root != `ReqHeader:Host ~ "^(www\\.)?domain\\.rs$" and not ReqURL ~ "^/(wp)(/|\\?|$)" and not ReqURL ~ "`+staticFileRE+`"` {
		t.Fatalf("root query: %s", root)
	}
	sub := varnishSiteQuery(&wpCacheSite{Domain: "domain.rs", Folder: "wp"}, nil)
	if !strings.Contains(sub, `and ReqURL ~ "^/wp(/|\\?|$)"`) {
		t.Fatalf("subfolder query: %s", sub)
	}
}
