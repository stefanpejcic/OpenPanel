---
sidebar_label: "Use WP-CLI"
description: "How to use WP-CLI with WordPress sites in OpenPanel: run wp commands from the web terminal or with a cron job, common commands for plugins, users, search-replace, cache and database."
---

# How to Use WP-CLI in OpenPanel

[WP-CLI](https://wp-cli.org/) is the command-line tool for WordPress. It lets you update plugins, reset passwords, search and replace URLs or export the database in seconds. WP-CLI is **preinstalled** in every OpenPanel PHP container - it's what WP Manager itself uses behind the scenes.

Many common tasks don't need WP-CLI at all: the [WP Manager](/docs/panel/applications/wordpress/) has buttons for updates, cache, debugging, maintenance mode, backups and cloning.

---

## Where to Run WP-CLI

WordPress runs in your domain's PHP container (for example `php-fpm-8.3`), so WP-CLI commands run there too. Two ways:

### Option 1: Web terminal (Enterprise)

1. Go to **OpenPanel → Containers → Terminal**.
2. Select the PHP container your site uses, e.g. `php-fpm-8.3`.
3. Run commands with the site path and `--allow-root`:

```bash
wp plugin list --path=/var/www/html/example.com --allow-root
```

Or go to the folder first, then leave out `--path`:

```bash
cd /var/www/html/example.com
wp plugin list --allow-root
```

See [Terminal](/docs/panel/containers/terminal/).

### Option 2: Cron job with "Run" (any plan with cron jobs)

1. Go to **OpenPanel → Cron Jobs → Create New**.
2. **Container**: `php-fpm-8.3`.
3. **Command**: your WP-CLI command, e.g. `wp core version --path=/var/www/html/example.com --allow-root`.
4. Save, then click **Run** to execute it immediately and see the output.

Delete the job afterwards if it was a one-time command - or keep it with a schedule for recurring tasks.

:::info Why `--allow-root`?
Inside the container, commands run as the container's root user (which maps to your own unprivileged account on the server). WP-CLI refuses to run as root without `--allow-root`.
:::

---

## Common Commands

All examples assume you are in the site folder (`cd /var/www/html/example.com`).

For more commands with examples, and fixes for the most common WP-CLI errors, see [Most useful WP-CLI commands with examples](https://wpxss.com/wp-cli/most-useful-wp-cli-commands-with-examples-how-to-fix-most-common-wpcli-errors/).

### Core, plugins and themes

```bash
wp core version --allow-root
wp core update --allow-root
wp plugin list --allow-root
wp plugin update --all --allow-root
wp plugin deactivate --all --allow-root        # troubleshooting a broken site
wp plugin activate woocommerce --allow-root
wp theme activate twentytwentyfive --allow-root
```

### Users and passwords

```bash
wp user list --allow-root
wp user update admin --user_pass='NewStrongPassword' --allow-root
wp user create jane jane@example.com --role=administrator --allow-root
```

### Search and replace (after a domain change)

```bash
wp search-replace 'https://old.com' 'https://new.com' --all-tables --dry-run --allow-root
wp search-replace 'https://old.com' 'https://new.com' --all-tables --allow-root
```

Always run with `--dry-run` first.

### Database

```bash
wp db export backup.sql --allow-root
wp db import backup.sql --allow-root
wp db optimize --allow-root
```

Keep exports outside the public folder, or delete them afterwards - a `.sql` file in the site root can be downloaded by anyone.

### Cache and rewrite rules

```bash
wp cache flush --allow-root
wp rewrite flush --allow-root
wp transient delete --all --allow-root
```

### Site settings

```bash
wp option get siteurl --allow-root
wp option update home 'https://example.com' --allow-root
wp option update siteurl 'https://example.com' --allow-root
wp maintenance-mode activate --allow-root
```

---

## Replace WP-Cron with a Real Cron Job

WordPress runs scheduled tasks (`wp-cron.php`) on page visits, which can be unreliable on quiet sites and slow on busy ones. To use a real cron job instead:

1. Add to `wp-config.php`: `define('DISABLE_WP_CRON', true);`
2. Create a cron job in the PHP container, **Every 5 Minutes**:

```bash
wp cron event run --due-now --path=/var/www/html/example.com --allow-root
```

---

## Related

- [Set up a WordPress staging site](/docs/articles/websites/wordpress-staging-site/)
- [Fix the WordPress white screen / critical error](/docs/articles/websites/wordpress-white-screen-critical-error/)
- [Cron Jobs](/docs/panel/cronjobs/)
