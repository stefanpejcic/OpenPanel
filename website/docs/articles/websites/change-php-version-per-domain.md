---
sidebar_label: "Change PHP Version per Domain"
description: "How to change the PHP version of a website in OpenPanel: select a PHP version per domain, set the default version for new domains, check compatibility, and change it from the terminal."
---

# How to Change the PHP Version per Domain

OpenPanel lets every website run its own PHP version - an old app can stay on PHP 7.4 while a new site uses PHP 8.4, on the same account. Each version runs in its own PHP-FPM container, started automatically when a domain needs it.

---

## Change the PHP Version of a Website

1. Go to **OpenPanel → PHP → Select PHP Version**.
2. Find the domain and choose the new version from the dropdown.
3. Save.

OpenPanel updates the domain's web server configuration and reloads it. If the chosen PHP version isn't running yet, its container is started - the first switch to a new version can take a few seconds.

The page also shows each version's support status, so you can spot sites on **outdated or unsupported** PHP versions.

See [Select PHP Version](/docs/panel/php/domains/).

:::info OpenLiteSpeed
On **OpenLiteSpeed**, the PHP version can't be set per domain - all domains use one version. Change it on the **Default PHP Version** page instead.
:::

### For WordPress sites

In **OpenPanel → WordPress**, open the site's **Manage** page - the PHP version is shown there too and can be changed for sites installed at the domain root.

---

## Set the Default Version for New Domains

Go to **OpenPanel → PHP → Default PHP Version** and choose the version new domains should use. Existing domains keep their current version. See [Default PHP Version](/docs/panel/php/default/).

---

## Before You Switch: Check Compatibility

Newer PHP versions are faster and more secure, but old code can break:

1. **Update your CMS, plugins and themes first.** Most compatibility problems come from outdated plugins.
2. **Test on a copy.** For WordPress, create a [staging site](/docs/articles/websites/wordpress-staging-site/), switch its PHP version, and click through the site.
3. **Check the requirements.** For example, WordPress recommends PHP 8.3+, Laravel 11 needs PHP 8.2+, Drupal 11 needs PHP 8.3+.
4. **Keep a backup.**

If the site breaks after switching (white page, "critical error", 500 error), switch back to the previous version - it takes effect immediately. Then see [Fix the WordPress white screen / critical error](/docs/articles/websites/wordpress-white-screen-critical-error/).

---

## PHP Settings and Extensions per Version

Each PHP version has its own settings. After switching, check that your settings and extensions are there for the new version:

- **php.ini values** (`memory_limit`, `upload_max_filesize`, ...) - **PHP → PHP.INI Editor** or **PHP → Options**, per version. See [Increase PHP limits](/docs/articles/websites/how-to-set-or-increase-PHP-INI-memory-limit-or-other-values/).
- **Per-website overrides** with `.user.ini` - see [PHP settings per website](/docs/articles/websites/php-user-ini-files/).
- **Extensions** (ionCube, Redis, imagick, ...) - see [How to install a PHP extension](/docs/articles/websites/how-to-install-php-extensions-in-openpanel/).

---

## From the Terminal (Administrators)

```bash
# show a domain's PHP version
opencli php-domain example.com

# change it
opencli php-domain example.com --update 8.4

# default version for new domains of a user
opencli php-default USERNAME --update 8.4
```

See [opencli php](/docs/articles/opencli/php/).

---

## Check Which Version a Site Uses

Create `phpinfo.php` in the site folder:

```php
<?php phpinfo();
```

Open `https://example.com/phpinfo.php`, check the version at the top, then **delete the file** - it reveals server details.

---

## Related

- [Host a PHP website](/docs/articles/websites/hosting-a-php-website-with-openpanel/)
- [Deploy a Laravel app](/docs/articles/websites/deploy-laravel-app/)
