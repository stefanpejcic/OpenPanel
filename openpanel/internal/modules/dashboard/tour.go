package dashboard

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"sync"

	openpanel "gist.github.com/stefanpejcic/openpanel"
	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

var tourStringRE = regexp.MustCompile(`(?m)^\s*(?:title|description):\s*'((?:[^'\\]|\\.)*)'`)

// tour-steps.js stays the one place the tour's English text lives, we just pull its literals out here
var tourStrings = sync.OnceValue(func() []string {
	src, err := openpanel.Static.ReadFile("static/js/tour-steps.js")
	if err != nil {
		return nil
	}
	var out []string
	for _, m := range tourStringRE.FindAllStringSubmatch(string(src), -1) {
		out = append(out, strings.ReplaceAll(m[1], `\'`, `'`))
	}
	return out
})

// handleTourI18n returns {"English": "translated"} for every tour string plus the popover buttons
func handleTourI18n(a *appctx.App, w http.ResponseWriter, r *http.Request) {
	t := web.RequestTranslator(a, r)
	out := map[string]string{
		"Next":   t.Get("Next"),
		"Finish": t.Get("Finish"),
		"steps":  t.Get("steps"),
		"pages":  t.Get("pages"),
	}
	for _, s := range tourStrings() {
		if tr := t.Get(s); tr != s {
			out[s] = tr
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	_ = json.NewEncoder(w).Encode(out)
}
