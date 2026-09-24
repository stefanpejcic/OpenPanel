---
sidebar_label: "Joomla"
description: "How to install Joomla with OpenPanel in one click: choosing a version, the Super User account, one-time admin login, SEF URLs, cache, PHP settings and backups."
---

# How to Install Joomla

[Joomla](https://www.joomla.org/) is a popular open-source CMS for websites, portals and online magazines. OpenPanel installs Joomla in one click and gives you a one-click admin login from the panel.

---

## Requirements

- **OpenPanel Enterprise** - the **Joomla** module must be enabled for your hosting plan.
- A domain pointed to the server.
- A PHP version supported by your Joomla version (PHP 8.1+ for Joomla 5).

---

## Install Joomla

1. Add the domain in **OpenPanel → Domains** ([Add a new domain](/docs/panel/domains/new/)).
2. Go to **OpenPanel → Websites → Install App** and click **Install Joomla**.
3. Fill in:
   - **Domain** (leave the subfolder empty to install at the root),
   - **Site name**,
   - **Joomla version** - *Latest* is recommended,
   - **Admin full name, username, password and email** - the Super User account.
4. Start the installation.

![Install Joomla form with the site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/joomla-form.png#gh-light-mode-only)
![Install Joomla form with the site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/joomla-form_dark.png#gh-dark-mode-only)

OpenPanel downloads the official Joomla package, creates a MySQL/MariaDB database and runs Joomla's CLI installer. The site is ready at `https://example.com` and the administrator at `https://example.com/administrator/`.

---

## After Installing

### Log in without a password

On the site's manage page in **OpenPanel → Websites → Sites**, click **Login as Admin** to open the Joomla administrator already logged in. Each link works once and expires after 10 minutes.

### Enable search-engine friendly URLs

In **System → Global Configuration → Site**, enable **Search Engine Friendly URLs** and **Use URL Rewriting**:

- On **Apache** and **OpenLiteSpeed**, rename `htaccess.txt` to `.htaccess` in the site folder.
- On **Nginx** and **OpenResty**, no file is needed - requests are already routed through `index.php`.

### PHP settings

Joomla's **System Information** page shows recommended PHP settings. To raise upload limits or memory, see [How to increase PHP limits](/docs/articles/websites/how-to-set-or-increase-PHP-INI-memory-limit-or-other-values/).

### Email

Configure **System → Global Configuration → Server → Mail** to send through SMTP for better delivery.

---

## Managing Joomla

The manage page shows the Joomla, PHP and database versions, files and database details, and has:

- **Cache** - clears Joomla's system cache.
- **Logs** - PHP warnings and errors from `administrator/logs/`.
- **Remove** - uninstalls Joomla, including the database and files.

Update Joomla itself from **System → Update → Joomla** in the administrator. For backups, use the account-level [Backups](/docs/panel/backups/backups/).

More: [Joomla in OpenPanel](/docs/panel/applications/joomla/).

---

## Related

- [Install WordPress](/docs/articles/websites/how-to-install-wordpress-with-openpanel/)
- [Install Drupal](/docs/articles/apps/install-drupal/)
- [All installable apps](/docs/panel/applications/autoinstaller/)
