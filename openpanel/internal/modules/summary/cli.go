package summary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
)

// RunCLI handles `openpanel daily-summary <username> [--date YYYY-MM-DD] [--base-url URL] [--no-save]`, the day defaults to yesterday
func RunCLI(ctx context.Context, a *appctx.App, args []string, w io.Writer) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		fmt.Fprintln(w, "usage: openpanel daily-summary <username> [--date YYYY-MM-DD] [--base-url URL] [--no-save]")
		return 2
	}
	username := args[0]
	day := time.Now().AddDate(0, 0, -1)
	base, save := "", true
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--no-save":
			save = false
		case "--date", "--base-url":
			if i+1 >= len(args) {
				fmt.Fprintln(w, "error: "+args[i]+" needs a value")
				return 2
			}
			if args[i] == "--base-url" {
				base = strings.TrimRight(args[i+1], "/") + "/"
			} else {
				d, err := time.ParseInLocation("2006-01-02", args[i+1], time.Local)
				if err != nil {
					fmt.Fprintln(w, "error: --date must look like 2026-10-15")
					return 2
				}
				day = d
			}
			i++
		}
	}

	s, allowed, err := Build(ctx, a, username, day)
	if err != nil {
		fmt.Fprintln(w, "error:", err)
		return 1
	}
	out := map[string]any{"username": username, "allowed": allowed}
	if allowed {
		saved := false
		if save {
			if err := Save(s); err != nil {
				fmt.Fprintln(w, "error: saving the summary:", err)
				return 1
			}
			saved = true
		}
		out["date"], out["saved"] = s.Date, saved
		out["attention"], out["look"] = len(s.Attention), len(s.Look)
		out["subject"], out["body"] = s.Subject(), s.Body(base)
		// OpenAdmin's email template draws these as bars
		out["usage"] = s.Usage
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if enc.Encode(out) != nil {
		return 1
	}
	return 0
}
