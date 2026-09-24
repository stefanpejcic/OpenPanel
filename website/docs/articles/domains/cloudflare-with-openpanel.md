---
sidebar_label: "Use Cloudflare with OpenPanel"
description: "How to use Cloudflare with OpenPanel: DNS records, Full (strict) SSL mode, Cloudflare Origin Certificates, fixing ERR_TOO_MANY_REDIRECTS and 526 errors, email records and showing real visitor IPs."
---

# How to Use Cloudflare with OpenPanel (SSL, Origin Certificates and Redirect Loops)

[Cloudflare](https://www.cloudflare.com/) sits in front of your server as a CDN, DNS provider and firewall. It works well with OpenPanel - as long as the **SSL mode** is set correctly. Most problems people run into (redirect loops, 525/526 errors, SSL not issuing) come from that one setting.

---

## Step 1: Add the DNS Records in Cloudflare

When your domain uses Cloudflare's nameservers, DNS is managed in the Cloudflare dashboard - not in OpenPanel's DNS zone editor.

Add at least:

| Type | Name | Content | Proxy |
|---|---|---|---|
| `A` | `example.com` (`@`) | your server's IPv4 | 🟠 Proxied |
| `CNAME` | `www` | `example.com` | 🟠 Proxied |
| `A` | `mail` | your server's IPv4 | ⚪ **DNS only** |
| `MX` | `example.com` | `mail.example.com` | - |
| `TXT` | `example.com` | your SPF record | - |

Copy the **SPF, DKIM and DMARC** TXT records from the domain's zone in **OpenPanel → Domains → DNS**, so email keeps working.

:::warning Mail must not be proxied
Cloudflare's proxy only handles web traffic. Any hostname used for **email** (`mail.example.com`, the MX target), **FTP** or the **control panel ports** (`2083`, `2087`) must be **DNS only** (grey cloud).
:::

Next, set the SSL mode (Step 2) and install a certificate on your server (Step 3) - either the automatic Let's Encrypt certificate or a [Cloudflare Origin Certificate installed as custom SSL](#option-b-cloudflare-origin-certificate).

---

## Step 2: Set SSL/TLS to Full (strict)

In Cloudflare, go to **SSL/TLS → Overview** and choose **Full (strict)**.

| Mode | What happens with OpenPanel |
|---|---|
| **Off** / **Flexible** | ❌ Cloudflare talks to your server over HTTP; the server redirects to HTTPS → **endless redirect loop** (`ERR_TOO_MANY_REDIRECTS`). |
| **Full** | ⚠️ Works, but Cloudflare doesn't verify your server's certificate. |
| **Full (strict)** | ✅ Encrypted all the way and verified. **Use this.** |

Full (strict) needs a valid certificate on your server - either the automatic Let's Encrypt certificate or a Cloudflare Origin Certificate.

---

## Step 3: Choose the Certificate on Your Server

### Option A: Keep the free Let's Encrypt certificate (easiest)

OpenPanel issues it automatically. For the **first** certificate, it's easiest to let the domain resolve directly to your server:

1. In Cloudflare, set the `@` and `www` records to **DNS only** (grey cloud).
2. Open `https://example.com` and wait until the site loads with a valid certificate (check the domain's **SSL** page in OpenPanel).
3. Switch the records back to **Proxied** (orange cloud), with SSL mode **Full (strict)**.

Renewals then work through Cloudflare automatically.

### Option B: Cloudflare Origin Certificate

A Cloudflare Origin Certificate is free, valid for up to 15 years, and trusted by Cloudflare only - perfect when the site is **always** proxied.

1. In Cloudflare, go to **SSL/TLS → Origin Server → Create Certificate**.
2. Keep **RSA (2048)**, add `example.com` and `*.example.com`, choose the validity, and click **Create**.
3. Copy the **Origin Certificate** and the **Private Key** (the key is shown only once).

Then install it in OpenPanel as a custom certificate:

1. Go to **OpenPanel → Domains** and click **SSL** next to the domain.
2. Under **Configure custom SSL**, paste the **Origin Certificate** into **Certificate** and the **Private Key** into **Private Key** - including the `-----BEGIN ...-----` and `-----END ...-----` lines.
3. Click **Configure Custom Certificate**. The SSL status changes to **Custom SSL**.

![Configure custom SSL form with fields for the certificate and the private key](/img/openpanel-screenshots/domains/ssl-custom.png#gh-light-mode-only)
![Configure custom SSL form with fields for the certificate and the private key](/img/openpanel-screenshots/domains/ssl-custom_dark.png#gh-dark-mode-only)

Repeat for each domain (and subdomain) that uses the certificate. More details: [Install a custom SSL certificate](/docs/articles/domains/install-custom-ssl-certificate/) and [SSL](/docs/panel/domains/ssl/#custom-ssl).

:::caution
Browsers don't trust Origin Certificates. If you ever switch the record to **DNS only**, visitors see a certificate warning - switch the domain back to Let's Encrypt on its SSL page first.
:::

---

## Recommended Cloudflare Settings

- **SSL/TLS → Edge Certificates → Always Use HTTPS**: On.
- **Automatic HTTPS Rewrites**: On.
- **Minimum TLS Version**: 1.2.
- **Caching**: Cloudflare caches static files by default. Don't use "Cache Everything" on WordPress, WooCommerce or other logged-in sites without bypass rules for `/wp-admin` and cart/checkout pages.

---

## Fixing Common Problems

### `ERR_TOO_MANY_REDIRECTS`

SSL mode is **Flexible**. Change it to **Full (strict)**. Also check that no `.htaccess` rule redirects based on `%{HTTPS}` - behind a proxy, use `X-Forwarded-Proto` instead (see [Redirect HTTP to HTTPS](/docs/articles/domains/redirect-http-https-www-domain/)).

### Error 525 (SSL handshake failed) or 526 (invalid SSL certificate)

Your server has no valid certificate yet for the domain, or it expired. Use Option A (grey cloud until Let's Encrypt is issued) or install an Origin Certificate (Option B).

### Error 521 / 522 (web server down / connection timed out)

Cloudflare can't reach the server. Check that the `A` record has the right IP, the server is up, and ports `80`/`443` are open. If your server firewall only allows certain IPs, allow [Cloudflare's IP ranges](https://www.cloudflare.com/ips/).

### Let's Encrypt certificate won't renew

Make sure SSL mode is **Full (strict)** and no Cloudflare rule blocks `/.well-known/acme-challenge/`. See [Let's Encrypt rate limits and renewal failures](/docs/articles/domains/lets-encrypt-rate-limits-renewal-failures/).

### Email stopped working

The `mail` / MX hostname is proxied. Set it to **DNS only**.

---

## Real Visitor IPs in Logs and Statistics

Behind Cloudflare, requests arrive from Cloudflare's IP addresses. Cloudflare passes the real visitor IP in the `CF-Connecting-IP` and `X-Forwarded-For` headers.

- **OpenPanel and OpenAdmin login pages** read `CF-Connecting-IP` automatically.
- For **website logs, statistics and IP blocking**, the administrator can make the proxy trust Cloudflare's ranges: OpenPanel ships a ready-made snippet in `/etc/openpanel/caddy/templates/cf.proxy` - import it in the Caddyfile and add `import cloudflare_trusted_proxies` in the domain's `reverse_proxy` blocks, as described in the file.
- In **WordPress**, plugins like *Cloudflare* or security plugins can use the `CF-Connecting-IP` header.

---

## Hosting Without a Public IP

Want to use Cloudflare without opening any ports on your server - for example at home? Use a Cloudflare Tunnel instead: [Host websites behind a Cloudflare Tunnel](/docs/articles/server/cloudflared-tunnel-openpanel/).

---

## Related

- [SSL troubleshooting](/docs/articles/domains/ssl-troubleshooting-guide/)
- [Speed up DNS propagation](/docs/articles/domains/speed-up-dns-propagation/)
- [Wildcard SSL certificates](/docs/articles/domains/wildcard-ssl-certificate/)
