---
sidebar_position: 11
---

# Show Processes

This interface displays all currently active MySQL queries (connections). 

Checking the currently running queries is useful for identifying slow queries that slow down the website loading speed.

![MySQL Processes table with the Refresh Processes button and the currently running queries](/img/openpanel-screenshots/mysql/processlist-list.png#gh-light-mode-only)
![MySQL Processes table with the Refresh Processes button and the currently running queries](/img/openpanel-screenshots/mysql/processlist-list_dark.png#gh-dark-mode-only)

Click **Refresh Processes** to reload the table with the current list of active processes.

## Kill a query

To stop a long-running query, click **Kill** next to it. The button changes to **Confirm** with a 5 second countdown. Click it again before the countdown ends to kill the query.

This runs `KILL QUERY`, so only the running query is cancelled and the connection stays open. The application that sent the query gets an error for that query.

The **Kill** button only shows for queries that are currently running. System threads and idle connections can not be killed.

Every killed query is recorded in the account's activity log.

## Bulk Actions

Tick the checkbox of one or more queries, or the checkbox in the table header to select every query shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![Two running queries selected with the bulk actions bar offering Kill](/img/openpanel-screenshots/mysql/processlist-bulk.png#gh-light-mode-only)
![Two running queries selected with the bulk actions bar offering Kill](/img/openpanel-screenshots/mysql/processlist-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Kill** | Kills the selected queries. |

Only rows that show a **Kill** button can be selected.

Click an action, confirm it, and it runs on the selected queries one after another. When it's done the page reloads with a notice listing the queries it worked for, or which ones failed and why.
