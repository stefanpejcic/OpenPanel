---
sidebar_label: "Test Your Email with mail-tester"
description: "How to test your OpenPanel mail setup with mail-tester.com: send a test email, read the score, and fix the most common problems - SPF, DKIM, DMARC, reverse DNS and blocklists."
---

# How to Test Your Email Setup with mail-tester

[mail-tester.com](https://www.mail-tester.com/) gives your email a spam score from **0 to 10** and explains exactly what's wrong. It's the quickest way to check SPF, DKIM, DMARC, reverse DNS and blocklists after setting up email in OpenPanel.

---

## Step 1: Get a Test Address

Open [mail-tester.com](https://www.mail-tester.com/). It shows a unique address like:

```
test-abc123xyz@srv1.mail-tester.com
```

Keep the page open.

![mail-tester.com home page showing the unique test address to send your email to and the Then check your score button](/img/openpanel-screenshots/email/mail-tester-address.png#screenshot)

---

## Step 2: Send a Real Email

Send an email **to that address** from the mailbox you want to test - for example from Roundcube webmail, or your email client set up with your OpenPanel mailbox.

- Use a **real subject and a few sentences** of text - an empty "test" message loses points for content, not for your setup.
- To test **website email** (WordPress, contact forms), send the test from the website instead - e.g. with a plugin's "send test email" option.

---

## Step 3: Check Your Score

Go back to mail-tester and click **Then check your score**. Aim for **9/10 or 10/10**.

![mail-tester results with a score of 7.7/10 and the sections for the message, SpamAssassin, authentication, content and blocklists](/img/openpanel-screenshots/email/mail-tester-score.png#screenshot)

Expand each section to see details:

| Section | What it checks | If it fails |
|---|---|---|
| **You're authenticated (SPF, DKIM, DMARC)** | Your DNS authentication records | See below |
| **Reverse DNS** | PTR record of the sending IP matches the hostname | [Set up reverse DNS](/docs/articles/email/reverse-dns-ptr-record/) |
| **You're not blocklisted** | Your server IP and domain on public blocklists | Request delisting; see [Why emails go to spam](/docs/articles/email/why-emails-go-to-spam/#5-check-blocklists) |
| **SpamAssassin** | Content and formatting rules | Improve the message content |
| **Broken links / formatting** | HTML and links | Fix the message template |

---

## Fixing the Most Common Problems

Expand **You're not fully authenticated** to see which check failed. In this example SPF and DMARC pass, but the message isn't signed with DKIM and the reverse DNS doesn't match the sending domain:

![mail-tester authentication results: SPF passed, message not signed with DKIM, DMARC passed, reverse DNS does not match the sending domain, and the domain and hostname checks passed](/img/openpanel-screenshots/email/mail-tester-authentication.png#screenshot)

### SPF: "Your server is not allowed to send"

The sending IP isn't in the domain's SPF record. Check the record in **OpenPanel → Domains → DNS** and make sure it includes your server's IP (and any relay you use). See [Add additional hosts to SPF](/docs/articles/domains/how-to-add-additional-hosts-in-spf/).

### DKIM: "Your message is not signed"

Generate the DKIM key and publish the `mail._domainkey` record - see [Set up DKIM](/docs/articles/email/how-to-setup-dkim-for-mailserver/). If your DNS is at Cloudflare or your registrar, add the record there.

### DMARC: "Your message failed the DMARC test" or no record

Publish a DMARC record - see [Set up DMARC](/docs/articles/email/how-to-set-up-dmarc/).

### Reverse DNS doesn't match

Set the PTR record of your server IP to the server's hostname at your provider.

### "Your hostname is ... local.host" / HELO mismatch

The mail server doesn't have a proper hostname. Set a domain for the server in **OpenAdmin → Settings → General** - see [Change the server hostname](/docs/articles/server/change-server-hostname-or-ip/).

:::tip
In **OpenPanel → Email → Deliverability**, you can check SPF, DKIM and DMARC for all your domains at once, and see the exact records OpenPanel expects. See [Email Deliverability](/docs/panel/emails/email_deliverability/).
:::

---

## Re-Test After Fixing

DNS changes can take a while to propagate. Wait until the new records are visible (`dig +short TXT example.com`), then send a **new** test email to a **new** mail-tester address. The free plan allows a few tests per day.

---

## Related

- [Why do my emails go to spam?](/docs/articles/email/why-emails-go-to-spam/)
- [Port 25 blocked: use an SMTP relay](/docs/articles/email/port-25-blocked-smtp-relay/)
- [Troubleshooting email errors](/docs/articles/email/troubleshooting-email-errors/)
