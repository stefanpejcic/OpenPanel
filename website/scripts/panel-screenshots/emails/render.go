// renders the user notification emails through OpenAdmin's user template, the one /send_email uses for type=user
// usage: go run render.go <openadmin webtemplates dir> <out dir> <openpanel static/flags dir>
package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type email struct {
	name, to, subject, body  string
	usage                    []usage
	upgradePlan, upgradeText string
	tips                     string
	details                  []detail
}

// usage mirrors OpenAdmin's emailUsageItem, colors come from usageStyle like parseEmailUsageItems does
type usage struct {
	Title, Used, Total, Label, Color string
	Percent, Limit, Width            int
}

func usageStyle(items []usage) []usage {
	for i := range items {
		it := &items[i]
		switch {
		case it.Percent >= 100:
			it.Label, it.Color = "Full", "#dc2626"
		case it.Limit > 0 && it.Percent >= it.Limit:
			it.Label, it.Color = "Over "+strconv.Itoa(it.Limit)+"%", "#d97706"
		default:
			it.Label, it.Color = "OK", "#059669"
		}
		it.Width = min(max(it.Percent, 0), 100)
	}
	return items
}

// detail mirrors OpenAdmin's emailDetail, FlagURL is a local file here so it's a template.URL to get past html/template's scheme check
type detail struct {
	Label, Value, URL string
	FlagURL           template.URL
}

// sec is a security email like OpenPanel's securityEmail builds, with fake request details
func sec(name, subject, text, tips string) email {
	return email{name: name, subject: subject, body: text, tips: tips, details: []detail{
		{Label: "Time", Value: "2026-09-25 18:30:12 UTC"},
		{Label: "IP address", Value: "203.0.113.10"},
		{Label: "Country", Value: "DE", FlagURL: template.URL("file://" + filepath.Join(os.Args[3], "de.png"))},
		{Label: "Browser", Value: "Chrome on Windows"},
	}}
}

// bodies match what OpenPanel and opencli send, with john / example.com as fake data
func emails() []email {
	return []email{
		sec("login", "New login to OpenPanel", "New password login from IP 203.0.113.10.",
			"If this wasn't you, change your password right away and turn on two-factor authentication."),
		sec("password", "Password changed for account john", "The password for account john was changed from the OpenPanel interface. All other sessions stay logged in until they expire.",
			"If you didn't change it, reset your password right away or contact your hosting provider."),
		sec("2fa-enabled", "Two-factor authentication enabled for account john", "Two-factor authentication was turned on. A code from your authenticator app is now needed on every login.",
			"If you didn't turn it on, contact your hosting provider, someone else may now control your 2FA codes."),
		sec("2fa-disabled", "Two-factor authentication disabled for account john", "Two-factor authentication was turned off. Only the password is needed to log in now.",
			"If you didn't turn it off, change your password and turn 2FA back on right away."),
		sec("passkey", "New passkey added to account john", "A passkey named \"MacBook Touch ID\" was added. It can be used to log in without a password.",
			"If you didn't add it, delete it from Account > Passkeys and change your password right away."),
		sec("passkey-removed", "Passkey removed from account john", "The passkey named \"MacBook Touch ID\" was removed and can't be used to log in anymore.",
			"If you didn't remove it, log in, check your passkeys on Account > Passkeys and change your password right away."),
		sec("api-token", "New API token created for account john", "An AI Assistant (MCP) token named \"Claude Desktop\" was created with read-only access that expires in 30 days. Anyone who has it can manage account john through the AI Assistant.",
			"If you didn't create it, revoke it right away on Account > AI Assistant (MCP) and change your password."),
		sec("contact-email", "Email address changed for account john", "The contact email address for account john was changed from john@example.com to new@example.com. This is the last notification sent to this address, new ones go to new@example.com.",
			"If you didn't change it, log in and set your email address back, then change your password."),
		sec("alert-disabled", "Notification preferences changed for account john", "These email alerts were turned off for account john:\n- New login\n- Password changed\n\nYou won't get emails for them anymore.",
			"If you didn't turn them off, turn them back on from Account > Notifications and change your password right away."),
		{name: "disk", subject: "Account john is almost out of disk space",
			body:        "Account john is close to its hosting plan limit, checked 2026-09-25 18:30:00 UTC.\n\nOnce the limit is reached, websites and email can stop working, and new files, uploads and emails will fail.",
			tips:        "To free up space, find the biggest folders on the Disk Usage and Inodes Explorer pages in OpenPanel, and delete old backups, logs and files you no longer need. If you need more space, ask your hosting provider for a bigger plan.",
			usage:       []usage{{Title: "Disk space", Used: "8.70 GB", Total: "10.00 GB", Percent: 87, Limit: 85}, {Title: "Inodes (files and folders)", Used: "42117", Total: "100000 files", Percent: 42, Limit: 95}},
			upgradePlan: "Business", upgradeText: "The Business plan gives you 50 GB of disk space instead of 10 GB."},
		{name: "mailbox", subject: "Email accounts on john are almost full",
			body:        "These email accounts on john are using 90% or more of their quota, checked 2026-09-25 18:30:00 UTC.\n\nOnce a mailbox is full, new emails sent to it are not delivered and bounce back to the sender.",
			tips:        "To free up space, delete old emails and empty the Trash and Spam folders in webmail, or raise the mailbox quota on the Email Accounts page in OpenPanel.",
			usage:       []usage{{Title: "info@example.com", Used: "920M", Total: "1G", Percent: 92, Limit: 90}, {Title: "sales@example.com", Used: "2G", Total: "2G", Percent: 100, Limit: 90}},
			upgradePlan: "Business", upgradeText: "The Business plan lets you set mailboxes up to 5 GB instead of 2 GB. After upgrading, raise the quota on the Email Accounts page."},
		{name: "ssl", subject: "SSL certificate problem on account john",
			body: "These SSL certificates on account john need attention, without a valid certificate visitors see a security warning instead of the website:\n\n- blog.example.com: the AutoSSL certificate expires on 2026-10-02 06:00 UTC (6 day(s) left). It should have been renewed automatically, so renewal is failing. Last error: HTTP 403 urn:ietf:params:acme:error:unauthorized - Invalid response from http://blog.example.com/.well-known/acme-challenge/x: 404\n- shop.example.com: the custom certificate expires on 2026-09-26 12:00 UTC (less than a day left).\n\nChecked: 2026-09-25 06:15:02 UTC",
			tips: "For AutoSSL, make sure the domain's A and AAAA records point to this server, no CAA record blocks Let's Encrypt, and no proxy or firewall blocks /.well-known/acme-challenge/. For a custom certificate, upload a renewed one on the Domains > SSL Certificates page in OpenPanel, or switch the domain to AutoSSL to get free certificates that renew on their own."},
		{name: "malware", subject: "Malware found on account john",
			body: "The scheduled malware scan found 2 infected file(s) on account john and moved them to quarantine, so they can't run or be downloaded anymore.\n\nQuarantined files:\n- /var/www/html/blog.example.com/wp-content/uploads/2026/09/cache.php (Php.Webshell.Generic)\n- /var/www/html/blog.example.com/wp-includes/js/jquery/ui/core.min.js (Js.Trojan.Redirect)\n\nScanned: 2026-09-27 00:12:40 UTC",
			tips: "Review the files on the Malware Scanner > Quarantine page in OpenPanel: delete them, or restore a file if it's a false positive. Malware usually gets in through outdated plugins, themes or stolen passwords, so update your websites and change your FTP and admin passwords."},
		{name: "ratelimit", subject: "Account john reached its hourly email limit",
			body:        "Account john reached its limit of 100 emails per hour, so 37 email(s) were not sent between 18:30 and 18:42.\n\nRejected emails by sender:\n- info@example.com: 35 rejected\n- shop@example.com: 2 rejected\n\nSending works again on its own once fewer than 100 emails were sent in the last hour. The rejected emails are not sent later, send them again if they're still needed.",
			tips:        "If you didn't send this many emails, a mailbox password may be stolen or a contact form on your website may be abused by spammers. Change the password of the mailbox above and check your website's forms.",
			usage:       []usage{{Title: "Emails sent in the last hour", Used: "100", Total: "100 emails", Percent: 100, Limit: 100}},
			upgradePlan: "Business", upgradeText: "The Business plan lets you send up to 500 emails per hour instead of 100."},
		{name: "malware-imunify", subject: "Malware found on account john",
			body: "An ImunifyAV malware scan found 2 infected file(s) on account john:\n\n- /var/www/html/blog.example.com/wp-content/uploads/2026/10/cache.php\n- /var/www/html/blog.example.com/wp-includes/js/jquery/ui/core.min.js\n\nScanned: 2026-10-15 03:40:12 UTC",
			tips: "Remove or clean these files, or ask your hosting provider for help if you're not sure what they are. Malware usually gets in through outdated plugins, themes or stolen passwords, so update your websites and change your FTP and admin passwords."},
		{name: "wp-vulnerability", subject: "Vulnerable WordPress plugins or themes on account john",
			body: "The scheduled security scan found known vulnerabilities on WordPress websites of account john:\n\n- blog.example.com: plugin contact-form-7: Contact Form 7 [contact-form-7] < 5.3.2\n- blog.example.com: plugin contact-form-7: Contact Form 7 [contact-form-7] < 5.9.2\n- shop.example.com/blog: theme astra: Astra < 4.1.6\n\nScanned: 2026-10-15 05:00:12 UTC",
			tips: "Update WordPress, the plugins and the themes listed above from the WordPress dashboard or OpenPanel > WordPress, most of these are fixed in a newer version. If a plugin or theme has no fix, deactivate and remove it until the author releases one."},
		{name: "visitors-blocked", subject: "Visitors blocked by rate limit on account john",
			body: "In the last 6 hours some visitors to your websites were blocked for sending too many requests per minute, and got a 429 Too Many Requests error:\n\n- blog.example.com: 340 of 9000 requests blocked, most from 203.0.113.9 (300)\n- shop.example.com: 12 of 500 requests blocked, most from 198.51.100.24 (12)\n\nChecked: 2026-10-15 12:00:02 UTC",
			tips: "One IP with most of the blocks is usually a bot or scraper, and blocking it is what the limit is for. If real visitors are being blocked, for example during a sale or after a newsletter, ask your hosting provider to raise the rate limit for your domain."},
		{name: "error-report", subject: "Daily error report for account john",
			body: "In the last 24 hours, 3 error(s) were found in the logs of account john.\n\nBy website:\n- blog.example.com: 2\n- shop.example.com: 1\n\nLatest errors:\n- 2026-10-15 19:49 UTC [critical] php-fpm-8.3 (blog.example.com): Fatal error: Uncaught Error: Call to undefined function wc_get_product() in /var/www/html/blog.example.com/wp-content/themes/shop/functions.php:42\n- 2026-10-15 14:02 UTC [error] apache (blog.example.com): AH01071: Got error 'Primary script unknown'\n- 2026-10-15 09:13 UTC [error] php-fpm-8.3 (shop.example.com): PHP Parse error: syntax error, unexpected token \"}\" in /var/www/html/shop.example.com/index.php on line 18",
			tips: "Open Error Logs in OpenPanel to see the full messages and stack traces, filter by website and copy them for your developer or hosting support. To stop these emails, turn off 'Daily error report' in Account > Notifications."},
		{name: "daily-summary", subject: "⚠ 1 thing needs your attention – 15 Oct summary for john",
			body: "Hi john,\n\nHere's how your account did on Thursday, 15 October. Your 2 websites handled 5,452 requests from 1,448 visitors, 12% more than the day before. 1 thing needs your attention.\n\nNEEDS YOUR ATTENTION\n\n1. blog.example.com has 3 known security holes\n   They're in Astra and Contact Form 7. An update usually fixes them.\n   → Update WordPress and its plugins: https://panel.example.com:2083/website?domain=blog.example.com\n\nWORTH A LOOK\n\n• Your account hit its CPU limit\n   It used 2.0 cores of 2.0 cores at 14:00, and was at the limit in 2 of the 24 hourly checks. Pages load slower when that happens.\n   → See resource usage: https://panel.example.com:2083/server/usage\n\n• Your disk is 86% full\n   You're using 8.6 GB of 10.0 GB. The Trash alone holds 1.4 GB.\n   → Empty the Trash: https://panel.example.com:2083/files.trash\n\nWHAT WENT WELL\n✓ The last malware scan, on 12 October, found nothing.\n✓ The web firewall stopped 37 attacks on your websites.\n✓ All your security certificates are valid.\n✓ Your account was backed up on 15 October.\n✓ Your busiest page was /shop on blog.example.com, with 812 views.\n\nYOUR WEBSITES\n- blog.example.com: 4,812 requests from 1,236 visitors, 182 MB of bandwidth (12% more than the day before, 14 missing pages)\n- shop.example.com: 640 requests from 212 visitors, 21 MB of bandwidth (about the same as the day before)\n\nYOUR ACCOUNT\nYou logged in 2 times:\n  08:57 from 203.0.113.10 (RS)\n  17:40 from 203.0.113.10 (RS)\n3 changes were made:\n  09:12 added domain shop.example.com\n  09:20 changed PHP version for blog.example.com to 8.3\n  10:05 created cron job\n\nSee the full summary: https://panel.example.com:2083/summary/2026-10-15",
			usage: []usage{{Title: "Disk space", Used: "8.6 GB", Total: "10.0 GB", Percent: 86, Limit: 85}, {Title: "Files and folders (inodes)", Used: "42,117", Total: "100,000 files", Percent: 42, Limit: 85},
				{Title: "CPU at the busiest hour", Used: "2.0 cores", Total: "2.0 cores", Percent: 100, Limit: 90}, {Title: "Memory at the busiest hour", Used: "1.2 GB", Total: "2.0 GB", Percent: 60, Limit: 90}},
			tips: "To stop these emails, turn off 'Daily summary' in Account > Notifications. The summaries stay in Statistics > Daily Summary either way."},
		{name: "service", subject: "Services restarted on account john",
			body: "Some services on account john had stopped, the server's automatic checks tried to start them again:\n\n- mysql: ran out of memory, restarted\n- php-fpm-8.3: restarted\n\nChecked: 2026-09-25 19:00:02 UTC",
			tips: "If a service keeps stopping, check its logs on the Containers page in OpenPanel. A service that ran out of memory needs a higher memory limit on the Containers page, or less memory use, for example fewer PHP workers or a smaller database cache."},
	}
}

const frame = `<!doctype html><html><head><meta charset="utf-8"><style>body{margin:0;font-family:-apple-system,Segoe UI,Roboto,Helvetica,Arial,sans-serif;background:#fff}
.hdr{padding:18px 24px;border-bottom:1px solid #e5e7eb;font-size:14px;color:#374151}.hdr h1{font-size:18px;margin:0 0 10px;color:#111827}.hdr span{color:#6b7280;display:inline-block;width:48px}</style></head><body>
<div id="mail"><div class="hdr"><h1>[panel.example.com] %s</h1><div><span>From</span>OpenPanel &lt;noreply@example.com&gt;</div><div><span>To</span>%s</div></div><div style="padding:16px">%s</div></div></body></html>`

func main() {
	tpl := template.Must(template.ParseFiles(filepath.Join(os.Args[1], "email_user_notifications.html")))
	for _, e := range emails() {
		subject := e.subject
		data := map[string]any{"Title": subject, "Message": e.body, "Hostname": "server.example.com", "NotificationsURL": "https://panel.example.com:2083/account/notifications", "Usage": usageStyle(e.usage),
			"UpgradePlan": e.upgradePlan, "UpgradeText": e.upgradeText, "UpgradeURL": "https://panel.example.com:2083/dashboard/upgrade",
			"Tips": e.tips, "Details": e.details}
		var b strings.Builder
		if err := tpl.ExecuteTemplate(&b, "email_user_notifications.html", data); err != nil {
			panic(err)
		}
		to := e.to
		if to == "" {
			to = "john@example.com"
		}
		page := fmt.Sprintf(frame, template.HTMLEscapeString(subject), to, b.String())
		if err := os.WriteFile(filepath.Join(os.Args[2], e.name+".html"), []byte(page), 0o644); err != nil {
			panic(err)
		}
		fmt.Println("rendered", e.name)
	}
}
