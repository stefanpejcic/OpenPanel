---
sidebar_label: "Password-Protect a Directory"
description: "How to password-protect a website or folder with HTTP basic authentication (.htpasswd) in OpenPanel on Apache, Nginx, OpenResty and OpenLiteSpeed."
---

# How to Password-Protect a Directory (Basic Authentication)

HTTP **basic authentication** shows a username/password prompt in the browser before anyone can open a folder or a whole website. It's a quick way to hide a staging site, an admin area, a client folder or a work-in-progress site.

How you set it up depends on the web server your account uses:

| Web server | Where the rules go |
|---|---|
| **Apache** | `.htaccess` file in the folder |
| **Nginx** / **OpenResty** | Domain's VHost file |
| **OpenLiteSpeed** | Domain's VHost file (a *realm*) |

Not sure which one you use? Check **OpenPanel → Server Information**, or see [switching the web server](/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/).

---

## Step 1: Create the Password File (`.htpasswd`)

The password file contains one `username:encrypted-password` line per user. Store it **outside** the website folder, so it can never be downloaded - for example:

```
/var/www/html/.htpasswds/example.com/.htpasswd
```

(`/var/www/html/example.com` is the website, so `/var/www/html/.htpasswds/` is not reachable from the web.)

### Generate the password line

Open **OpenPanel → Containers → Terminal** (Enterprise), select your **web server** or **PHP** container, and run one of these:

```bash
# Apache container - bcrypt
htpasswd -nbB john 'YourStrongPassword'

# any container with openssl - APR1/MD5, works on every web server
openssl passwd -apr1 'YourStrongPassword'
```

The first command prints the full line (`john:$2y$05$...`). For the second, put the username and a colon in front of the output:

```
john:$apr1$Qm0y2rBv$0cZs2rO4mTgSx7PzT1s3m1
```

Then create the file with the [File Manager](/docs/panel/files/files/): make the folder `.htpasswds/example.com/` in your home files, create `.htpasswd` inside it, and paste the line(s). Add one line per user.

:::tip No terminal?
Any trusted *htpasswd generator* can produce the same line - choose the **APR1 (MD5)** or **bcrypt** format. Never use plain-text passwords in the file.
:::

---

## Step 2a: Apache (`.htaccess`)

Create or edit `.htaccess` inside the folder you want to protect, for example `/var/www/html/example.com/private/.htaccess`:

```apache
AuthType Basic
AuthName "Restricted area"
AuthUserFile /var/www/html/.htpasswds/example.com/.htpasswd
Require valid-user
```

To protect the **whole website**, put the `.htaccess` in the domain's root folder. The change works immediately - no restart needed.

---

## Step 2b: Nginx and OpenResty

Nginx ignores `.htaccess` files, so the rules go in the domain's configuration. Go to **Domains → Edit VHosts File** for the domain ([Edit VHosts](/docs/panel/webserver/vhosts/), Enterprise) and add a `location` block **inside both** `server { }` blocks (the port 80 one and the port 443 one):

```nginx
location ^~ /private/ {
    auth_basic           "Restricted area";
    auth_basic_user_file /var/www/html/.htpasswds/example.com/.htpasswd;

    try_files $uri $uri/ /index.php?$args;

    location ~ \.php$ {
        include fastcgi_params;
        fastcgi_pass php-fpm-8.3:9000;   # use the same PHP line as the rest of the file
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    }
}
```

To protect the **whole website**, add the two `auth_basic` lines directly inside the existing `location / { ... }` blocks instead - and also inside the `location ~ \.php$` blocks, so PHP files are protected too.

Save the file - OpenPanel reloads the web server.

---

## Step 2c: OpenLiteSpeed

OpenLiteSpeed reads rewrite rules from `.htaccess`, but **not** authentication rules. Add a *realm* and a *context* in the domain's VHost file (**Domains → Edit VHosts File**):

```
realm RestrictedArea {
  userDB  {
    location              /var/www/html/.htpasswds/example.com/.htpasswd
  }
}

context /private/ {
  location                $DOC_ROOT/private/
  allowBrowse             1
  realm                   RestrictedArea
  authName                Restricted area
  required                user *
}
```

Use `context /` with `location $DOC_ROOT/` to protect the whole website. Save, and the web server is reloaded.

---

## Test It

Open the protected URL in a **private browser window** - you should see a login prompt. Or from a terminal:

```bash
curl -I https://example.com/private/                       # 401 Unauthorized
curl -I -u john:YourStrongPassword https://example.com/private/  # 200 OK
```

---

## Good to Know

- **Always use HTTPS.** Basic auth sends the password with every request; SSL keeps it encrypted. OpenPanel issues SSL certificates automatically.
- **Varnish**: requests with a login are never cached, so protected pages always come fresh from your site.
- **WordPress**: protecting `/wp-admin/` with basic auth is a good extra layer against brute-force attacks. Exclude `/wp-admin/admin-ajax.php` if your theme or plugins use it on the front end.
- **Remove protection** by deleting the `.htaccess` lines (Apache) or the added VHost blocks.
- To block visitors by **IP address** instead of a password, use the [IP Blocker](/docs/panel/security/ip-blocker/).

---

## Related

- [Share .htaccess rules across websites](/docs/articles/websites/sharing-htaccess-rules/)
- [Set up a WordPress staging site](/docs/articles/websites/wordpress-staging-site/)
- [Fix 403 Forbidden errors](/docs/articles/domains/error-on-website-disable-coraza-waf/)
