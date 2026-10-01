package waf

import "regexp"

// engineMonitor is Coraza's monitor mode: rules run and get logged but nothing is blocked
const engineMonitor = "DetectionOnly"

var engineRE = regexp.MustCompile(`(?m)^(\s*SecRuleEngine\s+)(On|Off|DetectionOnly)\b`)

// engineStatus is the domain's SecRuleEngine value: On, Off, DetectionOnly, or Unknown
func engineStatus(content string) string {
	if m := engineRE.FindStringSubmatch(content); m != nil {
		return m[2]
	}
	return "Unknown"
}

// setEngine sets SecRuleEngine in every directives block of a domain conf
func setEngine(content, value string) string {
	return engineRE.ReplaceAllString(content, "${1}"+value)
}

func validEngine(v string) bool {
	return v == "On" || v == "Off" || v == engineMonitor
}
