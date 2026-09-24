---
sidebar_label: "Set Up Reverse DNS (PTR)"
description: "How to set up reverse DNS (a PTR record) for your OpenPanel mail server: what hostname to use, where to set it at Hetzner, OVH, DigitalOcean, Vultr, Linode, Contabo and AWS, and how to check it."
---

# How to Set Up Reverse DNS (PTR Record) for Your Mail Server

**Reverse DNS** turns an IP address back into a hostname - the opposite of a normal `A` record. Gmail, Outlook, Yahoo and most other mail providers check it for every incoming connection. If your server's IP has **no PTR record**, or it doesn't match the server's hostname, your emails are likely to be **rejected or sent to spam**.

A correct setup ("forward-confirmed reverse DNS") looks like this:

```
203.0.113.10          → PTR → server.example.com
server.example.com    → A   → 203.0.113.10
mail server says:  "HELO server.example.com"
```

All three must match.

---

## Step 1: Choose the Hostname

Use the hostname your server and mail server identify themselves with - normally the **domain you set for OpenAdmin/OpenPanel**, e.g. `server.example.com`:

- Set it in **OpenAdmin → Settings → General → Domain** (see [General settings](/docs/admin/settings/general/#domain)). OpenPanel also uses it as the mail server's hostname and for its SSL certificate.
- Make sure the hostname has an **`A` record** pointing to the server's IP.

:::warning
If OpenAdmin is accessed by IP only (no domain set), the mail server doesn't have a proper hostname. Set a domain before sending email from the server.
:::

---

## Step 2: Set the PTR Record at Your Provider

The PTR record is controlled by **whoever owns the IP address** - your hosting or cloud provider - not by your domain's DNS. Set it in the provider's control panel:

| Provider | Where |
|---|---|
| **Hetzner Cloud** | Server → **Networking** → ... next to the IPv4 → **Edit Reverse DNS** |
| **Hetzner Robot** (dedicated) | **IPs** → click the IP → **Reverse DNS entry** |
| **OVHcloud** | **Bare Metal Cloud → Network → IP** → ... → **Modify the reverse path** |
| **DigitalOcean** | Set automatically from the **Droplet name** - rename the Droplet to `server.example.com` |
| **Vultr** | Instance → **Settings → IPv4** → edit the **Reverse DNS** field |
| **Linode / Akamai** | Linode → **Network** tab → ... next to IPv4 → **Edit RDNS** |
| **Contabo** | Customer Control Panel → **Reverse DNS management** |
| **AWS EC2** | Elastic IP → **Actions → Update reverse DNS** (requires an Elastic IP; also request removal of the port 25 limit) |
| **Google Cloud** | VM → Edit → network interface → **Public DNS PTR Record** (hostname must be verified) |
| **Home / office connection** | Usually not possible - use an [SMTP relay](/docs/articles/email/port-25-blocked-smtp-relay/) |

Most providers check that the hostname already resolves to the IP, so create the `A` record first. If there is no self-service option, open a support ticket with the provider.

Provider-specific guides: [Hetzner](/docs/articles/install-update/install-on-hetzner/), [OVHcloud](/docs/articles/install-update/install-on-ovh/), [Linode](/docs/articles/install-update/install-on-linode/), [Contabo](/docs/articles/install-update/install-on-contabo/).

---

## Step 3: Check It

After a few minutes (sometimes up to a few hours), check both directions from any Linux or Mac terminal:

```bash
dig -x 203.0.113.10 +short
# server.example.com.

dig +short server.example.com
# 203.0.113.10
```

On Windows: `nslookup 203.0.113.10`.

You can also send a test email to [mail-tester.com](https://www.mail-tester.com/) - it reports reverse DNS problems. See [Test your mail setup with mail-tester](/docs/articles/email/test-email-with-mail-tester/).

---

## FAQ

**Do I need a PTR record for every domain on the server?**
No. One IP address has one PTR record. All domains on the server send from the same IP and share it - SPF and DKIM take care of each individual domain.

**What about IPv6?**
If the server sends mail over IPv6, the IPv6 address needs its own PTR record too - otherwise Gmail rejects the message. Set it the same way, or disable IPv6 for outgoing mail.

**Can I set the PTR record in OpenPanel's DNS zone editor?**
No - reverse DNS zones belong to the IP owner (your provider).

---

## Related

- [Why do my emails go to spam?](/docs/articles/email/why-emails-go-to-spam/)
- [Set up DMARC](/docs/articles/email/how-to-set-up-dmarc/)
- [Set up DKIM](/docs/articles/email/how-to-setup-dkim-for-mailserver/)
