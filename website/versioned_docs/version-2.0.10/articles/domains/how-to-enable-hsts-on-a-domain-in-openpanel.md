# How to Enable HSTS on a Domain in OpenPanel

HSTS ([HTTP Strict Transport Security](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Strict-Transport-Security)) can be enabled either globally for all domains or individually for specific domains in OpenPanel.

---

## Enabling HSTS for a Specific Domain

### As an End User

If you are an OpenPanel user:

1. Ensure you have access to the **edit_vhost** feature. If enabled, you will see **Edit VirtualHosts** under the **Domains** section.
2. Click **Edit VirtualHosts** and select the domain you want to configure.
3. In the editor, add the HSTS header **only within the HTTPS section**.
4. Use the correct syntax depending on your web server (Nginx, Apache, OpenLiteSpeed, etc.).
5. Save your changes.

Example Configurations:

#### OpenLiteSpeed

Add this in the Rewrite/Headers section for the HTTPS listener:
```bash
Header always set Strict-Transport-Security "max-age=2592000; includeSubDomains; preload"
```

#### Apache
Add the following inside the block:
```bash
<IfModule mod_headers.c>
    Header always set Strict-Transport-Security "max-age=2592000; includeSubDomains; preload"
</IfModule>
```

#### Nginx
Add this inside the server block:
```bash
add_header Strict-Transport-Security "max-age=2592000; includeSubDomains; preload" always;
```


---

### As an Administrator

If you have OpenAdmin or root SSH access:

1. Open the domain’s Caddy vhost configuration file:

```bash
/etc/openpanel/caddy/domains/DOMAIN_NAME.TLD.conf
```

2. Add the following HSTS header **just before the `tls {` line**:

```bash
# HSTS
header {
  Strict-Transport-Security "max-age=2592000; includeSubDomains; preload"
}
```

3. Save the file and reload Caddy.

---

## Enabling HSTS for All Domains

To apply HSTS automatically to all domains on the server:

1. Edit the default domain templates:

```bash
/etc/openpanel/caddy/templates/domain.conf_with_modsec
/etc/openpanel/caddy/templates/domain.conf
```

2. Add the HSTS header **just before the `tls {` line** in each template:

```bash
# HSTS
header {
  Strict-Transport-Security "max-age=2592000; includeSubDomains; preload"
}
```

3. Save the templates. All new domains created afterward will inherit this HSTS configuration.
