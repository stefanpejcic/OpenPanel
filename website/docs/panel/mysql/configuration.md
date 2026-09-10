---
sidebar_position: 10
---

# Configuration

The MySQL Configuration page lets you edit low-level `my.cnf` settings for your database service, without needing terminal access.

## How to Change Configuration

1. Open **OpenPanel** and navigate to **MySQL > Configuration**.
2. Modify the values you want to change.
3. Click **Save** to apply changes.

Settings are written to your account's `custom.cnf` file, and the MySQL/MariaDB container is restarted to apply them.

## Available Options

The list of configurable keys is determined by the system administrator. By default, the following are available:

- `max_allowed_packet`
- `max_connect_errors`
- `max_connections`
- `open_files_limit`
- `performance_schema`
- `sql_mode`
- `thread_cache_size`
- `interactive_timeout`
- `wait_timeout`
- `log_output`
- `log_error`
- `log_error_verbosity`
- `general_log`
- `general_log_file`
- `long_query_time`
- `slow_query_log`
- `slow_query_log_file`
- `join_buffer_size`
- `key_buffer_size`
- `read_buffer_size`
- `read_rnd_buffer_size`
- `sort_buffer_size`
- `innodb_log_buffer_size`
- `innodb_log_file_size`
- `innodb_sort_buffer_size`
- `innodb_buffer_pool_chunk_size`
- `innodb_buffer_pool_instances`
- `innodb_buffer_pool_size`
- `max_heap_table_size`
- `tmp_table_size`

**Customizing Available Options**

Administrators can change which keys are exposed here by editing `/etc/openpanel/mysql/keys.txt` (one key per line). If that file doesn't exist, the default list above is used.
