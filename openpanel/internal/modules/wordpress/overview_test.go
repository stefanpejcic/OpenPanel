package wordpress

import "testing"

func TestAdminSummary(t *testing.T) {
	first, n := adminSummary([]wpUser{
		{Username: "zoe", Role: "administrator"},
		{Username: "bob", Role: "editor"},
		{Username: "anna", Role: "administrator"},
	})
	if first != "anna" || n != 2 {
		t.Errorf("adminSummary = %q, %d; want anna, 2", first, n)
	}
	if first, n := adminSummary(nil); first != "" || n != 0 {
		t.Errorf("no users = %q, %d", first, n)
	}
}

func TestCloneTargetName(t *testing.T) {
	cases := map[[2]string]string{
		{"example.com", ""}:       "example.com",
		{"example.com", "blog"}:   "example.com/blog",
		{"example.com", "/blog/"}: "example.com/blog",
		{"example.com", "  /  "}:  "example.com",
		{"example.com", "a/b"}:    "example.com/a/b",
	}
	for in, want := range cases {
		if got := cloneTargetName(in[0], in[1]); got != want {
			t.Errorf("cloneTargetName(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
}

// the Overview card reads both from one read-only GET
func TestOverviewIsARead(t *testing.T) {
	if !wpToolReadActions["overview"] || wpToolWriteActions["overview"] {
		t.Error("overview should be a GET-only tool action")
	}
}
