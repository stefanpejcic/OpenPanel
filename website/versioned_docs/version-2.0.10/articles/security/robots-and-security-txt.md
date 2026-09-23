# Customize robots.txt and security.txt

How to control search engine crawling and publish a security contact for your panel

---

Both **OpenPanel** and **OpenAdmin** ship with a default `robots.txt` (blocks all crawlers) and a `security.txt` (an [RFC 9116](https://www.rfc-editor.org/rfc/rfc9116) file that tells security researchers how to report a vulnerability). Each app serves its own pair of files, and each has its own override location — replacing them is a matter of dropping a file on disk, there's no editor built into either panel's UI.

| App | Served at | Override directory | Takes effect |
|---|---|---|---|
| OpenPanel | `https://yourpanel:2083/robots.txt`, `/security.txt` | `/etc/openpanel/openpanel/static/` | After restarting the `openpanel` service |
| OpenAdmin | `https://yourserver:2087/robots.txt`, `/security.txt` | `/usr/local/admin/` | Immediately, no restart |

---

## OpenPanel

Create the file at `/etc/openpanel/openpanel/static/robots.txt` and/or `/etc/openpanel/openpanel/static/security.txt`:

```bash
nano /etc/openpanel/openpanel/static/robots.txt
```

The override is only checked once, at startup, so restart the service to apply it:

```bash
cd /root && docker compose up -d openpanel
```

The same directory also holds the CSS/JS overrides — `css/custom.css` and `js/custom.js` under `/etc/openpanel/openpanel/static/` — if you're customizing one you may want the others too; see [Branding & White-Label](/docs/articles/dev-experience/customizing-openpanel-user-interface#set-a-custom-color-scheme).

## OpenAdmin

Create the file at `/usr/local/admin/robots.txt` and/or `/usr/local/admin/security.txt`:

```bash
nano /usr/local/admin/security.txt
```

Unlike OpenPanel, OpenAdmin checks this directory on every request, so the change is live as soon as you save the file — no restart needed.

The same directory also accepts a `custom.css` override for OpenAdmin's own interface.

## Verify

```bash
curl https://yourpanel:2083/robots.txt
curl https://yourpanel:2083/security.txt
```

## Defaults

If no override file exists, both apps fall back to these built-in defaults.

`robots.txt` — blocks all crawlers:

```
User-agent: *
Disallow: /
```

`security.txt`:

```
Contact: mailto:info@openpanel.com
Expires: 2030-12-12T11:00:00.000Z
Preferred-Languages: rs, en
Policy: https://github.com/stefanpejcic/OpenPanel/security/policy
```

Use these as a starting point — at minimum, update `Contact` and `Policy` to point at your own security reporting process before publishing your own `security.txt`.
