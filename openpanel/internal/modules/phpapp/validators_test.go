package phpapp

import "testing"

func TestIsValidSubdirectory(t *testing.T) {
	valid := []string{"", "blog", "app/sub"}
	invalid := []string{"../etc", "/absolute", "a/../b"}
	for _, s := range valid {
		if !isValidSubdirectory(s) {
			t.Errorf("expected %q to be a valid subdirectory", s)
		}
	}
	for _, s := range invalid {
		if isValidSubdirectory(s) {
			t.Errorf("expected %q to be an invalid subdirectory", s)
		}
	}
}

func TestIsArchiveURL(t *testing.T) {
	valid := []string{
		"https://example.com/project.zip",
		"https://example.com/project.tar.gz",
		"https://example.com/project.tgz",
		"https://example.com/project.tar",
	}
	invalid := []string{
		"", "laravel/laravel", "http://example.com/project.zip",
		"https://example.com/project.rar", "https://example.com/project.zip\nevil",
	}
	for _, s := range valid {
		if !isArchiveURL(s) {
			t.Errorf("expected %q to be a valid archive URL", s)
		}
	}
	for _, s := range invalid {
		if isArchiveURL(s) {
			t.Errorf("expected %q to be an invalid archive URL", s)
		}
	}
}

func TestIsValidInitialProject(t *testing.T) {
	valid := []string{"", "laravel/laravel", "https://example.com/project.zip"}
	invalid := []string{"@evil", "laravel", "laravel/laravel/extra", "javascript:alert(1)", "a\nb"}
	for _, s := range valid {
		if !isValidInitialProject(s) {
			t.Errorf("expected %q to be a valid initial project", s)
		}
	}
	for _, s := range invalid {
		if isValidInitialProject(s) {
			t.Errorf("expected %q to be an invalid initial project", s)
		}
	}
}
