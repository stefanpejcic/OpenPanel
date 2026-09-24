---
sidebar_position: 7
---

# Blocked User Agents

Block visitors whose HTTP `User-Agent` header matches an entry on a blacklist — useful for keeping out known bad bots and scrapers.

Use **OpenAdmin > Security > Blocked User Agents** to manage it. The page is available to the **Super Admin** and **Admin** roles - resellers can't access it.

![Blacklist user agents page with the Enable dropdown and the list of blocked user agents](/img/openadmin-screenshots/security/blacklist-useragents-page.png#gh-light-mode-only)
![Blacklist user agents page with the Enable dropdown and the list of blocked user agents](/img/openadmin-screenshots/security/blacklist-useragents-page_dark.png#gh-dark-mode-only)

## Enable

Set **Enable** to **Yes** and click **Save settings** to turn user-agent blocking on server-wide. Set it to **No** to disable it.

## Editing the List

Edit the list of blocked user-agent strings (one per line) in the text area, then click **Save settings**.

The list is stored at `/etc/openpanel/openpanel/conf/blacklist_useragents.txt` on the server.
