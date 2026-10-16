package summary

import (
	"fmt"
	"strings"
)

// Subject says before it's opened whether there's anything to do
func (s Summary) Subject() string {
	day := s.From.Format("2 Jan")
	switch {
	case len(s.Attention) > 0:
		return fmt.Sprintf("⚠ %s your attention – %s summary for %s", plural(len(s.Attention), "thing needs", "things need"), day, s.Username)
	case len(s.Look) > 0:
		return fmt.Sprintf("%s worth a look – %s summary for %s", plural(len(s.Look), "thing", "things"), day, s.Username)
	}
	return fmt.Sprintf("✓ All good – %s summary for %s", day, s.Username)
}

func writeItem(sb *strings.Builder, prefix string, it Item, base string) {
	sb.WriteString(prefix + it.Title + "\n")
	if it.Detail != "" {
		sb.WriteString("   " + it.Detail + "\n")
	}
	switch {
	case it.Action != "" && it.Link != "":
		sb.WriteString("   → " + it.Action + ": " + strings.TrimRight(base, "/") + it.Link + "\n")
	case it.Action != "":
		sb.WriteString("   → " + it.Action + "\n")
	}
	sb.WriteString("\n")
}

// websiteLine is one site's day in a sentence, with the extras only when they're not zero
func websiteLine(w Website) string {
	line := fmt.Sprintf("%s: %s requests from %s, %s of bandwidth", w.Name, num(w.Requests), plural(w.UniqueIPs, "visitor", "visitors"), humanBytes(w.Bytes))
	var extra []string
	if c := change(w.Requests, w.PrevRequests); c != "" {
		extra = append(extra, c)
	}
	if w.NotFound > 0 {
		extra = append(extra, plural(w.NotFound, "missing page", "missing pages"))
	}
	if w.ServerErrors > 0 {
		extra = append(extra, plural(w.ServerErrors, "error page", "error pages"))
	}
	if len(extra) > 0 {
		line += " (" + strings.Join(extra, ", ") + ")"
	}
	return line
}

// Body is the plain text email: the verdict, what to fix with links, what went well, then the numbers
func (s Summary) Body(base string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Hi %s,\n\nHere's how your account did on %s. %s\n\n", s.Username, s.From.Format("Monday, 2 January"), s.Headline)

	if len(s.Attention) > 0 {
		sb.WriteString("NEEDS YOUR ATTENTION\n\n")
		for i, it := range s.Attention {
			writeItem(&sb, fmt.Sprintf("%d. ", i+1), it, base)
		}
	}
	if len(s.Look) > 0 {
		sb.WriteString("WORTH A LOOK\n\n")
		for _, it := range s.Look {
			writeItem(&sb, "• ", it, base)
		}
	}
	if len(s.Good) > 0 {
		sb.WriteString("WHAT WENT WELL\n")
		for _, g := range s.Good {
			sb.WriteString("✓ " + g + "\n")
		}
		sb.WriteString("\n")
	}

	if len(s.Websites) > 0 {
		sb.WriteString("YOUR WEBSITES\n")
		for i, w := range s.Websites {
			if i == 10 {
				fmt.Fprintf(&sb, "- and %s, see the full summary\n", plural(len(s.Websites)-10, "more website", "more websites"))
				break
			}
			sb.WriteString("- " + websiteLine(w) + "\n")
		}
		sb.WriteString("\n")
	}

	var account []string
	account = append(account, s.Facts...)
	// disk, CPU and memory go to the email as bars, not text
	if n := len(s.Logins); n > 0 {
		account = append(account, fmt.Sprintf("You logged in %s:", plural(n, "time", "times")))
		for i, l := range s.Logins {
			if i == 5 {
				account = append(account, fmt.Sprintf("  and %s more", num(n-5)))
				break
			}
			where := l.IP
			if l.Country != "" {
				where += " (" + l.Country + ")"
			}
			account = append(account, "  "+l.Time.Format("15:04")+" from "+where)
		}
	}
	if len(account) > 0 || len(s.Changes) > 0 {
		sb.WriteString("YOUR ACCOUNT\n")
		for _, a := range account {
			sb.WriteString(a + "\n")
		}
		if n := len(s.Changes); n > 0 {
			fmt.Fprintf(&sb, "%s made:\n", plural(n, "change was", "changes were"))
			for i, c := range s.Changes {
				if i == 5 {
					fmt.Fprintf(&sb, "  and %s more\n", num(n-5))
					break
				}
				sb.WriteString("  " + c + "\n")
			}
		}
		sb.WriteString("\n")
	}
	fmt.Fprintf(&sb, "See the full summary: %ssummary/%s\n", base, s.Date)
	return sb.String()
}
