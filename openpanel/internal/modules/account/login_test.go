package account

import (
	"reflect"
	"testing"
)

func TestLocaleOptions(t *testing.T) {
	got := localeOptions([]string{"en", "de", "zh", "uk", "sr"})
	want := []localeOption{
		{Code: "en", FlagCode: "gb"},
		{Code: "de", FlagCode: "de"},
		{Code: "zh", FlagCode: "cn"},
		{Code: "uk", FlagCode: "ua"},
		{Code: "sr", FlagCode: "sr"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("localeOptions() = %+v, want %+v", got, want)
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := firstNonEmpty("", "", "c"); got != "c" {
		t.Errorf("firstNonEmpty = %q, want c", got)
	}
	if got := firstNonEmpty("a", "b"); got != "a" {
		t.Errorf("firstNonEmpty = %q, want a", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Errorf("firstNonEmpty = %q, want empty", got)
	}
}

func TestValidAutologinUsername(t *testing.T) {
	valid := []string{"john", "john.doe", "john_doe-2"}
	invalid := []string{"john doe", "john/doe", "../etc/passwd", ""}

	for _, u := range valid {
		if !validAutologinUsername.MatchString(u) {
			t.Errorf("expected %q to be a valid autologin username", u)
		}
	}
	for _, u := range invalid {
		if validAutologinUsername.MatchString(u) {
			t.Errorf("expected %q to be rejected", u)
		}
	}
}
