---
sidebar_label: "High CPU or RAM: Find the Culprit"
description: "How to find which account, website or service is causing high CPU load or memory usage on an OpenPanel server, how to spot bot traffic and slow requests in the logs, and how to limit it."
---

# High CPU or RAM Usage: How to Find Which Site Is Causing It

When the server is slow, the load is high, or you get *High CPU / RAM usage* notifications, the cause is usually **one account** - a busy site, a bot attack, a stuck cron job or a runaway app. Because OpenPanel runs every account's services in their own containers, you can quickly narrow it down: **server → account → service → website → requests**.

---

## Step 1: Confirm the Problem

In **OpenAdmin → Server → Resource Usage**, check load, CPU and RAM, and open the history to see when it started. See [Resource Usage](/docs/admin/advanced/resource-usage/).

From the terminal:

```bash
uptime            # load average - compare with the number of CPU cores (nproc)
free -h           # RAM and swap
top -o %CPU       # press Shift+M to sort by memory
```

If Sentinel sent a notification, its crash report in `/var/log/openpanel/admin/crashlog/` shows the top CPU and memory processes at that moment.

---

## Step 2: Find the Account

Every account's services run as that account's own containers. List the busiest containers across all accounts:

```bash
source /usr/local/opencli/lib/podman.sh
for u in $(ls /home); do
  [ -f /home/$u/docker-compose.yml ] || continue
  podman_user $u stats --no-stream --format "{{.CPUPerc}} {{.MemUsage}} $u/{{.Name}}" 2>/dev/null
done | sort -rn | head -15
```

Example output:

```
187.40% 912MB / 2GB     john/php-fpm-8.3
 64.10% 1.1GB / 2GB     john/mariadb
  3.29% 327MB / 1GB     shop/n8n
```

Here, account `john` is the culprit - its PHP and database containers.

In the UI, open **OpenAdmin → Accounts → Users → *username*** to see that account's resource usage and history. Users can check their own in **OpenPanel → Resource Usage**.

System services (the panel, Caddy, MySQL for panel data, mail) are listed with `podman stats --no-stream`.

---

## Step 3: Find the Service

The container name tells you what's busy:

| Busy container | Usual cause |
|---|---|
| `php-fpm-X.Y` | Heavy PHP pages, bots, a plugin in a loop, WP-Cron, a big import |
| `mysql` / `mariadb` | Slow or unindexed queries, a huge table, a WooCommerce search |
| `apache` / `nginx` / `openresty` / `openlitespeed` | Lots of requests - traffic spike or bots |
| `cron` | A cron job that runs too often or never finishes |
| Node.js / Python / n8n app | The app itself - check its logs |

Look inside the container:

```bash
opencli docker USERNAME php-fpm-8.3      # opens a shell, then run: top
```

For databases, see the running queries in **OpenPanel → MySQL → Running Queries** ([Show Processes](/docs/panel/mysql/processlist/)).

---

## Step 4: Find the Website and the Requests

An account can have many domains. Caddy writes one access log per domain in `/var/log/caddy/domlogs/<domain>/access.log`.

**Which domain gets the most requests?**

```bash
for d in /var/log/caddy/domlogs/*/; do echo "$(wc -l < $d/access.log) $(basename $d)"; done | sort -rn | head
```

**Top IPs and URLs for a domain** (last 5,000 requests):

```bash
L=/var/log/caddy/domlogs/example.com/access.log
tail -n 5000 $L | jq -r .request.client_ip | sort | uniq -c | sort -rn | head
tail -n 5000 $L | jq -r .request.uri | sort | uniq -c | sort -rn | head
```

**Slowest requests** (over 1 second):

```bash
tail -n 5000 $L | jq -r 'select(.duration > 1) | "\(.duration) \(.request.uri)"' | sort -rn | head
```

Typical findings:

- One IP (or a range) with thousands of requests → a bot or attack.
- Many hits on `/wp-login.php` or `/xmlrpc.php` → brute-force attempts on WordPress.
- Many hits on search, filter or `?add-to-cart=` URLs → crawlers hitting uncached pages.
- The same slow URL over and over → a slow page or plugin to optimize.

Users can see the same data in **OpenPanel → Statistics → Raw Access Logs** and [Visitor Statistics](/docs/panel/statistics/goaccess/).

---

## Step 5: Fix It

**Bad bots and attacks**
- Block IPs or ranges with the [IP Blocker](/docs/panel/security/ip-blocker/) (per account) or [Sentinel Firewall](/docs/admin/security/firewall/) (server-wide).
- Enable the [Web Application Firewall](/docs/panel/security/waf/) for the domain.
- Block aggressive crawlers in `robots.txt`, or put the site behind [Cloudflare](/docs/articles/domains/cloudflare-with-openpanel/).
- Protect `wp-login.php` with [basic authentication](/docs/articles/websites/password-protect-directory/).

**Heavy website**
- Add caching - [Redis object cache for WordPress](/docs/articles/websites/wordpress-redis-object-cache/), Varnish or a page-cache plugin.
- Disable plugins one by one on a [staging copy](/docs/articles/websites/wordpress-staging-site/) to find the slow one.
- Replace WP-Cron with a real cron job - see [WP-CLI](/docs/articles/websites/wordpress-wp-cli/#replace-wp-cron-with-a-real-cron-job).

**Limit the account**
Each account's services run with CPU and memory limits from its hosting plan, so one account can't use the whole server. Lower the limits for the account or its plan - see [Are plan limits hard or soft?](/docs/articles/containers/are-plan-limits-hard-or-soft-limits/) and [Hosting plans](/docs/admin/plans/hosting_plans/).

**Restart a stuck service**
Stop or restart it from **OpenPanel → Services**, or schedule regular restarts: [Restart a service with a cron job](/docs/articles/containers/restart-service-with-cron/).

**The server is simply too small**
If usage is high across many accounts, add RAM/CPU - see [System requirements and sizing](/docs/articles/install-update/system-requirements/) and [Add swap](/docs/articles/server/how-to-add-swap/) as a buffer.

---

## Get Alerts Early

Set CPU, RAM, load and swap thresholds in **OpenAdmin → Settings → Notifications** - Sentinel checks them every 5 minutes and emails you when they're exceeded. See [Notifications](/docs/admin/settings/notifications/).

---

## Related

- [Free up disk space](/docs/articles/server/how-to-free-up-disk-space-on-linux/)
- [Where OpenPanel stores logs](/docs/articles/support/where-openpanel-stores-logs/)
