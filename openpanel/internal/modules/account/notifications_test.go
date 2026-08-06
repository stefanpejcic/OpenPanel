package account

import "testing"

func TestNotificationLabel(t *testing.T) {
	if got := notificationLabel("notify_password_change"); got != " password change" {
		t.Errorf("got %q", got)
	}
	if got := notificationLabel("some_other_key"); got != "Some Other Key" {
		t.Errorf("got %q", got)
	}
}

func TestTitleCase(t *testing.T) {
	if got := titleCase("hello world"); got != "Hello World" {
		t.Errorf("got %q", got)
	}
	if got := titleCase(""); got != "" {
		t.Errorf("got %q", got)
	}
}
