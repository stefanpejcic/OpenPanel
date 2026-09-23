---
sidebar_position: 10
---

# Show Processes

![PostgreSQL Processes table listing the active backend processes](/img/openpanel-screenshots/postgresql/processlist-list.png#gh-light-mode-only)
![PostgreSQL Processes table listing the active backend processes](/img/openpanel-screenshots/postgresql/processlist-list_dark.png#gh-dark-mode-only)

This interface displays all currently active PostgreSQL queries (connections) from `pg_stat_activity`.

Checking the currently running queries is useful for identifying slow queries that slow down your application.

Use the **Show Columns** dropdown to choose which columns are visible in the table (e.g. `pid`, `usename`, `client_addr`, `state`, `query`, and more). Your column selection is remembered between visits.
