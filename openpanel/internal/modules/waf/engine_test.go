package waf

import (
	"os"
	"strings"
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/i18n"
)

func TestEngineStatusAndSet(t *testing.T) {
	if engineStatus(testProfilesConf) != "On" {
		t.Fatal("expected On")
	}
	monitor := setEngine(testProfilesConf, engineMonitor)
	if engineStatus(monitor) != engineMonitor || strings.Count(monitor, "SecRuleEngine DetectionOnly") != 2 {
		t.Fatalf("monitor not set in both blocks:\n%s", monitor)
	}
	if off := setEngine(monitor, "Off"); engineStatus(off) != "Off" || strings.Contains(off, "DetectionOnly") {
		t.Fatal("monitor -> off failed")
	}
	if engineStatus("nothing here") != "Unknown" || validEngine("on") || !validEngine(engineMonitor) {
		t.Fatal("unknown/valid checks")
	}
}

func TestSetWAFForDomainsIncludesMonitor(t *testing.T) {
	path, _ := withTestDomain(t)
	writeProfilesConf(t, path)
	content, _ := readFile(path)
	if err := writeFile(path, setEngine(content, engineMonitor)); err != nil {
		t.Fatal(err)
	}
	if changed, _, _ := setWAFForDomains(t.Context(), []string{"example.com"}, true); changed != 1 {
		t.Fatal("monitor domain should be switched to On by the all-domains switch")
	}
	content, _ = readFile(path)
	if engineStatus(content) != "On" {
		t.Fatal("not On")
	}
}

func TestWAFBulkActionsOrder(t *testing.T) {
	withTestPlugins(t)
	mgr := i18n.NewManager(t.TempDir(), nil)
	acts := wafBulkActions(mgr.Translator("en"))
	var keys []string
	for _, a := range acts {
		keys = append(keys, a.Key)
	}
	if strings.Join(keys, ",") != "set_level,apply_profile,enable,disable" {
		t.Fatalf("order: %v", keys)
	}
	if len(acts[0].Input.Options) != 3 || acts[0].Input.Initial() != "standard" {
		t.Fatalf("level options: %+v", acts[0].Input)
	}
	if acts[1].Input.Options[0].Value != "none" || len(acts[1].Input.Options) != 3 {
		t.Fatalf("profile options should be none + installed plugins: %+v", acts[1].Input.Options)
	}
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return string(b), err
}

func writeFile(path, content string) error { return os.WriteFile(path, []byte(content), 0o644) }
