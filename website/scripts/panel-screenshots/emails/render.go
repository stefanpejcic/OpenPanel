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
		sec("contact-email", "Email address changed for account john", "The contact email address for account john was changed from john@example.com to new@example.com. This is the last notification sent to this address, new ones go to new@example.com.",
			"If you didn't change it, log in and set your email address back, then change your password."),
		sec("alert-disabled", "Notification preferences changed for account john", "These email alerts were turned off for account john:\n- New login\n- Password changed\n\nYou won't get emails for them anymore.",
			"If you didn't turn them off, turn them back on from Account > Notifications and change your password right away."),
		{name: "disk", subject: "Account john is almost out of disk space",
			body:        "Account john is close to its hosting plan limit, checked 2026-09-25 18:30:00 UTC.\n\nOnce the limit is reached, websites and email can stop working, and new files, uploads and emails will fail.",
			tips:        "To free up space, find the biggest folders on the Disk Usage and Inodes Explorer pages in OpenPanel, and delete old backups, logs and files you no longer need. If you need more space, ask your hosting provider for a bigger plan.",
			usage:       []usage{{Title: "Disk space", Used: "8.70 GB", Total: "10.00 GB", Percent: 87, Limit: 85}, {Title: "Inodes (files and folders)", Used: "42117 files", Total: "100000 files", Percent: 42, Limit: 95}},
			upgradePlan: "Business", upgradeText: "The Business plan gives you 50 GB of disk space instead of 10 GB."},
		{name: "mailbox", subject: "Email accounts on john are almost full",
			body:        "These email accounts on john are using 90% or more of their quota, checked 2026-09-25 18:30:00 UTC.\n\nOnce a mailbox is full, new emails sent to it are not delivered and bounce back to the sender.",
			tips:        "To free up space, delete old emails and empty the Trash and Spam folders in webmail, or raise the mailbox quota on the Email Accounts page in OpenPanel.",
			usage:       []usage{{Title: "info@example.com", Used: "920M", Total: "1G", Percent: 92, Limit: 90}, {Title: "sales@example.com", Used: "2G", Total: "2G", Percent: 100, Limit: 90}},
			upgradePlan: "Business", upgradeText: "The Business plan lets you set mailboxes up to 5 GB instead of 2 GB. After upgrading, raise the quota on the Email Accounts page."},
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
