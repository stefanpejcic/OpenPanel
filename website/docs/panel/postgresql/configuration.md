---
sidebar_position: 11
---

# Configuration

![PostgreSQL Configuration page with a table of settings, editable values and the Save Changes button](/img/openpanel-screenshots/postgresql/configuration-page.png#gh-light-mode-only)
![PostgreSQL Configuration page with a table of settings, editable values and the Save Changes button](/img/openpanel-screenshots/postgresql/configuration-page_dark.png#gh-dark-mode-only)

The PostgreSQL Configuration page lets you edit low-level `postgresql.conf` settings for your database service, without needing terminal access.

## How to Change Configuration

1. Open **OpenPanel** and navigate to **PostgreSQL > Configuration**.
2. Modify the values you want to change.
3. Click **Save Changes** to apply changes.

Settings are written to a configuration override for your account, and the PostgreSQL container is restarted to apply them.

## Optimize Database

When the Configuration page opens, OpenPanel checks your PostgreSQL service in the background and suggests settings that fit how it's actually used. While it checks you'll see *Checking your database for possible optimizations...*.

If it finds something to improve, an **Optimize Database** banner shows at the top of the page and an optimize icon (⚡) shows next to each setting that has a suggestion.

1. Click the optimize icon next to a setting, or **Review all** in the banner.
2. The **Confirm Changes** dialog shows the previous value, the new value and why it's suggested for each setting.
3. Click **Save Changes** to apply them and restart the service, or **Cancel** to close the dialog without changing anything.

The banner also shows what the suggestions are based on: the PostgreSQL version, the memory limit and current memory use of the service, the number of databases, their total size and the uptime.

Suggestions are based on:

- **Memory and CPU limits** of your PostgreSQL service, and whether it was ever stopped for running out of memory. Used for `shared_buffers`, `effective_cache_size`, `work_mem`, `maintenance_work_mem`, `max_connections` and the number of parallel workers.
- **Your data**: the total size of your databases, so `shared_buffers` is only raised as far as your data needs.
- **Usage statistics** since the service started: how many reads came from memory, temporary files written to disk by sorts, and forced checkpoints. These are only used after the service has been running for at least an hour.
- **Connections**: how many of the allowed connections are in use, and connections left idle in a transaction.
- **Settings that cost performance**: logging every statement, `wal_level = logical` without replication slots, and autovacuum turned off.
- **Log**: recent *too many clients* and *checkpoints are occurring too frequently* messages.

Only settings that are listed on the page and exist on your PostgreSQL version are suggested.

:::info
Optimizations are general suggestions and may not result in increased database performance for all use cases.
:::

## Available Options

The list of configurable keys is determined by the system administrator. By default, the following are available:

- `max_connections`
- `shared_buffers`
- `work_mem`
- `maintenance_work_mem`
- `effective_cache_size`
- `max_worker_processes`
- `max_parallel_workers`
- `wal_level`
- `synchronous_commit`
- `checkpoint_timeout`
- `checkpoint_completion_target`
- `logging_collector`
- `log_directory`
- `log_filename`
- `log_min_duration_statement`
- `log_statement`
- `autovacuum`
- `autovacuum_max_workers`
- `autovacuum_naptime`
- `autovacuum_vacuum_scale_factor`

**Customizing Available Options**

Administrators can change which keys are exposed here by editing `/etc/openpanel/postgres/keys.txt` (one key per line). If that file doesn't exist, the default list above is used.
