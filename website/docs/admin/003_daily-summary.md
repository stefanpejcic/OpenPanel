---
sidebar_position: 3
---

# Daily summary

The **Daily summary** tab, next to **Dashboard**, **Notifications** and **License**, shows one day of the whole server: the [daily summaries](/docs/panel/statistics/daily_summary/) of all OpenPanel accounts added up, plus the server's own health, mail, backups, PageSpeed checks and web firewall. On a [multi-server](/docs/admin/system/multi-server/) master it covers every server in the cluster, and each row says which server it's on.

An orange exclamation mark on the tab means something needs attention. It goes away once you open the tab, until the summary of a newer day needs attention again.

## How it's made

Nothing is worked out when you open the page, it only reads a file saved earlier:

| Time | What runs | What it saves |
|---|---|---|
| 01:45 | `opencli email-server pflogsumm` | the mail report for the previous day |
| 04:00 | `opencli websites-pagespeed --all` | the PageSpeed check of every website, a copy per day in `/etc/openpanel/openpanel/websites/history/` |
| 06:15 | `opencli user-daily_summary --all` | each account's summary in `/etc/openpanel/openpanel/core/users/<user>/summaries/<date>.json` |
| 06:45 | `openadmin daily-summary` | the admin summary of the previous day in `/etc/openpanel/openadmin/summaries/<date>.json` |

`openadmin daily-summary` adds up the account summaries and reads the server's own logs for that day: the 5 minute Sentinel checks, the notifications, the mail report, the backup log (also after logrotate moved it), the PageSpeed copies and the web firewall logs. On a cluster master it asks every server for its part and saves the combined day. The files are kept for 30 days.

To make or remake a day by hand, for example after fixing something:

```bash
/usr/local/admin/openadmin-amd64 daily-summary --date 2026-11-18
```

Run it on the servers that are members first, then on the master. Without `--date` it makes yesterday.

Resellers see only their own accounts, worked out from the account summaries when they open the page.

## What it shows

Use the day picker on the right of the tab bar to open one of the saved days.

### Totals

The top of the page shows the day, how many servers and accounts it covers, and a one-line summary, with the totals for that day:

- **Requests**, with the change from the day before
- **Visitors**
- **Data sent**
- **Server errors**
- **WAF blocks**

### Needs attention

Issues found on accounts, like a scheduled task that keeps failing, a website returning errors, a plugin with known security holes or malware found by ImunifyAV (always on top). They are grouped by account and website, with the website's favicon, and each account links to its page (on a master through the server it's on). When there are more than 8, the rest are counted. Smaller findings are counted as "worth a look".

When one problem is shared by at least 40% (and 3 or more) of the accounts that need attention, a line names it, for example *The main problem is All-in-One WP Migration Unlimited Extension: it has known security holes on 14 of the 28 accounts that need attention (50%), updating it clears most of the list.*

### Busiest websites

The 5 websites with the most requests, with their favicon, owner and a bar for their share.

### Slowest websites

The 5 lowest Google PageSpeed scores from that day's checks, mobile first since most visitors are on phones, with the desktop score and when the first content shows and the page becomes usable on mobile. Only checks made on that day count. Shown when the PageSpeed checks ran, which needs a Google API key in `/etc/openpanel/openpanel/service/pagespeed.api`.

### Busiest databases

The database servers with the most queries, each account's MySQL (logo), MariaDB (logo), PostgreSQL or MongoDB on its own row, with the queries, size, a red badge for errors and a bar for the share of queries.

### Web firewall

Shown when the [web firewall](/docs/admin/security/waf/) blocked something that day: how many requests it blocked, the rules that caught the most, the most attacked websites and the addresses that were blocked the most. For websites behind Cloudflare the address is the visitor's real IP and country from the access log, not Cloudflare's.

### Server health

Peaks from the 5 minute Sentinel checks (load against the CPU cores, CPU, RAM, disk) and the alerts raised that day. The same alert for several IPs, like an admin logging in from new IPs, is one row with the addresses under it, and several admins logging in from the same new IP are one row too.

### Backups

When account backups are [configured centrally](/docs/admin/backups/), the result of that day's run: how many accounts were backed up, failed or skipped, how long it took, the size and the largest archives. A schedule set from **Scheduled actions** (for example `0 22 * * *`) shows as *Custom*.

### Mail

From each mail server's summary report for that day: received, delivered, deferred, bounced and rejected messages. Deferrals count delivery retries, not messages.

## When it's shown

The tab is shown only when the **`summary`** module is enabled on [**Settings > Modules**](/docs/admin/settings/modules/#daily-summary).

Account summaries are saved each day for users that have the **summary** feature in their [feature set](/docs/admin/plans/feature-manager/). Until the first admin summary is saved the page says there is no summary yet.
