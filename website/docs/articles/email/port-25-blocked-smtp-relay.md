---
sidebar_label: "Port 25 Blocked: Use an SMTP Relay"
description: "Which hosting providers block outgoing port 25 (AWS, Google Cloud, Azure, DigitalOcean, Hetzner, Linode, Vultr, Oracle), how to request unblocking, and how to send email through an SMTP relay in OpenPanel."
---

# Port 25 Blocked? How to Send Email Through an SMTP Relay

Mail servers deliver email to each other on **port 25**. To fight spam, many cloud providers **block outgoing port 25** by default. Your websites and the panel work normally, but email sent from the server never arrives - it stays in the queue with errors like `Connection timed out` to `gmail-smtp-in.l.google.com:25`.

You have two options: ask the provider to **unblock port 25**, or send outgoing email through an **SMTP relay**.

---

## Check If Port 25 Is Blocked

From the server:

```bash
timeout 5 bash -c '</dev/tcp/gmail-smtp-in.l.google.com/25' && echo "port 25 open" || echo "port 25 blocked"
```

or `nc -vz -w 5 gmail-smtp-in.l.google.com 25`.

---

## Which Providers Block Port 25

Policies change, so check your provider's current documentation. As a rule of thumb:

| Provider | Default | How to unblock |
|---|---|---|
| **AWS EC2** | Blocked / throttled | Request removal of the email sending limit (and set reverse DNS) in the AWS console |
| **Google Cloud** | Always blocked | Not possible - use a relay |
| **Microsoft Azure** | Blocked for most subscriptions | Only for some subscription types - usually use a relay |
| **Oracle Cloud** | Blocked | Support request - not available on the free tier; use a relay |
| **DigitalOcean** | Blocked for new accounts | Support ticket, usually after account history |
| **Vultr** | Blocked by default | Support ticket |
| **Linode / Akamai** | Restricted on new accounts | Support ticket |
| **Hetzner Cloud** | Blocked (25 and 465) for new accounts | Request after the first paid invoice |
| **OVHcloud**, **Contabo** | Usually open | - (IPs may be on blocklists) |
| **Home internet** | Almost always blocked | Use a relay |

Our install guides cover the provider-specific steps: [AWS](/docs/articles/install-update/install-on-aws/), [Hetzner](/docs/articles/install-update/install-on-hetzner/), [Linode](/docs/articles/install-update/install-on-linode/), [Oracle Cloud](/docs/articles/install-update/install-on-oracle-cloud/).

---

## Option 1: Configure an SMTP Relay for the Whole Server (Administrators)

With a relay (also called *smarthost*), the OpenPanel mail server hands all outgoing email to a relay service on an open port (usually **587**), and the relay delivers it. Incoming email is not affected.

Popular relay services: Amazon SES, Brevo (Sendinblue), Mailgun, SendGrid, Postmark, SMTP2GO, or your own mail server elsewhere.

1. Create an account with the relay service and **verify your sending domains** there (they give you SPF and DKIM records to add).
2. In OpenAdmin, go to **Emails → Settings → Relay Hosts**.
3. Fill in:
   - **DEFAULT_RELAY_HOST** and **RELAY_HOST**: the relay's SMTP server, e.g. `smtp-relay.brevo.com`
   - **RELAY_PORT**: `587`
   - **RELAY_USER** / **RELAY_PASSWORD**: the SMTP credentials from the relay service
4. Click **Save Relay**.

See [Email Settings - Relay Hosts](/docs/admin/emails/settings/#relay-hosts).

### Update SPF

Add the relay's `include:` to each sending domain's SPF record, for example:

```
v=spf1 ip4:203.0.113.10 +a +mx include:spf.brevo.com ~all
```

Otherwise receivers see mail "from" your domain arriving from the relay's servers and may mark it as spam. See [Add additional hosts to SPF](/docs/articles/domains/how-to-add-additional-hosts-in-spf/).

---

## Option 2: Send from Websites Through SMTP (No Server Mail Needed)

Even without the OpenPanel mail server, websites can send email (contact forms, WooCommerce orders, password resets) by connecting to an external SMTP service on port **587** or **465**:

- **WordPress**: install an SMTP plugin (*WP Mail SMTP*, *FluentSMTP*, *Post SMTP*) and enter the relay's credentials.
- **Laravel**: set `MAIL_MAILER=smtp`, `MAIL_HOST`, `MAIL_PORT=587`, `MAIL_USERNAME`, `MAIL_PASSWORD` in `.env`.
- **Other PHP apps**: use PHPMailer or the app's SMTP settings instead of PHP `mail()`.
- **Google Workspace / Microsoft 365**: if your mailboxes are there, use their SMTP relay.

---

## After Switching

- Send a test to [mail-tester.com](https://www.mail-tester.com/) - see [Test your mail setup](/docs/articles/email/test-email-with-mail-tester/).
- Check the queue in OpenAdmin (**Emails → Queue**) - stuck messages are retried automatically once the relay works.

---

## Related

- [Why do my emails go to spam?](/docs/articles/email/why-emails-go-to-spam/)
- [Set up reverse DNS (PTR)](/docs/articles/email/reverse-dns-ptr-record/)
- [Troubleshooting email errors](/docs/articles/email/troubleshooting-email-errors/)
