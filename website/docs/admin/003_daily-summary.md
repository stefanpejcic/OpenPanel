---
sidebar_position: 3
---

# Daily Summary

The **Daily Summary** tab, next to **Dashboard**, **Notifications** and **License**, adds up the [daily summaries](/docs/panel/statistics/daily_summary/) of all OpenPanel accounts for the most recent day that has them, up to 7 days back. On a [multi-server](/docs/admin/system/multi-server/) master it covers the accounts of every server. Resellers see only their own accounts.

An orange exclamation mark on the tab means something needs attention. It goes away once you open the tab, until the summary of a newer day needs attention again.

## What it shows

The top of the page shows the day, how many servers and accounts it covers, and a one-line summary, with the totals for that day:

- **Requests**, with the change from the day before
- **Visitors**
- **Data sent**
- **Server errors**
- **WAF blocks**

Below that:

- **Needs attention**: issues found on accounts, like a scheduled task that keeps failing or a website returning errors. Each one links to the account, and on a master shows which server it's on. When there are more than 8, the rest are counted. Smaller findings are counted as "worth a look".
- **Busiest websites**: the 5 websites with the most requests, with a bar for their share.

## When it's shown

The tab is shown only when the **`summary`** module is enabled on [**Settings > Modules**](/docs/admin/settings/modules/#daily-summary).

Summaries are saved each day at 06:15 for the previous day, for users that have the **summary** feature in their [feature set](/docs/admin/plans/feature-manager/). Until the first one is saved the page says there is no summary yet.
