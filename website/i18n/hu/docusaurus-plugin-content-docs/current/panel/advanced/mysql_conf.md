---
sidebar_position: 7
---

# MySQL konfiguráció

A **MySQL konfiguráció szerkesztése** oldalon módosíthatja az adatbázis-szolgáltatás (MySQL vagy MariaDB) beállításait.

> **MEGJEGYZÉS**: Egy mező üresen hagyása azt jelenti, hogy a rendszer az adatbázis-verzió alapértelmezett értékét fogja használni.

A konfigurálható opciók a következők:

- max_allowed_packet
- max_connect_errors
- max_connections
- open_files_limit
- teljesítmény_séma
- sql_mode
- thread_cache_size
- interactive_timeout
- várakozás_időtúllépés
- log_output
- log_error
- log_error_verbosity
- általános_napló
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

A rendszergazdák teljes körűen szabályozhatják ezeket a beállításokat szükség szerint.

:::info
Megjegyzés: A módosítások alkalmazása újraindítja az adatbázis-szolgáltatást.
:::
