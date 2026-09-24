---
sidebar_label: "Redirects: HTTPS, www, Domain"
description: "How to redirect HTTP to HTTPS, www to non-www (or the reverse), and one domain to another in OpenPanel, including keeping the URL path and choosing 301 vs 302 redirects."
---

# How to Redirect HTTP to HTTPS, www to non-www, or One Domain to Another

This guide covers the three most common redirects and how each one works in OpenPanel:

| Redirect | How |
|---|---|
| **HTTP → HTTPS** | Automatic once the domain has an SSL certificate |
| **www ↔ non-www** | `www` works automatically - choose one version with a redirect rule |
| **Domain → another domain** | The **Redirects** feature on the Domains page |

---

## HTTP to HTTPS

When you add a domain, OpenPanel requests a free **Let's Encrypt** SSL certificate for it automatically. Once the certificate is issued, visitors are sent to the `https://` version of your site - you don't need to add any rules.

For this to work:

- The domain (and `www`) must point to the server's IP, so the certificate can be issued.
- Ports `80` and `443` must be open.

If the site still loads over `http://`, check the SSL status in **OpenPanel → Domains → SSL** and see [SSL troubleshooting](/docs/articles/domains/ssl-troubleshooting-guide/).

To tell browsers to *always* use HTTPS, even before the first redirect, enable [HSTS](/docs/articles/domains/how-to-enable-hsts-on-a-domain-in-openpanel/).

:::tip Fallback rule
If you want an explicit HTTP→HTTPS rule in the site itself (for example for an app that checks the scheme), add this to `.htaccess` on Apache / OpenLiteSpeed:

```apache
RewriteEngine On
RewriteCond %{HTTP:X-Forwarded-Proto} =http
RewriteRule ^ https://%{HTTP_HOST}%{REQUEST_URI} [L,R=301]
```

Check `X-Forwarded-Proto`, not `%{HTTPS}` - SSL is handled by the proxy in front of your web server, so `%{HTTPS}` is always off and would cause a redirect loop.
:::

---

## www to non-www (or non-www to www)

Every domain you add in OpenPanel automatically:

- gets a `www` DNS record (a `CNAME` to the main domain) in its DNS zone,
- answers on both `example.com` and `www.example.com`,
- gets an SSL certificate for both.

So `www` works out of the box, **as long as `www` points to the server**. If your DNS is hosted elsewhere (Cloudflare, your registrar), add the record there:

```
www    CNAME    example.com.
```

Both versions show the same site. For SEO it's better to choose one and redirect the other to it:

### WordPress and most CMSs

WordPress redirects automatically to the address set in **Settings → General → Site Address (URL)**. Set it to `https://example.com` or `https://www.example.com` - no other rules needed. Joomla, Drupal and others have similar settings.

### Apache / OpenLiteSpeed (`.htaccess`)

Redirect `www` to non-www:

```apache
RewriteEngine On
RewriteCond %{HTTP_HOST} ^www\.(.+)$ [NC]
RewriteRule ^ https://%1%{REQUEST_URI} [L,R=301]
```

Redirect non-www to `www`:

```apache
RewriteEngine On
RewriteCond %{HTTP_HOST} !^www\. [NC]
RewriteRule ^ https://www.%{HTTP_HOST}%{REQUEST_URI} [L,R=301]
```

### Nginx / OpenResty

`.htaccess` files are ignored by Nginx. Add the rule at the top of the domain's `server` blocks in **Domains → Edit VHosts File** ([Edit VHosts](/docs/panel/webserver/vhosts/), Enterprise):

```nginx
if ($host = www.example.com) {
    return 301 https://example.com$request_uri;
}
```

Not sure which web server you use? See [switching the web server](/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/).

---

## Redirect a Domain to Another Domain or URL

Use the **Redirects** feature to send all traffic for a domain somewhere else - for example after a rebrand, or to point extra domains (`example.net`, `example.org`) to your main site.

1. Go to **OpenPanel → Domains**.
2. Click **Redirect** next to the domain.
3. Enter the destination URL, starting with `http://` or `https://`, and save.

The redirect is applied by the proxy **before** the request reaches your web server, so it works even if the domain has no website files. It applies to both `example.net` and `www.example.net`.

:::info
The **Domain Redirects** feature is part of **OpenPanel Enterprise** and must be enabled for your plan. See [Redirects](/docs/panel/domains/redirects/).
:::

### Keep the page path

By default every page redirects to the exact URL you entered - `old.com/blog/post` goes to `https://new.com`. To keep the path, add `{uri}` to the end of the destination:

```
https://new.com{uri}
```

Now `old.com/blog/post?id=1` redirects to `https://new.com/blog/post?id=1`.

### 301 vs 302

Domain redirects created from the Domains page are **temporary (302)** redirects. That's fine for parked and extra domains. When you **move a site permanently** and want search engines to transfer rankings to the new domain, use a **301** instead - keep the old domain's website active with a rule like this in its `.htaccess`:

```apache
RewriteEngine On
RewriteRule ^ https://new.com%{REQUEST_URI} [L,R=301]
```

or in its Nginx vhost:

```nginx
return 301 https://new.com$request_uri;
```

Then add the new domain in Google Search Console and use its **Change of address** tool.

### Remove a redirect

On the Domains page, click **Redirect** next to the domain and choose **Delete** to remove the rule.

---

## Troubleshooting

| Problem | Fix |
|---|---|
| `ERR_TOO_MANY_REDIRECTS` | Two rules redirect back and forth - often an `.htaccess` HTTPS rule using `%{HTTPS}`, or Cloudflare set to *Flexible*. See [Cloudflare with OpenPanel](/docs/articles/domains/cloudflare-with-openpanel/). |
| `www` shows "site can't be reached" | The `www` DNS record is missing or points elsewhere. |
| Certificate error on `www` | `www` didn't resolve to the server when the certificate was requested. Fix DNS, then wait a few minutes - the certificate is requested again on the next visit. |
| Redirect not working after changing it | Browsers cache 301 redirects. Test in a private window or with `curl -I https://example.com`. |

---

## Related

- [Add a subdomain, addon or parked domain](/docs/articles/domains/subdomain-addon-parked-domain/)
- [Why does my domain show the default page?](/docs/articles/domains/domain-shows-default-page/)
- [Custom SSL certificates](/docs/articles/domains/install-custom-ssl-certificate/)
