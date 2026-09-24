---
sidebar_label: "Set Up DMARC"
description: "How to set up a DMARC record for your domain in OpenPanel: what the default record does, how to read reports, and how to move safely from p=none to quarantine and reject."
---

# How to Set Up DMARC for Your Domain

**DMARC** tells receiving mail servers what to do with email that claims to come from your domain but fails **SPF** and **DKIM** checks - and where to send reports about it. Since 2024, Gmail and Yahoo **require** a DMARC record for anyone sending bulk email, and it helps every domain get into the inbox.

---

## The Default Record in OpenPanel

When you add a domain, OpenPanel creates a basic DMARC record in its DNS zone:

```
_dmarc.example.com.   TXT   "v=DMARC1; p=none;"
```

`p=none` means "monitor only - don't block anything". It satisfies the "has a DMARC record" requirement, but doesn't protect your domain from spoofing yet, and you don't receive any reports.

Check your domains in **OpenPanel → Email → Deliverability** - it shows whether SPF, DKIM and DMARC are published correctly. See [Email Deliverability](/docs/panel/emails/email_deliverability/).

---

## Before You Tighten DMARC: SPF and DKIM

DMARC passes only when **SPF or DKIM passes and aligns** with the domain in the `From:` address. So first make sure both are set up:

- **SPF** - created automatically, e.g. `v=spf1 ip4:203.0.113.10 +a +mx ~all`. If you also send through other services (Google Workspace, Mailchimp, a transactional email provider), add their `include:` to the same record - see [Add additional hosts to SPF](/docs/articles/domains/how-to-add-additional-hosts-in-spf/).
- **DKIM** - see [Set up DKIM](/docs/articles/email/how-to-setup-dkim-for-mailserver/).

---

## Step 1: Start Monitoring with Reports

Edit the `_dmarc` TXT record in **OpenPanel → Domains → DNS** (or at your external DNS provider) and add a report address:

```
v=DMARC1; p=none; rua=mailto:dmarc@example.com; fo=1
```

| Tag | Meaning |
|---|---|
| `v=DMARC1` | Version - always first |
| `p=none` | Policy: `none` (monitor), `quarantine` (spam folder) or `reject` (bounce) |
| `rua=mailto:...` | Where to send daily aggregate reports |
| `fo=1` | Report when either SPF or DKIM fails |
| `pct=` | Apply the policy to only a percentage of messages (default 100) |
| `sp=` | Policy for subdomains (default: same as `p`) |

Create the `dmarc@example.com` mailbox (or use a free DMARC report service, which gives you a dashboard instead of XML files).

---

## Step 2: Read the Reports (1-2 Weeks)

Reports list every server that sent mail using your domain, and whether SPF and DKIM passed. Look for:

- **Your own server** - should pass SPF and DKIM.
- **Services you use** (newsletter tool, CRM, invoicing, Google Workspace) - if they fail, add them to SPF and set up DKIM in that service.
- **Unknown senders** - spammers spoofing your domain. These are what DMARC will block.

---

## Step 3: Enforce the Policy

When all legitimate mail passes, tighten the policy step by step:

```
v=DMARC1; p=quarantine; pct=25; rua=mailto:dmarc@example.com
```

Increase `pct` to 100, then move to:

```
v=DMARC1; p=reject; rua=mailto:dmarc@example.com
```

With `p=reject`, spoofed email using your domain is refused by recipient servers.

---

## Check Your Record

```bash
dig +short TXT _dmarc.example.com
```

Or send a test email to [mail-tester.com](https://www.mail-tester.com/) - see [Test your mail setup](/docs/articles/email/test-email-with-mail-tester/).

---

## Common Mistakes

| Mistake | Result |
|---|---|
| Two `_dmarc` TXT records | DMARC is ignored - keep exactly one. |
| Record placed at `example.com` instead of `_dmarc.example.com` | No DMARC found. |
| Jumping straight to `p=reject` | Legitimate mail from third-party services gets rejected. |
| Domain uses Cloudflare DNS, record edited in OpenPanel | Changes have no effect - edit DNS where the domain's nameservers point. |

---

## Related

- [Why do my emails go to spam?](/docs/articles/email/why-emails-go-to-spam/)
- [Set up reverse DNS (PTR)](/docs/articles/email/reverse-dns-ptr-record/)
