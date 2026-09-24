---
sidebar_label: "Wildcard SSL Certificates"
description: "Does OpenPanel support wildcard SSL certificates (*.example.com)? No - but every domain and subdomain gets its own free Let's Encrypt certificate automatically. Here's how that works and what to use instead."
---

# Does OpenPanel Support Wildcard SSL Certificates?

**No.** OpenPanel does not issue or install wildcard SSL certificates (`*.example.com`).

You don't need one: **every domain and subdomain you add gets its own free SSL certificate automatically**, including its `www` version. Add `shop.example.com`, `blog.example.com` and `api.example.com`, and each one is secured on its own - no extra steps.

---

## How SSL Works in OpenPanel

- When you add a domain or subdomain, OpenPanel registers it with its built-in proxy (Caddy).
- The first time the domain is opened over `https://`, a free **Let's Encrypt** certificate is issued for `domain` and `www.domain`.
- Certificates are **renewed automatically** before they expire.

Certificates are only issued for domains that exist in OpenPanel - not for random subdomains that happen to point to your server. This protects the server from certificate abuse and rate-limit problems.

See [SSL](/docs/panel/domains/ssl/).

---

## Why Wildcard Certificates Aren't Used

A wildcard certificate from Let's Encrypt can only be issued with a **DNS-01 challenge**, which needs API access to the domain's DNS provider. OpenPanel issues certificates automatically for any domain, wherever its DNS is hosted, so it uses the HTTP-based challenge that works for every domain - and that challenge can't issue wildcards.

---

## What to Use Instead

| You want... | Do this |
|---|---|
| SSL on many subdomains | Add each subdomain on the **Domains** page - each gets its own certificate automatically. See [Add a subdomain](/docs/articles/domains/subdomain-addon-parked-domain/). |
| SSL on dynamic subdomains (SaaS, multisite) | Add the subdomains you use as domains in OpenPanel, or put the site behind [Cloudflare](/docs/articles/domains/cloudflare-with-openpanel/), which can serve a wildcard certificate at its edge. |
| A paid certificate (OV/EV) | Install it as a [custom SSL certificate](/docs/articles/domains/install-custom-ssl-certificate/) on each domain it covers. |

:::info Installing an existing wildcard certificate
If you already bought a wildcard certificate, you can still install it as a **custom certificate** on each domain or subdomain it covers - see [Install a custom SSL certificate](/docs/articles/domains/install-custom-ssl-certificate/). Renewing it is then up to you.
:::

---

## Related

- [Let's Encrypt rate limits and renewal failures](/docs/articles/domains/lets-encrypt-rate-limits-renewal-failures/)
- [SSL troubleshooting](/docs/articles/domains/ssl-troubleshooting-guide/)
