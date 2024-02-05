---
sidebar_position: 4
---

# Process Manager

The process manager interface allows you to view all running processes and kill (force stop the process).

![process_manager.png](/img/panel/v1/advanced/process_manager.png)

Processes are sorted by their current CPU % usage.

Due to security considerations, certain processes, such as [phpMyAdmin](/docs/panel/databases/phpmyadmin) and [Web Terminal](/docs/panel/advanced/terminal), which involve sensitive data like passwords in their command line attributes, have been substituted with placeholders (xxxxxxxx) in the CMD column.

To terminate a process click on the 'kill' button next to it.

:::danger
**NOTE:** Stopping MySQL, PHP-FPM, or Nginx/Apache processes may result in downtime for your websites. In the event that a service fails to restart after terminating the process, you can utilize the [Service Status](/docs/panel/advanced/server_settings#service-status) page to restart the appropriate service.
:::
