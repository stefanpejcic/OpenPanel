package mysql

import "testing"

func TestZeroUserDatabasesToast(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		id, msg := zeroUserDatabasesToast([]DatabaseRow{{Database: "app_db", AssignedUsers: "appuser"}})
		if id != "" || msg != "" {
			t.Errorf("got id=%q msg=%q, want both empty", id, msg)
		}
	})

	t.Run("system database excluded", func(t *testing.T) {
		id, msg := zeroUserDatabasesToast([]DatabaseRow{{Database: "mysql", AssignedUsers: "", IsSystem: true}})
		if id != "" || msg != "" {
			t.Errorf("got id=%q msg=%q, want both empty for system database", id, msg)
		}
	})

	t.Run("one", func(t *testing.T) {
		id, msg := zeroUserDatabasesToast([]DatabaseRow{{Database: "orphan_db", AssignedUsers: ""}})
		if id != "no-users:orphan_db" {
			t.Errorf("id = %q", id)
		}
		if msg != "Database orphan_db has no users assigned." {
			t.Errorf("msg = %q", msg)
		}
	})

	t.Run("two", func(t *testing.T) {
		id, msg := zeroUserDatabasesToast([]DatabaseRow{
			{Database: "orphan_a", AssignedUsers: ""},
			{Database: "orphan_b", AssignedUsers: ""},
		})
		if id != "no-users:orphan_a,orphan_b" {
			t.Errorf("id = %q", id)
		}
		if msg != "Databases orphan_a, orphan_b have no users assigned." {
			t.Errorf("msg = %q", msg)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		id, msg := zeroUserDatabasesToast([]DatabaseRow{
			{Database: "orphan_a", AssignedUsers: ""},
			{Database: "orphan_b", AssignedUsers: ""},
			{Database: "orphan_c", AssignedUsers: ""},
		})
		if id != "no-users:multiple" {
			t.Errorf("id = %q", id)
		}
		if msg != "3 databases have no users assigned." {
			t.Errorf("msg = %q", msg)
		}
	})
}

func TestAtoiDefault(t *testing.T) {
	if got := atoiDefault("42", 0); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
	if got := atoiDefault("not-a-number", 7); got != 7 {
		t.Errorf("got %d, want 7 (fallback)", got)
	}
	if got := atoiDefault("", 7); got != 7 {
		t.Errorf("got %d, want 7 (fallback for empty)", got)
	}
}
