---
sidebar_position: 8
---

# PostgreSQL Configuration

On the **PostgreSQL > Configuration** page, you can modify settings for your PostgreSQL database service.

> **NOTE**: Leaving a field empty means the system will use the default value for your PostgreSQL version.

The configurable options include:

- max_connections
- shared_buffers
- work_mem
- maintenance_work_mem
- effective_cache_size
- max_worker_processes
- max_parallel_workers
- wal_level
- synchronous_commit
- checkpoint_timeout
- checkpoint_completion_target
- logging_collector
- log_directory
- log_filename
- log_min_duration_statement
- log_statement
- autovacuum
- autovacuum_max_workers
- autovacuum_naptime
- autovacuum_vacuum_scale_factor

Administrators have full control to adjust these settings as needed.

:::info
Please note: Applying changes will restart the database service.
:::
