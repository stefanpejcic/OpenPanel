package emails

import (
	"strconv"
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
