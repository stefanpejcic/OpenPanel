---
sidebar_position: 9
---

# Show Processes

This interface displays all operations currently running on your MongoDB server, taken from `$currentOp`.

Checking the running operations is useful for finding slow queries, long aggregations or operations stuck waiting for a lock that slow down your application.

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
