package waf

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// levelRuleID is the per-domain SecAction that sets paranoia level and anomaly thresholds, it has to run before CRS 901 initialization which only fills in unset values
const levelRuleID = "10100"

// wafLevel is one protection level preset, standard is the CRS default so it writes no line
type wafLevel struct {
	Key         string
	Name        string
	Description string
	Paranoia    int
	Inbound     int
	Outbound    int
}

const defaultLevel = "standard"

var levelCatalog = []wafLevel{
	{Key: "compatibility", Name: "Compatibility", Paranoia: 1, Inbound: 10, Outbound: 8,
		Description: "Fewer false blocks, only requests that clearly look like attacks are blocked. Use it if normal visitors keep getting blocked."},
	{Key: "standard", Name: "Standard", Paranoia: 1, Inbound: 5, Outbound: 4,
		Description: "Balanced protection that suits most websites."},
	{Key: "strict", Name: "Strict", Paranoia: 2, Inbound: 5, Outbound: 4,
		Description: "Extra checks for sites that need more security. Can block some normal requests, so check the logs after turning it on."},
}

func levelByKey(key string) (wafLevel, bool) {
	for _, l := range levelCatalog {
		if l.Key == key {
			return l, true
		}
	}
	return wafLevel{}, false
}

var errUnknownLevel = errors.New("unknown protection level")

func isLevelLine(stripped string) bool {
	return strings.HasPrefix(stripped, "SecAction") && strings.Contains(stripped, "id:"+levelRuleID+",")
}

func levelLine(l wafLevel) string {
	return `SecAction "id:` + levelRuleID + `,phase:1,pass,nolog,` +
		`setvar:tx.blocking_paranoia_level=` + strconv.Itoa(l.Paranoia) + `,` +
		`setvar:tx.inbound_anomaly_score_threshold=` + strconv.Itoa(l.Inbound) + `,` +
		`setvar:tx.outbound_anomaly_score_threshold=` + strconv.Itoa(l.Outbound) + `"`
}

var (
	levelParanoiaRE = regexp.MustCompile(`blocking_paranoia_level=(\d+)`)
	levelInboundRE  = regexp.MustCompile(`inbound_anomaly_score_threshold=(\d+)`)
)

// parseWAFLevel maps the domain's level SecAction back to a preset, no line means standard and hand-edited values show as custom
func parseWAFLevel(contentStr string) string {
	for _, line := range strings.Split(contentStr, "\n") {
		stripped := strings.TrimSpace(line)
		if !isLevelLine(stripped) {
			continue
		}
		pm, im := levelParanoiaRE.FindStringSubmatch(stripped), levelInboundRE.FindStringSubmatch(stripped)
		if pm == nil || im == nil {
			return "custom"
		}
		for _, l := range levelCatalog {
			if strconv.Itoa(l.Paranoia) == pm[1] && strconv.Itoa(l.Inbound) == im[1] {
				return l.Key
			}
		}
		return "custom"
	}
	return defaultLevel
}

// levelStrength orders the presets for the dots on /server/waf, custom gets none since it can be anything
func levelStrength(key string) int {
	switch key {
	case "compatibility":
		return 1
	case "standard":
		return 2
	case "strict":
		return 3
	}
	return 0
}
