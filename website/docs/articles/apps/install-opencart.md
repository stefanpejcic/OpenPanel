---
sidebar_label: "OpenCart"
description: "How to install OpenCart and launch an online store with OpenPanel: one-click install, admin login, SEO URLs, securing the storage folder, PHP settings and email."
---

# How to Install OpenCart and Launch an Online Store

[OpenCart](https://www.opencart.com/) is a lightweight, free and open-source shopping cart. It's simple to run and fast on modest hardware. OpenPanel installs it in one click.

---

## Requirements

- **OpenPanel Enterprise** - the **OpenCart** module must be enabled for your hosting plan.
- A domain pointed to the server.
- A PHP version supported by your OpenCart release (PHP 8.0+ for OpenCart 4.x).

---

## Install OpenCart

1. Add the domain in **OpenPanel → Domains** ([Add a new domain](/docs/panel/domains/new/)).
2. Go to **OpenPanel → Websites → Install App** and click **Install OpenCart**.
3. Choose the **domain**, **OpenCart version** (*Latest*), and the **admin username, password and email**.
4. Start the installation.

![Install OpenCart form with the site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/opencart-form.png#gh-light-mode-only)
![Install OpenCart form with the site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/opencart-form_dark.png#gh-dark-mode-only)

OpenPanel downloads the official release, creates a MySQL/MariaDB database, prepares `config.php` and `admin/config.php`, runs OpenCart's CLI installer, and deletes the `install/` folder afterwards.

The store opens at `https://example.com`, the admin at `https://example.com/admin/`. You can also use **Login as Admin** on the store's manage page in **OpenPanel → Websites → Sites**.

---

## Recommended Settings

### Use HTTPS everywhere

SSL is issued automatically. Make sure `HTTP_SERVER` and `HTTPS_SERVER` in `config.php` and `admin/config.php` start with `https://`.

### SEO-friendly URLs

In the admin, go to **System → Settings → Edit store → Server** and enable **Use SEO URLs**.

- On **Apache** and **OpenLiteSpeed**, rename `.htaccess.txt` to `.htaccess` in the store folder.
- On **Nginx** and **OpenResty**, add OpenCart's rewrite rules in **Domains → Edit VHosts File** (Enterprise, see [Edit VHosts](/docs/panel/webserver/vhosts/)).

### Move the storage folder (OpenCart 4)

After the first login, OpenCart suggests moving `system/storage/` outside the web root. Follow the dashboard notice - it moves the folder for you and updates the config.

### Email

In **System → Settings → Mail**, choose **SMTP** and enter your mailbox details so order emails are delivered - see [Why do my emails go to spam?](/docs/articles/email/why-emails-go-to-spam/).

---

## Managing the Store

The manage page shows versions, files and database details, and has:

- **Login as Admin** - one-time login link, valid for 10 minutes.
- **Cache** - the same as *Refresh cache* in the admin; use it after installing extensions.
- **Logs** - `system/storage/logs/error.log`.
- **Remove** - uninstalls the store and its database.

Use the account-level [Backups](/docs/panel/backups/backups/) before updating OpenCart or installing extensions.

More: [OpenCart in OpenPanel](/docs/panel/applications/opencart/).

---

## Related

- [Install PrestaShop](/docs/articles/apps/install-prestashop/)
- [Increase PHP limits](/docs/articles/websites/how-to-set-or-increase-PHP-INI-memory-limit-or-other-values/)
- [All installable apps](/docs/panel/applications/autoinstaller/)
