package domains

import (
	"context"
	"strings"

	appctx "gist.github.com/stefanpejcic/openpanel/internal/app"
	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
	"gist.github.com/stefanpejcic/openpanel/internal/web"
)

// docrootPicks lists domains with their document root, for the docroot and vhost pages
func docrootPicks(domains []appctx.Domain) []web.DomainPick {
	picks := make([]web.DomainPick, 0, len(domains))
	for _, d := range domains {
		picks = append(picks, web.DomainPick{Domain: d.DomainURL, Detail: d.Docroot})
	}
	return picks
}

// sslPicks lists domains with their certificate issuer and expiry, same wording as the domains list SSL column
func sslPicks(ctx context.Context, a *appctx.App, domains []appctx.Domain) []web.DomainPick {
	picks := make([]web.DomainPick, 0, len(domains))
	for _, d := range domains {
		p := web.DomainPick{Domain: d.DomainURL}
		if strings.HasSuffix(d.DomainURL, ".onion") {
			p.Badge, p.Tone = "None", "neutral"
			picks = append(picks, p)
			continue
		}
		https := isRewriteCondEnabled(ctx, a, d.DomainURL).HTTPS
		info := readSSLCertInfo(d.DomainURL, https)
		p.Detail = info.Issuer
		switch {
		case https != "Automatic" && https != "Custom":
			p.Badge, p.Tone = "Unknown", "error"
		case info.Problem == "expired":
			p.Badge, p.BadgeArg, p.Tone = "Expired", info.Expires, "error"
		case info.Problem == "expiring":
			p.Badge, p.BadgeArg, p.Tone = "Expires", info.Expires, "warning"
		case info.Problem == "missing":
			p.Badge, p.Tone = "Not issued yet", "error"
		case info.Problem == "unreadable":
			p.Badge, p.Tone = "Unreadable", "warning"
		case info.Expires != "":
			p.Badge, p.BadgeArg, p.Tone = "Valid until", info.Expires, "success"
		default:
			p.Badge, p.Tone = "Check", "neutral"
		}
		picks = append(picks, p)
	}
	return picks
}

// logPicks lists domains with their access log size; a missing or empty log has nothing to open
func logPicks(domains []appctx.Domain) []web.DomainPick {
	picks := make([]web.DomainPick, 0, len(domains))
	for _, d := range domains {
		p := web.DomainPick{Domain: d.DomainURL}
		info, err := tenantfs.Stat("/var/log/caddy/domlogs/" + d.DomainURL + "/access.log")
		switch {
		case err != nil:
			p.Disabled, p.Note = true, "No log yet"
		case info.Size() == 0:
			p.Disabled, p.Note = true, "Log is empty"
		default:
			p.Detail = web.FormatBytes(info.Size())
		}
		picks = append(picks, p)
	}
	return picks
}

// suspendPicks lists domains with their status; forSuspend drops the button on suspended ones, otherwise on active ones
func suspendPicks(ctx context.Context, a *appctx.App, domains []appctx.Domain, forSuspend bool) []web.DomainPick {
	picks := make([]web.DomainPick, 0, len(domains))
	for _, d := range domains {
		p := web.DomainPick{Domain: d.DomainURL, Badge: "Active", Tone: "success"}
		suspended := isRewriteCondEnabled(ctx, a, d.DomainURL).Suspended == "Suspended"
		if suspended {
			p.Badge, p.Tone = "Suspended", "error"
		}
		// the status badge already says why there's no button
		p.Disabled = forSuspend == suspended
		picks = append(picks, p)
	}
	return picks
}
