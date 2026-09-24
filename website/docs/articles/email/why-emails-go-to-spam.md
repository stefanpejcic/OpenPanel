---
sidebar_label: "Why Emails Go to Spam"
description: "Why do emails from your server go to spam? A complete checklist for OpenPanel: SPF, DKIM, DMARC, reverse DNS (PTR), hostname, blocklists, port 25, content and sending reputation."
---

# Why Do My Emails Go to Spam? Deliverability Checklist

Emails landing in spam - or not arriving at all - almost always come down to a few technical checks that Gmail, Outlook and other providers run on every message. Work through this checklist from top to bottom; most problems are fixed by the first four steps.

**Quick test first:** send an email to [mail-tester.com](https://www.mail-tester.com/) - it checks most of the items below in one go. See [Test your mail setup](/docs/articles/email/test-email-with-mail-tester/).

---

## 1. SPF - Is Your Server Allowed to Send?

**SPF** is a TXT record listing which servers may send email for your domain. OpenPanel creates one automatically when you add a domain:

```
example.com.   TXT   "v=spf1 ip4:203.0.113.10 +a +mx ~all"
```

Check:

- There is **exactly one** SPF record per domain (two records = SPF fails).
- It contains your **server IP**, plus every other service that sends as your domain (Google Workspace, newsletter tools, SMTP relay) via `include:`.
- If your DNS is hosted elsewhere (Cloudflare, registrar), the record exists **there**.

See [Add additional hosts to SPF](/docs/articles/domains/how-to-add-additional-hosts-in-spf/).

---

## 2. DKIM - Is Your Email Signed?

**DKIM** adds a cryptographic signature to every message, verified with a public key in DNS (`mail._domainkey.example.com`). Without it, Gmail and Outlook trust your email much less.

See [Set up DKIM](/docs/articles/email/how-to-setup-dkim-for-mailserver/).

---

## 3. DMARC - Do You Have a Policy?

**DMARC** ties SPF and DKIM to the `From:` address and tells receivers what to do with failures. Gmail and Yahoo require it for bulk senders. OpenPanel adds a basic `p=none` record for every domain.

See [Set up DMARC](/docs/articles/email/how-to-set-up-dmarc/).

:::tip
Check SPF, DKIM and DMARC for all your domains at once in **OpenPanel → Email → Deliverability** - it compares the live DNS records with what OpenPanel expects. See [Email Deliverability](/docs/panel/emails/email_deliverability/).
:::

---

## 4. Reverse DNS (PTR) and Hostname

The server's IP must resolve back to its hostname, and the mail server must introduce itself with that same hostname:

```
203.0.113.10 → PTR → server.example.com → A → 203.0.113.10
```

- Set a **domain for the server** in **OpenAdmin → Settings → General** - it's also used as the mail server hostname.
- Set the **PTR record** at your hosting provider.

See [Set up reverse DNS (PTR)](/docs/articles/email/reverse-dns-ptr-record/).

---

## 5. Check Blocklists

If your server IP (or domain) is on a spam blocklist, many providers reject or filter your email.

- Check the IP at [MXToolbox Blacklist Check](https://mxtoolbox.com/blacklists.aspx) or [multirbl.valli.org](https://multirbl.valli.org/).
- Find out **why** it was listed first - usually a hacked website or mailbox sending spam. Check the mail queue in OpenAdmin (**Emails → Queue**) for unusual volumes, and scan sites for malware ([ImunifyAV](/docs/articles/security/setup-imunifyav/)).
- Then request delisting on each blocklist's website.
- New cloud IPs are sometimes listed from a previous owner - ask your provider for a clean IP, or use an [SMTP relay](/docs/articles/email/port-25-blocked-smtp-relay/).

For **Microsoft** (Outlook, Hotmail) delivery problems, use the [Microsoft Sender Support](https://sender.office.com/) form; for Gmail, register with [Google Postmaster Tools](https://postmaster.google.com/).

---

## 6. Is Port 25 Open?

If emails don't arrive at all (not even in spam) and sit in the queue, your provider may block outgoing **port 25**. See [Port 25 blocked: use an SMTP relay](/docs/articles/email/port-25-blocked-smtp-relay/).

---

## 7. Content and Sending Behaviour

With a correct technical setup, remaining spam-folder problems are usually about content and reputation:

- **Warm up** a new server or domain - start with small volumes and increase gradually.
- Don't send **newsletters or bulk mail** from your web server; use a dedicated email marketing service.
- Include a **plain-text version**, a real **From name**, and an **unsubscribe link** for newsletters.
- Avoid URL shorteners, all-caps subjects, large images with little text, and attachments like `.zip` or `.exe`.
- Make sure the **From address** is on your own domain - sending "from" `@gmail.com` through your server always fails DMARC.
- Websites should send through an **authenticated mailbox** (SMTP) on the site's own domain, not anonymous PHP `mail()`.

---

## 8. Using an External DNS Provider?

If your domain's nameservers point to Cloudflare, your registrar or another DNS host, **OpenPanel's DNS zone isn't used**. Copy the SPF, DKIM, DMARC and MX records from **OpenPanel → Domains → DNS** to your DNS provider. With Cloudflare, mail hostnames must be **DNS only** (grey cloud) - see [Use Cloudflare with OpenPanel](/docs/articles/domains/cloudflare-with-openpanel/).

---

## Checklist Summary

- [ ] One SPF record including the server IP and all senders
- [ ] DKIM key published and messages signed
- [ ] DMARC record published
- [ ] Server domain set; PTR record matches it
- [ ] IP and domain not on blocklists
- [ ] Outgoing port 25 open, or relay configured
- [ ] mail-tester score 9/10 or higher

---

## Related

- [Troubleshooting email errors](/docs/articles/email/troubleshooting-email-errors/)
- [Set up email clients](/docs/articles/email/how-to-setup-your-email-client/)
- [Enable email hosting](/docs/articles/user-experience/how-to-setup-email-in-openpanel/)
