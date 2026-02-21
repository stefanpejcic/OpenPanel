# Setting Custom SSL Certificates for OpenPanel, OpenAdmin, and Webmail

⚠️ Currently, adding custom SSL certificates for OpenPanel and Webmail is only possible from the terminal.

- Let's Encrypt certificates are stored in `/etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/`
- Custom certificates must be placed in `/etc/openpanel/caddy/ssl/custom/`

---

## Custom SSL for OpenAdmin and OpenPanel

Follow these steps to configure a custom SSL certificate for OpenAdmin and OpenPanel interfaces:

### 1. Add the certificate

Create a directory `custom` in `/etc/openpanel/caddy/ssl/`:

```bash
mkdir -p /etc/openpanel/caddy/ssl/custom/
```

Then create directory name same as your domain
```bash
mkdir -p /etc/openpanel/caddy/ssl/acme-v02.api.letsencrypt.org-directory/YOUR_DOMAIN_HERE/
```

Upload your SSL files to this directory and name them:

* `$DOMAIN.crt`
* `$DOMAIN.key`

**Example:**

```
/etc/openpanel/caddy/ssl/custom/srv.openpanel.com/srv.openpanel.com.crt
/etc/openpanel/caddy/ssl/custom/srv.openpanel.com/srv.openpanel.com.key
```

### 2. Set the domain

```bash
opencli domain set <DOMAIN_NAME>
```

**Example:**
```bash
opencli domain set srv.openpanel.com
```

### 3. Configure Caddy

Open the **Caddyfile**:
```bash
nano /etc/openpanel/caddy/Caddyfile
```

Find the section for your domain:

```
# START HOSTNAME DOMAIN #
example.net {
    reverse_proxy localhost:2087
}
```

Add the `tls` line after `reverse_proxy`:

```
tls /etc/openpanel/caddy/ssl/custom/srv.openpanel.com/srv.openpanel.com.crt /etc/openpanel/caddy/ssl/custom/srv.openpanel.com/srv.openpanel.com.key
```

👉 Replace `srv.openpanel.com` with your actual domain.

Restart Caddy to apply the certificate:

```bash
docker restart caddy
```

---

## Custom SSL for Webmail

### 1. Add the certificate

Create a directory for your domain:

```bash
mkdir -p /etc/openpanel/caddy/ssl/custom/YOUR_DOMAIN_HERE/
```

Upload your SSL files and name them:

* `$DOMAIN.crt`
* `$DOMAIN.key`

**Example:**

```
/etc/openpanel/caddy/ssl/custom/webmail.openpanel.com/webmail.openpanel.com.crt
/etc/openpanel/caddy/ssl/custom/webmail.openpanel.com/webmail.openpanel.com.key
```

### 2. Assign the domain

```bash
opencli email-webmail domain <DOMAIN_NAME>
```

**Example:**

```bash
opencli email-webmail domain webmail.openpanel.com
```

### 3. Configure Caddy

Open the **Caddyfile**:

```bash
nano /etc/openpanel/caddy/Caddyfile
```

Locate the Webmail section:

```
# START WEBMAIL DOMAIN #
webmail.example.net {
    reverse_proxy localhost:8080
}
# END WEBMAIL DOMAIN #
```

Add the `tls` line after `reverse_proxy`:

```
tls /etc/openpanel/caddy/ssl/custom/webmail.openpanel.com/webmail.openpanel.com.crt /etc/openpanel/caddy/ssl/custom/webmail.openpanel.com/webmail.openpanel.com.key
```

👉 Replace `webmail.openpanel.com` with your domain.

Restart Caddy to apply the changes:

```bash
docker restart caddy
```

---

## Custom SSL for Domains (End-Users)

End-users can add their own SSL certificates directly from the **OpenPanel interface** if the `ssl` module and features are enabled.

📖 Documentation: [Custom SSL for Domains](https://openpanel.com/docs/panel/domains/ssl/#custom-ssl)
