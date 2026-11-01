package appkit

import "testing"

func TestInstallDoneEvent(t *testing.T) {
	ev := InstallDoneEvent("done", "a.com/x", InstallInfo{AdminPath: "/administrator", AdminUsername: "admin"})
	if ev["site_url"] != "https://a.com/x" || ev["admin_url"] != "https://a.com/x/administrator" || ev["admin_username"] != "admin" {
		t.Fatalf("unexpected event: %v", ev)
	}
	if _, ok := ev["db_name"]; ok {
		t.Fatal("empty db_name should be left out")
	}
}
