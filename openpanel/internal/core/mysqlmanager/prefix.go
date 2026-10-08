package mysqlmanager

import (
	"strings"

	"gist.github.com/stefanpejcic/openpanel/internal/core/config"
)

// Prefix is "<context>_" when openpanel.config has mysql_enforce_username_prefix=yes, empty otherwise
func Prefix(cfg config.Config, userContext string) string {
	if userContext == "" || !strings.EqualFold(strings.TrimSpace(cfg.Get("mysql_enforce_username_prefix", "")), "yes") {
		return ""
	}
	return userContext + "_"
}

// WithPrefix puts the enforced prefix on a new db/user name, leaving it alone if it's already there
func WithPrefix(cfg config.Config, userContext, name string) string {
	p := Prefix(cfg, userContext)
	if p == "" || name == "" || strings.HasPrefix(name, p) {
		return name
	}
	return p + name
}
