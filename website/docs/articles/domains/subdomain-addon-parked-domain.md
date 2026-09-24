---
sidebar_label: "Subdomains, Addon & Parked Domains"
description: "How to add a subdomain, an addon domain or a parked (alias) domain in OpenPanel. All domains are equal - each can have its own website, SSL and DNS, or share a site with another domain."
---

# How to Add a Subdomain, Addon Domain or Parked Domain

If you're coming from cPanel, you're used to three separate pages: **Subdomains**, **Addon Domains** and **Parked (Alias) Domains**. In OpenPanel there's just one: **Domains**. **All domains are equal** - every domain or subdomain you add gets its own:

- website folder (document root),
- free SSL certificate,
- DNS records,
- web server configuration and statistics.

What makes a domain a "subdomain", "addon" or "parked" domain is only how you set it up.

| cPanel term | In OpenPanel |
|---|---|
| **Subdomain** (`blog.example.com`) | Add `blog.example.com` on the Domains page |
| **Addon domain** (a second website, `another.com`) | Add `another.com` on the Domains page |
| **Parked / alias domain** (`example.net` shows the same site as `example.com`) | Add `example.net` with the **same document root** as `example.com` - or redirect it |

---

## Add a Subdomain

1. Go to **OpenPanel → Domains** and click **Add Domain**.
2. Enter the full subdomain, e.g. `blog.example.com`.
3. Optionally change the **Document Root** (default: `/var/www/html/blog.example.com`).
4. Click **Add Domain**.

When the main domain (`example.com`) is already in your account:

- the subdomain's DNS records are **added to the main domain's DNS zone** - nothing to change at your registrar if `example.com` already uses your nameservers,
- the subdomain **does not count** toward your plan's domain limit.

If `example.com` is hosted elsewhere, add an `A` record for `blog` pointing to your server's IP at your current DNS provider.

You can now install WordPress or any other app on the subdomain, just like on a main domain - see [Install Apps](/docs/panel/applications/autoinstaller/).

---

## Add an Addon Domain (Another Website)

An addon domain is simply another domain with its own website.

1. Go to **OpenPanel → Domains → Add Domain**.
2. Enter the domain, e.g. `another.com`, and click **Add Domain**.
3. Point the domain to the server: either set its **nameservers** at the registrar to your server's nameservers (see [Configure nameservers](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)), or create an `A` record for `another.com` and `www` pointing to the server's IP.

Addon domains count toward your plan's domain limit.

---

## Add a Parked (Alias) Domain

A parked domain shows the **same website** as another domain - for example `example.net` and `example.co` showing `example.com`. There are two ways:

### Option 1: Same content on both domains

1. Go to **OpenPanel → Domains → Add Domain** and enter `example.net`.
2. Set **Document Root** to the **main domain's folder**, e.g. `/var/www/html/example.com`.
3. Click **Add Domain**.

Both domains now serve the same files, each with its own SSL certificate.

:::caution
Most CMSs, including WordPress, redirect visitors to their own configured address. For WordPress, Option 2 is usually better. Showing the same content on two domains can also hurt SEO (duplicate content).
:::

### Option 2: Redirect to the main domain (recommended)

1. Add `example.net` on the Domains page (any document root).
2. Click **Redirect** next to it and enter `https://example.com{uri}` - the `{uri}` part keeps the page path.

Visitors and search engines are sent to your main domain. See [How to redirect a domain](/docs/articles/domains/redirect-http-https-www-domain/#redirect-a-domain-to-another-domain-or-url).

---

## Change or Remove a Domain

- **Move a site to another folder**: [Change document root](/docs/panel/domains/docroot/).
- **Temporarily disable a domain**: [Suspend a domain](/docs/panel/domains/suspend/).
- **Remove a domain**: delete it from the Domains page. For a subdomain, this also removes its records from the main domain's DNS zone.

---

## Troubleshooting

| Problem | Fix |
|---|---|
| New subdomain shows the "default page" | Upload files to its document root or install an app. See [Domain shows default page](/docs/articles/domains/domain-shows-default-page/). |
| "Domain limit reached" | Your plan's domain limit counts main domains and addon domains - not subdomains of domains you already own. |
| Subdomain doesn't resolve | The main domain uses nameservers elsewhere - add the `A` record there. [Speed up DNS propagation](/docs/articles/domains/speed-up-dns-propagation/). |
| Can't add a subdomain of a domain owned by another account | The server administrator controls this with the `permit_subdomain_sharing` setting. |

---

## Related

- [Redirect HTTP to HTTPS, www or another domain](/docs/articles/domains/redirect-http-https-www-domain/)
- [Add new domain](/docs/panel/domains/new/)
- [Troubleshooting DNS](/docs/articles/domains/troubleshooting-dns/)
