---
sidebar_label: "Redis Object Cache for WordPress"
description: "How to enable Redis object caching for WordPress in OpenPanel: start the Redis service, install the Redis Object Cache plugin, configure wp-config.php, verify it works, and run several sites on one Redis."
---

# How to Enable Redis Object Cache for WordPress

A **Redis object cache** stores the results of WordPress database queries in memory, so repeated page loads don't hit MySQL again. It makes the WordPress admin, WooCommerce stores, membership sites and any logged-in pages noticeably faster.

OpenPanel includes a Redis service that each account can start on its own, as long as the **Redis** feature is enabled for the account's hosting plan (see [Feature Manager](/docs/admin/plans/feature-manager/)). This guide covers the whole setup: start Redis, install the plugin, connect it, and check that it works.

:::tip Redis vs. page cache
Redis caches **database queries** (object cache). A page cache (like Varnish or a caching plugin) caches **whole pages** for visitors who aren't logged in. They work well together.
:::

---

## Step 1: Start the Redis Service

1. Go to **OpenPanel → Cache → Redis**.
2. Click **Click to Enable**.

![Redis Status row with the Click to Enable or Click to Disable button](/img/openpanel-screenshots/caching/redis-status.png#gh-light-mode-only)
![Redis Status row with the Click to Enable or Click to Disable button](/img/openpanel-screenshots/caching/redis-status_dark.png#gh-dark-mode-only)

Redis starts in its own container and is reachable from your websites at:

- **Host**: `redis` - not `127.0.0.1` or `localhost`
- **Port**: `6379`

![Redis page with the service status, TCP server and port, container resource usage and logs](/img/openpanel-screenshots/caching/redis-page.png#gh-light-mode-only)
![Redis page with the service status, TCP server and port, container resource usage and logs](/img/openpanel-screenshots/caching/redis-page_dark.png#gh-dark-mode-only)

Optionally click **Edit limits** to give Redis more memory - 128-256 MB is plenty for most sites. See [Redis](/docs/panel/caching/Redis/).

:::info Valkey
Prefer [Valkey](https://valkey.io/), the open-source Redis fork? Enable it under **Cache → Valkey** instead and use `valkey` as the host below - the plugin works the same way.
:::

---

## Step 2: Tell WordPress Where Redis Is

Open the site's `wp-config.php` in the [File Manager](/docs/panel/files/files/) and add these lines **above** `/* That's all, stop editing! */`:

```php
define( 'WP_REDIS_HOST', 'redis' );
define( 'WP_REDIS_PORT', 6379 );
define( 'WP_REDIS_PREFIX', 'example.com:' );   // unique per site
define( 'WP_REDIS_TIMEOUT', 1 );
define( 'WP_REDIS_READ_TIMEOUT', 1 );
```

Or with [WP-CLI](/docs/articles/websites/wordpress-wp-cli/):

```bash
wp config set WP_REDIS_HOST redis --allow-root
wp config set WP_REDIS_PORT 6379 --raw --allow-root
wp config set WP_REDIS_PREFIX 'example.com:' --allow-root
```

`WP_REDIS_PREFIX` keeps each site's keys separate, so several WordPress sites can safely share one Redis service.

---

## Step 3: Install and Enable the Plugin

1. In WordPress, go to **Plugins → Add New**, search for **Redis Object Cache** (by Till Krüss), then **Install** and **Activate** it.
2. Go to **Settings → Redis** and click **Enable Object Cache**.

With WP-CLI:

```bash
wp plugin install redis-cache --activate --allow-root
wp redis enable --allow-root
```

The plugin creates `wp-content/object-cache.php`, which makes WordPress use Redis.

---

## Step 4: Check That It Works

- In WordPress, **Settings → Redis** should show **Status: Connected**, with the host `redis`.
- In OpenPanel, open the site in **WordPress (WP Manager)** - the **Cache** card shows the current cache type, which should now be **Redis**.
- With WP-CLI: `wp redis status --allow-root` and `wp cache type --allow-root`.

Browse the site and the admin for a minute, then check **Settings → Redis → Diagnostics** - the hit ratio should climb above 80-90%.

---

## Clearing the Cache

- In WP Manager, click **Clear Cache** on the **Cache** card.
- Or in WordPress: **Settings → Redis → Flush Cache**.
- Or: `wp cache flush --allow-root`.

![Cache card with the cache type and the Clear Cache button](/img/openpanel-screenshots/applications/wordpress-cache.png#gh-light-mode-only)
![Cache card with the cache type and the Clear Cache button](/img/openpanel-screenshots/applications/wordpress-cache_dark.png#gh-dark-mode-only)

Flush the cache after migrating or restoring a site, or when changes don't show up.

---

## Troubleshooting

| Problem | Fix |
|---|---|
| **Status: Not connected** / *Connection refused* | The Redis service is stopped - enable it in **Cache → Redis**. Make sure `WP_REDIS_HOST` is `redis`, not `127.0.0.1`. |
| Site shows **Error establishing a Redis connection** | Redis was stopped while `object-cache.php` is active. Start Redis again, or delete `wp-content/object-cache.php` to disable the object cache. |
| Two sites show each other's content | Both use the same keys - give each site a different `WP_REDIS_PREFIX`. |
| Redis memory keeps filling up | Raise Redis's memory limit with **Edit limits**, or flush the cache. Restarting Redis clears all cached data. |
| No speed difference for visitors | Object cache mostly helps dynamic and logged-in pages. For anonymous visitors, add page caching as well. |

---

## Related

- [Redis](/docs/panel/caching/Redis/)
- [WP Manager](/docs/panel/applications/wordpress/)
- [Use WP-CLI](/docs/articles/websites/wordpress-wp-cli/)
- [Purge the Varnish cache](/docs/articles/websites/purge-varnish-cache-from-terminal/)
