---
sidebar_position: 7
---

# MySQL Configuration

On the **Edit MySQL Configuration** page, you can modify settings for your database service (MySQL or MariaDB).

Leaving a field empty means the system will use the default value for your database version.

The configurable options include:

- max_allowed_packet	
- max_connect_errors	
- max_connections	
- open_files_limit	
- performance_schema	
- sql_mode	
- thread_cache_size	
- interactive_timeout	
- wait_timeout	
- log_output	
- log_error	
- log_error_verbosity	
- general_log	
- general_log_file	
- long_query_time	
- slow_query_log	
- slow_query_log_file	
- join_buffer_size	
- key_buffer_size	
- read_buffer_size	
- read_rnd_buffer_size	
- sort_buffer_size	
- innodb_log_buffer_size	
- innodb_log_file_size	
- innodb_sort_buffer_size	
- innodb_buffer_pool_chunk_size	
- innodb_buffer_pool_instances	
- innodb_buffer_pool_size	
- max_heap_table_size	
- tmp_table_size

Administrators have full control to adjust these settings as needed.

:::info
Please note: Applying changes will restart the database service.
:::
