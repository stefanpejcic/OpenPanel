package account

import (
	"testing"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestLogoutTarget(t *testing.T) {
	cases := []struct {
		cfg  config.Config
		want string
	}{
		{config.Config{}, "/login"},
		{config.Config{"logout_url": "https://example.com/bye"}, "https://example.com/bye"},
		{config.Config{"logout_redirect": "https://old.example.com"}, "https://old.example.com"},
		{config.Config{"logout_url": " ", "logout_redirect": "https://old.example.com"}, "https://old.example.com"},
		{config.Config{"logout_url": "https://new.example.com", "logout_redirect": "https://old.example.com"}, "https://new.example.com"},
	}
	for _, c := range cases {
		if got := logoutTarget(&appctx.App{Config: c.cfg}); got != c.want {
			t.Errorf("%v: got %q, want %q", c.cfg, got, c.want)
		}
	}
}
