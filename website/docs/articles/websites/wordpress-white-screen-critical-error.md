---
sidebar_label: "WordPress White Screen / Critical Error"
description: "How to fix the WordPress white screen of death and 'There has been a critical error on this website' in OpenPanel: enable debugging, read the error log, disable plugins and themes, raise PHP memory, and restore a backup."
---

# How to Fix the WordPress White Screen or "Critical Error"

Two of the most common WordPress problems look like this:

- a completely **blank white page** (the "white screen of death"), or
- **"There has been a critical error on this website."**

Both mean PHP hit a fatal error - almost always caused by a **plugin**, a **theme**, **PHP memory** running out, or a **PHP version** the code doesn't support. This guide shows how to find the cause and fix it in OpenPanel, even when you can't log in to wp-admin.

---

## Step 1: Check the Recovery Email

For a critical error, WordPress emails the site's admin address with the subject *"Your Site is Experiencing a Technical Issue"*. It names the plugin or theme that failed and contains a **recovery mode** link that lets you log in with that plugin disabled.

If you got the email, follow the link, deactivate or update the named plugin, and you're done. If not, continue.

---

## Step 2: Turn On Debugging and Read the Error

1. Open **OpenPanel → WordPress** and click **Manage** on the site.
2. Go to the **Debugging** tab and enable **WP_DEBUG** and **WP_DEBUG_LOG** (leave **WP_DEBUG_DISPLAY** off on a live site, so visitors don't see errors).
3. Reload the broken page.
4. Open `wp-content/debug.log` in the [File Manager](/docs/panel/files/files/) and scroll to the bottom.

You'll see a line like:

```
PHP Fatal error:  Uncaught Error: Call to undefined function ... in /var/www/html/example.com/wp-content/plugins/some-plugin/some-file.php:123
```

The path tells you the culprit - here, the `some-plugin` plugin.

Also check the PHP container's log in **Containers → Logs** for the domain's PHP container (e.g. `php-fpm-8.3`).

:::tip
Turn debugging off again when you're finished - `debug.log` can grow large and may reveal file paths.
:::

---

## Step 3: Fix the Cause

### A plugin

Disable the plugin without wp-admin, in the [File Manager](/docs/panel/files/files/): rename its folder, e.g. `wp-content/plugins/some-plugin` → `some-plugin.off`. WordPress deactivates it automatically.

Not sure which plugin? Rename the whole `wp-content/plugins` folder to `plugins.off`. If the site comes back, rename it back and disable plugins one by one to find the broken one.

With WP-CLI you can do the same in one line - see [Use WP-CLI](/docs/articles/websites/wordpress-wp-cli/):

```bash
wp plugin deactivate some-plugin --allow-root
wp plugin deactivate --all --allow-root
```

### The theme

If the error points to `wp-content/themes/your-theme/`, rename that theme's folder. WordPress falls back to a default theme (install one, like *Twenty Twenty-Five*, if none is present).

### PHP memory limit

If the log says `Allowed memory size of ... bytes exhausted`, raise PHP's `memory_limit` to `256M` or `512M` - see [How to increase PHP memory_limit](/docs/articles/websites/how-to-set-or-increase-PHP-INI-memory-limit-or-other-values/). You can also add to `wp-config.php`:

```php
define( 'WP_MEMORY_LIMIT', '256M' );
```

### PHP version

After switching PHP versions, old plugins can break (`Deprecated`, `Fatal error: Uncaught TypeError`, `Call to undefined function`). Switch the domain back to the previous PHP version, then update the plugin - see [Change the PHP version per domain](/docs/articles/websites/change-php-version-per-domain/).

### A failed update

If an update was interrupted, delete the `.maintenance` file from the site root, then reinstall WordPress core from WP Manager's **Security** tab (**Reinstall WordPress core**) - this replaces core files without touching your content.

---

## Step 4: Restore a Backup (If Nothing Else Works)

If the site broke after a change you can't undo, restore the last working backup from WP Manager's **Backups** tab (files, database or both) or from the account [Backups](/docs/panel/backups/backups/).

---

## Other Blank-Page Causes

| Symptom | Likely cause |
|---|---|
| White page **only in wp-admin** | A plugin or theme that loads in the admin area only - check `debug.log`. |
| **500 Internal Server Error** instead of a white page | Often a broken `.htaccess` (Apache). Rename it and re-save *Settings → Permalinks*. |
| **502 Bad Gateway** | The PHP container isn't running or crashed - see [502 errors](/docs/articles/domains/bad-gateway-502-error-troubleshooting/). |
| **Error establishing a database connection** | See [Error establishing a database connection](/docs/articles/databases/error-establishing-a-database-connection/). |
| **403 Forbidden** | The web application firewall may be blocking a request - see [403 errors](/docs/articles/domains/error-on-website-disable-coraza-waf/). |

---

## Related

- [WP Manager - Debugging](/docs/panel/applications/wordpress/#debugging)
- [Set up a WordPress staging site](/docs/articles/websites/wordpress-staging-site/) - test updates before they break production
