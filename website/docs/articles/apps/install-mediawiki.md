---
sidebar_label: "MediaWiki"
description: "How to install MediaWiki (the software behind Wikipedia) with OpenPanel: one-click install, LocalSettings.php, uploads, short URLs, extensions, backups, cloning and updates."
---

# How to Install MediaWiki (Wikipedia's Wiki Software)

[MediaWiki](https://www.mediawiki.org/) is the free software that powers Wikipedia. It's a great fit for documentation, knowledge bases and community wikis. OpenPanel installs it in one click and includes backups, cloning and one-click updates.

---

## Requirements

- **OpenPanel Enterprise** - the **MediaWiki** module must be enabled for your hosting plan.
- A domain or subdomain, e.g. `wiki.example.com`.
- A PHP version that meets the MediaWiki release's minimum - OpenPanel checks this before installing and tells you if the domain's PHP version is too old.

---

## Install MediaWiki

1. Add the domain in **OpenPanel → Domains** ([Add a new domain](/docs/panel/domains/new/)).
2. Go to **OpenPanel → Websites → Install App** and click **Install MediaWiki**.
3. Choose the **domain**, **MediaWiki version** (*Latest*), **site name**, and the **admin username, password and email**.
4. Start the installation.

![Install MediaWiki form with the site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/mediawiki-form.png#gh-light-mode-only)
![Install MediaWiki form with the site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/mediawiki-form_dark.png#gh-dark-mode-only)

OpenPanel downloads the official release, creates a MySQL/MariaDB database, runs MediaWiki's installer, and adds a cron job for MediaWiki's job queue.

Open `https://wiki.example.com`, or click **Login as Admin** on the wiki's manage page.

---

## Common Configuration

All wiki settings live in `LocalSettings.php` in the wiki folder. Edit it with the [File Manager](/docs/panel/files/files/).

### Allow file uploads

```php
$wgEnableUploads = true;
```

Uploaded files go to the `images/` folder. Raise PHP's `upload_max_filesize` if you need larger files - see [How to increase PHP limits](/docs/articles/websites/how-to-set-or-increase-PHP-INI-memory-limit-or-other-values/).

### Restrict who can edit

For a private or company wiki:

```php
$wgGroupPermissions['*']['createaccount'] = false; // no public sign-ups
$wgGroupPermissions['*']['edit'] = false;          // only logged-in users edit
$wgGroupPermissions['*']['read'] = false;          // make the wiki private
```

### Install extensions

Download an extension into `extensions/` and enable it in `LocalSettings.php`:

```php
wfLoadExtension( 'VisualEditor' );
```

Many popular extensions (VisualEditor, Cite, ParserFunctions) are already bundled with MediaWiki - just add the `wfLoadExtension` line.

### Caching

For busy wikis, enable the **Redis** service and set `$wgMainCacheType` to use it - see [Redis](/docs/panel/caching/Redis/) (use `redis` as the host).

---

## Managing the Wiki

The wiki's manage page in **OpenPanel → Websites → Sites** has:

- **Login as Admin** - a one-time admin login link.
- **Update** - updates MediaWiki core and runs the database update script; `LocalSettings.php` and `images/` are kept.
- **Clone** - copies the wiki to another domain, for example for testing an upgrade.
- **Backups** - on-demand backups of the database, files or both, with one-click restore.
- **Logs** and **Remove**.

:::tip
Before **Update**, create a backup in the **Backups** tab - and check that your extensions support the new MediaWiki version.
:::

More: [MediaWiki in OpenPanel](/docs/panel/applications/mediawiki/).

---

## Related

- [Install DokuWiki](/docs/articles/apps/install-dokuwiki/) - a simpler wiki without a database
- [Install SofaWiki](/docs/articles/apps/install-sofawiki/)
- [All installable apps](/docs/panel/applications/autoinstaller/)
