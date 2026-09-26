---
sidebar_position: 9
---

# Show Processes

This interface displays all operations currently running on your MongoDB server, taken from `$currentOp`.

Checking the running operations is useful for finding slow queries, long aggregations or operations stuck waiting for a lock that slow down your application.

![MongoDB Running Queries table listing the current operations with their user, namespace, running time and a Kill button](/img/openpanel-screenshots/mongodb/processlist-list.png#gh-light-mode-only)
![MongoDB Running Queries table listing the current operations with their user, namespace, running time and a Kill button](/img/openpanel-screenshots/mongodb/processlist-list_dark.png#gh-dark-mode-only)

Each row shows:

- **ID**: the operation ID.
- **User**: the MongoDB user that started the operation.
- **Client**: where the connection comes from.
- **Application**: the application name the client driver reported, if any.
- **Operation**: the operation type, for example `query`, `insert`, `update` or `command`. Operations waiting for a lock are marked with *waiting for lock*. Internal MongoDB tasks show their name instead, for example `TTLMonitor`.
- **Namespace**: the database and collection, for example `app.orders`.
- **Time**: how many seconds the operation has been running.
- **Command**: the command that was sent. Long commands are shortened, hover over one to see more.

Use the search box to filter by user, namespace, operation or command. Click **Refresh Processes** to reload the list.

## Kill a query

To stop a long-running operation, click **Kill** next to it. The button changes to **Confirm** with a 5 second countdown. Click it again before the countdown ends to kill the operation.

This runs `killOp`, so only the operation is stopped and the connection stays open. The application that sent it gets an error for that operation.

The **Kill** button only shows for operations that are currently running and came from a client connection. Internal MongoDB tasks can not be killed.

Every killed operation is recorded in the account's activity log.

## Bulk Actions

Tick the checkbox of one or more operations, or the checkbox in the table header to select every operation shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![A running operation selected with the bulk actions bar offering Kill](/img/openpanel-screenshots/mongodb/processlist-bulk.png#gh-light-mode-only)
![A running operation selected with the bulk actions bar offering Kill](/img/openpanel-screenshots/mongodb/processlist-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Kill** | Kills the selected operations. |

Only rows that show a **Kill** button can be selected.

Click an action, confirm it, and it runs on the selected operations one after another. When it's done the page reloads with a notice listing the operations it worked for, or which ones failed and why.
