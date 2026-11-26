---
sidebar_position: 2
---

# Queue

View and manage mail that's queued but not yet delivered by the mailserver (Postfix).

Use **OpenAdmin > Emails > Queue** to access it.

## Queue table

Each row shows one queued message, including its queue ID, sender, recipient(s), size, and the reason delivery is being retried.

The table shows 100 messages per page with **Previous** and **Next** below it. Above the table are the total and the count per queue: click **hold**, **deferred** or **active** to show only that queue. The search box finds messages by queue ID, sender, recipient or reason across the whole queue, and sorting a column keeps the search and the queue filter. The queue is read once and reused for 30 seconds, so paging through a large queue is fast; click **Refresh** to read it again. A retry or delete always reads it again.

![Email Queue page with the queued messages table and the Refresh button](/img/openadmin-screenshots/emails/queue-page.png#gh-light-mode-only)
![Email Queue page with the queued messages table and the Refresh button](/img/openadmin-screenshots/emails/queue-page_dark.png#gh-dark-mode-only)

## Actions

- **Retry** – Force an immediate delivery retry, for a **selected** message or for **all** queued messages.
- **Delete** – Permanently remove a message from the queue, for a **selected** message or for **all** queued messages.

Bulk actions (retry/delete **all**) apply to the entire queue at once, not only the current page or search results, use with care. The checkbox in the table header selects the messages on the current page.
