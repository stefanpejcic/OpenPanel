package botcontrol

import (
	"regexp"
	"strconv"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// botControlConfPath holds the crawler snippets, the admin sets the limit there and the panel only reads it
var botControlConfPath = "/etc/openpanel/caddy/botcontrol.conf"

const defaultRateLimit = 10

var rateLimitEventsRE = regexp.MustCompile(`(?s)\(botcontrol_ratelimit\)\s*\{.*?\bevents[ \t]+(\d+)`)

// RateLimit is the requests per minute each AI crawler gets on a rate-limited domain, 10 when the file has no usable value
func RateLimit() int {
	b, err := tenantfs.ReadFile(botControlConfPath)
	if err != nil {
		return defaultRateLimit
	}
	m := rateLimitEventsRE.FindSubmatch(b)
	if m == nil {
		return defaultRateLimit
	}
	n, err := strconv.Atoi(string(m[1]))
	if err != nil || n < 1 {
		return defaultRateLimit
	}
	return n
}
