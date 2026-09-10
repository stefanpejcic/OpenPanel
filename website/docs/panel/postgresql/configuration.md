---
sidebar_position: 12
---

# Configuration

The PostgreSQL Configuration page lets you edit low-level `postgresql.conf` settings for your database service, without needing terminal access.

## How to Change Configuration

1. Open **OpenPanel** and navigate to **PostgreSQL > Configuration**.
2. Modify the values you want to change.
3. Click **Save** to apply changes.

Settings are written to a configuration override for your account, and the PostgreSQL container is restarted to apply them.

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
