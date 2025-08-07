---
sidebar_position: 3  
---

# PHP.INI Editor  

The PHP.INI Editor allows you to view and modify the configuration file for each installed PHP version.  

You can increase limits, enable new defaults, or adjust settings as needed. Changes will restart the corresponding PHP service.

![openpanel_php_ini_editor](/img/panel/v2/openpanel_php_ini_editor.gif)

> NOTE: THe following settings: `max_execution_time` `max_input_time` `max_input_vars` `memory_limit` `post_max_size` `upload_max_filesize` must be set via [**PHP Limits** page](/docs/panel/php/limits).

---

> Note: PHP.INI Editor is only available if `php_ini` feature is enabled by Administrator.

To modify the total CPU or Memory resource limits for PHP service (container) use the [**Containers** page](/docs/panel/containers/).
