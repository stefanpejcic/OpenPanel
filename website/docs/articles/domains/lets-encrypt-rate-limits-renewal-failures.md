---
sidebar_label: "Let's Encrypt Rate Limits & Renewals"
description: "Why Let's Encrypt SSL certificates fail to issue or renew in OpenPanel: rate limits (too many certificates, failed validations), DNS and firewall problems, CAA records and Cloudflare - and how to fix each one."
---

# Let's Encrypt Rate Limits and SSL Renewal Failures

OpenPanel issues and renews free **Let's Encrypt** certificates automatically. Most of the time you never have to think about it - but when a certificate fails to issue or renew, the cause is almost always one of the few problems below.

---

## How Issuance and Renewal Work

- A certificate is requested the **first time a domain is opened over `https://`**, for the domain and its `www`.
- Let's Encrypt checks that the domain really points to your server (an HTTP challenge on port `80`, or a TLS challenge on port `443`).
- Certificates are valid for a short period and are **renewed automatically** well before they expire.
- If issuance fails, it's retried automatically with increasing delays, so a temporary problem fixes itself once the cause is gone.

To see what's happening for a domain, administrators can run:

```bash
opencli domains-ssl example.com status   # AutoSSL, Custom SSL or no SSL
opencli domains-ssl example.com info     # certificate details and expiry
opencli domains-ssl example.com logs     # recent issuance/renewal log lines
```

---

## Let's Encrypt Rate Limits

Let's Encrypt limits how many certificates can be requested, to protect its service. The limits you're most likely to hit:

| Limit | Value | Typical cause |
|---|---|---|
| **Failed validations** | 5 per hostname, per hour | DNS not pointing to the server yet, port 80/443 blocked |
| **Duplicate certificates** | 5 per exact set of names, per week | Reinstalling the server or deleting and re-adding the same domain many times |
| **Certificates per registered domain** | 50 per week (e.g. for `example.com` and all its subdomains) | Adding many subdomains of one domain at once |
| **New orders** | 300 per account, per 3 hours | Migrating hundreds of domains at once |

See the current values in the [Let's Encrypt documentation](https://letsencrypt.org/docs/rate-limits/).

### What to do when you hit a limit

- **Wait.** Limits reset on a rolling window - an hour for failed validations, a week for duplicate certificates.
- **Fix the cause first**, so the next automatic retry succeeds (see below). Repeatedly clicking **Generate now** while DNS is still wrong only adds more failed validations.
- **Don't reinstall or delete domains** to "reset" SSL - that uses up the duplicate-certificate limit.
- **Migrating many domains?** Move DNS in batches, so certificates are requested gradually.

OpenPanel's proxy (Caddy) also falls back to another free certificate authority when Let's Encrypt is unavailable or rate-limited, so most sites still get a valid certificate.

---

## Common Reasons Issuance or Renewal Fails

### 1. DNS doesn't point to the server

The domain's `A` record (and `www`) must point to the server's IPv4 address, and DNS must be fully propagated.

```bash
dig +short example.com
dig +short www.example.com
```

Both should return your server IP. Check globally with [whatsmydns.net](https://www.whatsmydns.net/#A). If only `www` is missing, the certificate for `www` fails - add the record. See [Troubleshooting DNS](/docs/articles/domains/troubleshooting-dns/).

### 2. An `AAAA` (IPv6) record points somewhere else

Let's Encrypt prefers IPv6. If the domain has an `AAAA` record for an old server or an IPv6 address that isn't configured on this server, validation fails. Remove the wrong `AAAA` record, or make sure it points to this server.

### 3. Ports 80 or 443 are blocked

Validation needs incoming connections on **port 80 and 443**. Check any firewall in front of the server (cloud security groups, Hetzner/OVH/Oracle firewalls) as well as [Sentinel Firewall](/docs/admin/security/firewall/).

### 4. Cloudflare proxy

With Cloudflare's orange cloud on, Cloudflare answers the validation requests. Use **Full (strict)** SSL mode, or temporarily switch the record to **DNS only** while the first certificate is issued - see [Use Cloudflare with OpenPanel](/docs/articles/domains/cloudflare-with-openpanel/).

### 5. A CAA record blocks Let's Encrypt

A **CAA** DNS record lists which authorities may issue certificates for the domain. If it exists and doesn't include Let's Encrypt, issuance is refused:

```bash
dig +short CAA example.com
```

Add `0 issue "letsencrypt.org"` to the domain's CAA records, or remove the CAA record.

### 6. A redirect or the WAF interferes

A redirect from the domain to another site, or a strict web application firewall rule, can block the validation request under `/.well-known/acme-challenge/`. Temporarily remove the redirect and try again - see [403 errors](/docs/articles/domains/error-on-website-disable-coraza-waf/).

### 7. The domain isn't in OpenPanel

Certificates are issued only for domains added on the **Domains** page (and their `www`). Other hostnames pointing to the server - for example `test.example.com` when only `example.com` is added - don't get a certificate. Add them as domains.

---

## Force a New Attempt

Once the cause is fixed:

1. Go to **OpenPanel → Domains → SSL** for the domain.
2. Click **Generate now** (or **Switch to Let's Encrypt and generate** if it was using a custom certificate).
3. Open `https://example.com` in a browser.

If it still fails, check the logs with `opencli domains-ssl example.com logs` and see the [SSL troubleshooting guide](/docs/articles/domains/ssl-troubleshooting-guide/).

---

## Related

- [Wildcard SSL certificates](/docs/articles/domains/wildcard-ssl-certificate/)
- [Install a custom SSL certificate](/docs/articles/domains/install-custom-ssl-certificate/)
- [SSL in OpenPanel](/docs/panel/domains/ssl/)
