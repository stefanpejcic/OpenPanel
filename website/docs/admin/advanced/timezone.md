---
sidebar_position: 4
---

# Server Time

Use the Server Time section to update the system-wide timezone for both the host operating system and the OpenPanel user interface.

Changing the timezone ensures that logs, scheduled tasks, and UI timestamps align with your desired regional time settings.

After applying a new timezone, it is recommended to restart the server, or at minimum, restart the following services to apply changes correctly:

```systemctl restart admin cron```
