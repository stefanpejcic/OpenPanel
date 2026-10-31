package emails

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestParseEmailListRow(t *testing.T) {
	row := parseEmailListRow("* info@demo.rs ( 0 / 2.0G ) [45%]")
	if row.Address != "info@demo.rs" {
		t.Errorf("Address = %q", row.Address)
	}
	if row.QuotaDisplay != "( 0 / 2.0G ) [45%]" {
		t.Errorf("QuotaDisplay = %q", row.QuotaDisplay)
	}
	if row.PercentVal != 45 {
		t.Errorf("PercentVal = %d, want 45", row.PercentVal)
	}
	if row.CappedVal != 45 {
		t.Errorf("CappedVal = %d, want 45", row.CappedVal)
	}
	if row.BarColor != "bg-green-500" {
		t.Errorf("BarColor = %q, want bg-green-500", row.BarColor)
	}
}

func TestParseEmailListRowBarColors(t *testing.T) {
	cases := []struct {
		percent int
		color   string
	}{
		{80, "bg-green-500"},
		{81, "bg-orange-500"},
		{90, "bg-orange-500"},
		{91, "bg-red-500"},
		{100, "bg-red-500"},
	}
	for _, c := range cases {
		line := "* a@b.com ( 0 / 1G ) [" + strconv.Itoa(c.percent) + "%]"
		row := parseEmailListRow(line)
		if row.BarColor != c.color {
			t.Errorf("percent=%d: BarColor = %q, want %q", c.percent, row.BarColor, c.color)
		}
	}
}

func TestParseEmailListRowNoMatch(t *testing.T) {
	row := parseEmailListRow("* a@b.com ( 0 / 1G )")
	if row.PercentVal != 0 {
		t.Errorf("PercentVal = %d, want 0 (no bracket present)", row.PercentVal)
	}
}

func TestParseSingleEmailQuota(t *testing.T) {
	q := parseSingleEmailQuota("* info@demo.rs ( 0 / 2.0G ) [45%]")
	if q.Address != "info@demo.rs" {
		t.Errorf("Address = %q", q.Address)
	}
	if q.PercentVal != 45 {
		t.Errorf("PercentVal = %d, want 45", q.PercentVal)
	}
	if q.AllocatedQuota != "2.0" || q.AllocatedValue != "G" {
		t.Errorf("AllocatedQuota=%q AllocatedValue=%q, want 2.0/G", q.AllocatedQuota, q.AllocatedValue)
	}
}

func TestParseSingleEmailQuotaCapped(t *testing.T) {
	q := parseSingleEmailQuota("* a@b.com ( 0 / 1G ) [150%]")
	if q.PercentVal != 150 {
		t.Errorf("PercentVal = %d, want 150 (uncapped)", q.PercentVal)
	}
	if q.CappedVal != 100 {
		t.Errorf("CappedVal = %d, want 100", q.CappedVal)
	}
}

func TestIsValidEmail(t *testing.T) {
	valid := []string{"user@example.com", "a.b+c%d-e@sub.example.co", "x_y@z.io"}
	for _, v := range valid {
		if !isValidEmail(v) {
			t.Errorf("expected %q to be valid", v)
		}
	}
	invalid := []string{"", "not-an-email", "user@", "@example.com", "user@example", "user example@x.com"}
	for _, v := range invalid {
		if isValidEmail(v) {
			t.Errorf("expected %q to be invalid", v)
		}
	}
}

func TestIsIPv4(t *testing.T) {
	valid := []string{"1.2.3.4", "255.255.255.255", "0.0.0.0"}
	for _, v := range valid {
		if !isIPv4(v) {
			t.Errorf("expected %q to be valid ipv4", v)
		}
	}
	invalid := []string{"256.1.1.1", "1.2.3", "1.2.3.4.5", "example.com", ""}
	for _, v := range invalid {
		if isIPv4(v) {
			t.Errorf("expected %q to be invalid ipv4", v)
		}
	}
}

func TestQuotaToBytes(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"1B", 1}, {"1K", 1024}, {"1M", 1024 * 1024}, {"1G", 1024 * 1024 * 1024}, {"2T", 2 * 1024 * 1024 * 1024 * 1024},
	}
	for _, c := range cases {
		got, err := quotaToBytes(c.in)
		if err != nil {
			t.Errorf("quotaToBytes(%q) error: %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("quotaToBytes(%q) = %v, want %v", c.in, got, c.want)
		}
	}
	if _, err := quotaToBytes("1X"); err == nil {
		t.Error("expected error for invalid unit")
	}
}

func TestParseMaxQuota(t *testing.T) {
	n, u := parseMaxQuota("0")
	if n != 0 || u != "T" {
		t.Errorf("parseMaxQuota(0) = %v/%v, want 0/T", n, u)
	}
	n, u = parseMaxQuota("5G")
	if n != 5 || u != "G" {
		t.Errorf("parseMaxQuota(5G) = %v/%v, want 5/G", n, u)
	}
	n, u = parseMaxQuota("2.5m")
	if n != 2.5 || u != "M" {
		t.Errorf("parseMaxQuota(2.5m) = %v/%v, want 2.5/M", n, u)
	}
}

func TestEmailQuotaToast(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		id, msg := emailQuotaToast([]EmailListRow{{Address: "a@b.com", PercentVal: 50}})
		if id != "" || msg != "" {
			t.Errorf("got id=%q msg=%q, want empty", id, msg)
		}
	})
	t.Run("one", func(t *testing.T) {
		id, msg := emailQuotaToast([]EmailListRow{{Address: "a@b.com", PercentVal: 85}})
		if id != "quota:a@b.com" {
			t.Errorf("id = %q", id)
		}
		if msg != "Email a@b.com is reaching its quota (85%)." {
			t.Errorf("msg = %q", msg)
		}
	})
	t.Run("two", func(t *testing.T) {
		id, msg := emailQuotaToast([]EmailListRow{
			{Address: "a@b.com", PercentVal: 85},
			{Address: "c@d.com", PercentVal: 90},
		})
		if id != "quota:a@b.com,c@d.com" {
			t.Errorf("id = %q", id)
		}
		if msg != "Emails a@b.com, c@d.com are reaching their quota." {
			t.Errorf("msg = %q", msg)
		}
	})
	t.Run("multiple", func(t *testing.T) {
		id, msg := emailQuotaToast([]EmailListRow{
			{Address: "a@b.com", PercentVal: 85},
			{Address: "c@d.com", PercentVal: 90},
			{Address: "e@f.com", PercentVal: 95},
		})
		if id != "quota:multiple" {
			t.Errorf("id = %q", id)
		}
		if msg != "3 emails are reaching their quota." {
			t.Errorf("msg = %q", msg)
		}
	})
}

func TestAddressesOf(t *testing.T) {
	lines := []string{"* a@b.com ( 0 / 1G ) [0%]", "* c@d.com ( 0 / 1G ) [0%]"}
	got := addressesOf(lines)
	if len(got) != 2 || got[0] != "a@b.com" || got[1] != "c@d.com" {
		t.Errorf("addressesOf = %v", got)
	}
}

func TestReadRestrictedAddresses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "postfix-send-access.cf")
	if err := os.WriteFile(path, []byte("# comment\nInfo@Demo.rs\t\tREJECT\n\nsales@demo.rs REJECT\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	set := readRestrictedAddresses(path)
	if !set["info@demo.rs"] || !set["sales@demo.rs"] || len(set) != 2 {
		t.Fatalf("unexpected set: %v", set)
	}
	if len(readRestrictedAddresses(filepath.Join(t.TempDir(), "missing"))) != 0 {
		t.Fatal("missing file should give empty set")
	}
}

func TestParseEmailListRowNearQuota(t *testing.T) {
	if !parseEmailListRow("* a@b.com ( 0 / 2.0G ) [85%]").NearQuota {
		t.Fatal("85% should be near quota")
	}
	if parseEmailListRow("* a@b.com ( 0 / 2.0G ) [80%]").NearQuota {
		t.Fatal("80% should not be near quota")
	}
}

func TestEmailQuotaFor(t *testing.T) {
	cases := []struct {
		name, gb, format, planMax string
		planNum                   float64
		unit, want                string
		ok                        bool
	}{
		{"unlimited on a plan without a limit", "0", "G", "0", 0, "", "", true},
		// this used to delete the quota and give an unlimited mailbox on a capped plan
		{"unlimited on a capped plan gets the plan max", "0", "G", "2G", 2, "G", "2G", true},
		{"under the plan max", "1", "G", "2G", 2, "G", "1G", true},
		{"over the plan max is clamped", "5", "G", "2G", 2, "G", "2G", true},
		{"no plan limit keeps the request", "5", "G", "0", 0, "", "5G", true},
		{"invalid unit", "5", "X", "2G", 2, "G", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := emailQuotaFor(c.gb, c.format, c.planMax, c.planNum, c.unit)
			if got != c.want || ok != c.ok {
				t.Fatalf("emailQuotaFor(%q, %q) = (%q, %v), want (%q, %v)", c.gb, c.format, got, ok, c.want, c.ok)
			}
		})
	}
}

func TestNewEmailQuotaError(t *testing.T) {
	const twoGB = 2 << 30
	cases := []struct {
		gb, format string
		max        float64
		bad        bool
	}{
		{"", "G", twoGB, false},
		{"2", "G", twoGB, false},
		{"1.5", "G", twoGB, false},
		{"750", "M", twoGB, false},
		{"3", "G", twoGB, true},
		{"0", "G", twoGB, true},
		{"0.0", "G", twoGB, true},
		{"0", "G", 0, false},
		{"9", "T", 0, false},
		{"-1", "G", twoGB, true},
		{"1e3", "M", twoGB, true},
		{"0x10", "M", twoGB, true},
		{"NaN", "G", 0, true},
		{"1", "X", twoGB, true},
		{"1", "K", twoGB, true},
		{"512", "k", twoGB, false},
	}
	for _, c := range cases {
		if got := newEmailQuotaError(c.gb, c.format, c.max); (got != "") != c.bad {
			t.Errorf("newEmailQuotaError(%q, %q, %v) = %q, want bad=%v", c.gb, c.format, c.max, got, c.bad)
		}
	}
}

func TestNewEmailFieldError(t *testing.T) {
	cases := []struct {
		user, pass string
		bad        bool
	}{
		{"info", "Abcdef12345!", false},
		{"first.last+tag", "Abcdef12345!", false},
		{"a@b", "Abcdef12345!", true},
		{"has space", "Abcdef12345!", true},
		{strings.Repeat("a", 61), "Abcdef12345!", true},
		{"info", "Ab1!", true},
		{"info", strings.Repeat("Ab1!", 8), true},
		{"info", "abcdefgh", true},
	}
	for _, c := range cases {
		if got := newEmailFieldError(c.user, c.pass, 50); (got != "") != c.bad {
			t.Errorf("newEmailFieldError(%q, %q) = %q, want bad=%v", c.user, c.pass, got, c.bad)
		}
	}
}

func TestPlanMaxQuotaBytes(t *testing.T) {
	for in, want := range map[string]float64{"0": 0, "": 0, "2G": 2 << 30, "512M": 512 << 20, "1T": 1 << 40} {
		if got := planMaxQuotaBytes(in); got != want {
			t.Errorf("planMaxQuotaBytes(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestQuotaPresets(t *testing.T) {
	labels := func(ps []QuotaPreset) string {
		var out []string
		for _, p := range ps {
			out = append(out, p.Value+p.Unit+"="+p.Label)
		}
		return strings.Join(out, ",")
	}
	cases := []struct {
		num  float64
		unit string
		want string
	}{
		{2, "G", "256M=256 MB,512M=512 MB,1G=1 GB,2G=2 GB"},
		{50, "G", "1G=1 GB,2G=2 GB,5G=5 GB,10G=10 GB,50G=50 GB"},
		{500, "M", "256M=256 MB,500M=500 MB"},
		{0, "T", "256M=256 MB,512M=512 MB,1G=1 GB,2G=2 GB,0G=Unlimited"},
	}
	for _, c := range cases {
		maxBytes := c.num * quotaUnitMultipliers[c.unit[0]]
		if c.num == 0 {
			maxBytes = 0
		}
		ps := quotaPresets(c.num, c.unit, maxBytes)
		if got := labels(ps); got != c.want {
			t.Errorf("quotaPresets(%v%s) = %s, want %s", c.num, c.unit, got, c.want)
		}
		if c.num > 0 && ps[len(ps)-1].Note != "Plan maximum" {
			t.Errorf("quotaPresets(%v%s): last preset should be the plan maximum", c.num, c.unit)
		}
	}
}

func TestAddressTaken(t *testing.T) {
	lines := []string{"* info@demo.rs ( 0 / 2.0G ) [0%]", "* sales@demo.rs ( 0 / 2.0G ) [0%]"}
	if !addressTaken(lines, "Info@Demo.rs") || addressTaken(lines, "hello@demo.rs") {
		t.Error("addressTaken should match existing addresses case-insensitively and nothing else")
	}
}

func TestAddFailureReason(t *testing.T) {
	cases := []struct{ stderr, stdout, want string }{
		{"Error: 'x' already exists\n", "", "Error: 'x' already exists"},
		{"", "Error: OpenPanel Community edition does not support emails. Please consider purchasing.\nhttps://my.openpanel.com/x\n", "OpenPanel Community edition does not support emails. Please consider purchasing."},
		{"  ", "\n\nsomething odd\n", "something odd"},
		{"", "", "command failed"},
	}
	for _, c := range cases {
		if got := addFailureReason(c.stderr, c.stdout); got != c.want {
			t.Errorf("addFailureReason(%q, %q) = %q, want %q", c.stderr, c.stdout, got, c.want)
		}
	}
}
