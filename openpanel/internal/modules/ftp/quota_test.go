package ftp

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

func TestNormalizeQuota(t *testing.T) {
	cases := []struct {
		value, unit, want string
		ok                bool
	}{
		{"", "M", "unlimited", true},
		{"0", "G", "unlimited", true},
		{"unlimited", "", "unlimited", true},
		{"500", "M", "500M", true},
		{"500", "", "500M", true},
		{"2", "G", "2G", true},
		{"2g", "M", "2G", true},
		{"750M", "G", "750M", true},
		{" 1 GB ", "", "1G", true},
		{"abc", "M", "", false},
		{"-5", "M", "", false},
		{"1.5", "G", "", false},
		{"10T", "", "", false},
		{"500; rm -rf /", "M", "", false},
	}
	for _, c := range cases {
		got, ok := normalizeQuota(c.value, c.unit)
		if got != c.want || ok != c.ok {
			t.Errorf("normalizeQuota(%q, %q) = %q, %v; want %q, %v", c.value, c.unit, got, ok, c.want, c.ok)
		}
	}
}

func TestRawQuotaString(t *testing.T) {
	cases := map[string]string{`500`: "500", `"2G"`: "2G", `null`: "", ``: ""}
	for in, want := range cases {
		if got := rawQuotaString(json.RawMessage(in)); got != want {
			t.Errorf("rawQuotaString(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoadAccountsWithQuota(t *testing.T) {
	dir := t.TempDir()
	list := filepath.Join(dir, "users.list")
	content := "bob@example.com|pass|/var/www/html/|1000|1000|500\n" +
		"alice@example.com|pass|/var/www/html/alice|1000|1000\n" +
		"carol@example.com|pass|/var/www/html/carol|1000|1000|bogus\n" +
		"dave@example.com|pass|/var/www/html/dave|1000|1000|1\n"
	if err := os.WriteFile(list, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	usage := "bob@example.com|256000|500|0\ndave@example.com|2048|1|1\nalice@example.com|10|0|0\n"
	if err := os.WriteFile(filepath.Join(dir, "quota.usage"), []byte(usage), 0o644); err != nil {
		t.Fatal(err)
	}

	accounts := loadAccounts(list)
	if len(accounts) != 4 {
		t.Fatalf("expected 4 accounts, got %d", len(accounts))
	}
	bob, alice, carol, dave := accounts[0], accounts[1], accounts[2], accounts[3]

	if bob.QuotaMB != 500 || !bob.HasUsage || bob.UsedKB != 256000 || bob.OverQuota {
		t.Errorf("bob: %+v", bob)
	}
	if bob.UsedPercent() != 50 || bob.QuotaLabel() != "500 MB" || bob.UsedLabel() != "250.0 MB" {
		t.Errorf("bob labels: %d %q %q", bob.UsedPercent(), bob.QuotaLabel(), bob.UsedLabel())
	}
	// unlimited accounts never show usage, even if a stale row is left in quota.usage
	if alice.HasQuota() || alice.HasUsage {
		t.Errorf("alice should be unlimited: %+v", alice)
	}
	if carol.HasQuota() {
		t.Errorf("carol has an invalid quota field and should be unlimited: %+v", carol)
	}
	if !dave.OverQuota || dave.UsedPercent() != 100 {
		t.Errorf("dave should be over quota and capped at 100%%: %+v %d", dave, dave.UsedPercent())
	}
}

func TestQuotaLabels(t *testing.T) {
	if got := formatMB(2048); got != "2 GB" {
		t.Errorf("formatMB(2048) = %q", got)
	}
	if got := formatMB(1536); got != "1.5 GB" {
		t.Errorf("formatMB(1536) = %q", got)
	}
	if got := formatKB(512); got != "512 KB" {
		t.Errorf("formatKB(512) = %q", got)
	}
	if got := formatKB(3 * 1024 * 1024); got != "3.00 GB" {
		t.Errorf("formatKB(3G) = %q", got)
	}

	for _, c := range []struct {
		mb          int
		value, unit string
	}{{0, "", "M"}, {500, "500", "M"}, {2048, "2", "G"}, {1536, "1536", "M"}} {
		v, u := Account{QuotaMB: c.mb}.QuotaInput()
		if v != c.value || u != c.unit {
			t.Errorf("QuotaInput(%d) = %q %q, want %q %q", c.mb, v, u, c.value, c.unit)
		}
	}
}

func TestFTPBulkActionsIncludeQuota(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	keys := map[string]bool{}
	for _, action := range ftpBulkActions(mgr.Translator("en")) {
		keys[action.Key] = true
	}
	if !keys["quota"] || !keys["quota_remove"] {
		t.Errorf("expected quota and quota_remove bulk actions, got %v", keys)
	}
}

func TestRenderFTPQuotaPage(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	acc := Account{Username: "bob@example.com", Path: "/var/www/html/", QuotaMB: 2048, UsedKB: 1024, HasUsage: true}
	value, unit := acc.QuotaInput()
	data := FTPQuotaPageData{LayoutData: baseLayout(mgr, "/ftp/quota/bob@example.com"), Account: acc, QuotaValue: value, QuotaUnit: unit}
	w := httptest.NewRecorder()
	if err := ftpQuotaPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"bob@example.com", `name="quota"`, `value="2"`, `<option value="G" selected>`, "2 GB", "1.0 MB"} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in quota page", want)
		}
	}
}

func TestRenderFTPAccountsPageQuotaColumn(t *testing.T) {
	mgr := i18n.NewManager(t.TempDir(), nil)
	data := FTPAccountsPageData{
		LayoutData: baseLayout(mgr, "/ftp"), FTPHost: "1.2.3.4", FTPPort: "21",
		Accounts: []Account{
			{Username: "full@example.com", Path: "/var/www/html/", QuotaMB: 1, UsedKB: 2048, HasUsage: true, OverQuota: true},
			{Username: "free@example.com", Path: "/var/www/html/"},
		},
	}
	w := httptest.NewRecorder()
	if err := ftpAccountsPage.Render(w, 200, data); err != nil {
		t.Fatalf("Render: %v", err)
	}
	body := w.Body.String()
	for _, want := range []string{"Over quota", "Unlimited", "/ftp/quota/full@example.com", "2.0 MB / 1 MB", "width: 100%", "ftpColumnsVisibility", "uid: false", "gid: false", `x-model="columns.uid"`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in accounts page", want)
		}
	}
}
