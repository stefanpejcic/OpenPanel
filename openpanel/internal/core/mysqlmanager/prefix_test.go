package mysqlmanager

import (
	"testing"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

func TestWithPrefix(t *testing.T) {
	on := config.Config{"mysql_enforce_username_prefix": "yes"}
	cases := []struct {
		cfg             config.Config
		ctx, name, want string
	}{
		{config.Config{}, "john", "shop", "shop"},
		{config.Config{"mysql_enforce_username_prefix": ""}, "john", "shop", "shop"},
		{config.Config{"mysql_enforce_username_prefix": "no"}, "john", "shop", "shop"},
		{on, "john", "shop", "john_shop"},
		{config.Config{"mysql_enforce_username_prefix": "YES"}, "john", "shop", "john_shop"},
		{on, "john", "john_shop", "john_shop"},
		{on, "john", "", ""},
		{on, "", "shop", "shop"},
	}
	for _, c := range cases {
		if got := WithPrefix(c.cfg, c.ctx, c.name); got != c.want {
			t.Errorf("WithPrefix(%v, %q, %q) = %q, want %q", c.cfg, c.ctx, c.name, got, c.want)
		}
	}
}
