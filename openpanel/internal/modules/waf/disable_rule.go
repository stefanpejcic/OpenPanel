package waf

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/auth"
	"gist.github.com/stefanpejcic/openpanel/internal/core/logger"
	"gist.github.com/stefanpejcic/openpanel/internal/core/reqip"
)

// CRS blocks on the total anomaly score (949 inbound, 959 outbound) and 980 only reports it, removing those turns blocking off for the whole domain
var scoringRulePrefixes = []string{"949", "959", "980"}

func isScoringRule(ruleID string) bool {
	for _, p := range scoringRulePrefixes {
		if strings.HasPrefix(ruleID, p) {
			return true
		}
	}
	return false
}

var (
	errInvalidRuleID = errors.New("rule ID must be a number")
	errScoringRule   = errors.New("this rule only adds up the score of the other rules, disabling it would turn off blocking for the whole domain")
	errNoDomainConf  = errors.New("WAF config not found for this domain")
	errCaddyReload   = errors.New("rule disabled but reloading Caddy failed")
)

// disableRuleForDomain adds one rule ID to the domain's SecRuleRemoveById list, a no-op when it's already there
func disableRuleForDomain(ctx context.Context, domain, ruleID string) (alreadyDisabled bool, err error) {
	if !ruleIDRE.MatchString(ruleID) || ruleID == excludedRuleID {
		return false, errInvalidRuleID
	}
	if isScoringRule(ruleID) {
		return false, errScoringRule
	}
	path := domainConfigPath(domain)
	content, readErr := os.ReadFile(path)
	if readErr != nil {
		return false, errNoDomainConf
	}
	removedRules, removedTags := parseWAFRemovals(string(content))
	for _, id := range removedRules {
		if id == ruleID {
			return true, nil
		}
	}
	removedRules = append([]string{excludedRuleID}, append(removedRules, ruleID)...)
	removedTags = append([]string{excludedTag}, filterOut(removedTags, excludedTag)...)
	if writeErr := os.WriteFile(path, []byte(rewriteDirectivesBlock(string(content), removedRules, removedTags)), 0o644); writeErr != nil {
		return false, writeErr
	}
	if reloadErr := reloadCaddy(ctx); reloadErr != nil {
		return false, errCaddyReload
	}
	return false, nil
}

// handleWAFDisableRule handles POST /server/waf/{domain}/disable-rule, the "Disable" button on the WAF log page
func handleWAFDisableRule(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	domain := firstPathSegment(r.PathValue("domain"))
	userID, _ := auth.UserID(r)
	username, err := injected(a, r)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if !a.CheckDomainBelongsToUser(r.Context(), userID, domain) {
		writeJSONError(w, http.StatusForbidden, "You do not own this domain.")
		return
	}
	_ = r.ParseForm()
	ruleID := strings.TrimSpace(r.Form.Get("rule_id"))

	already, err := disableRuleForDomain(r.Context(), domain, ruleID)
	switch {
	case errors.Is(err, errInvalidRuleID), errors.Is(err, errScoringRule):
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, errNoDomainConf):
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	case errors.Is(err, errCaddyReload):
		writeJSON(w, http.StatusMultiStatus, map[string]any{"rule_id": ruleID, "warning": err.Error()})
		return
	case err != nil:
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !already {
		_ = logger.RecordUserAction(a.Config, username, "disabled WAF rule "+ruleID+" for domain "+domain+" from the WAF log", reqip.ClientIP(r))
	}
	writeJSON(w, http.StatusOK, map[string]any{"rule_id": ruleID, "domain": domain, "already_disabled": already})
}
