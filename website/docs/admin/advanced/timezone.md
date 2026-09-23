---
sidebar_position: 4
---

# Server Time

Use the Server Time section to update the system-wide timezone for both the host operating system and the OpenPanel user interface.

Changing the timezone ensures that logs, scheduled tasks, and UI timestamps align with your desired regional time settings.

![Change TimeZone page with the current timezone and the timezone dropdown](/img/openadmin-screenshots/advanced/timezone-page.png)

:::info
After changing timezone, we recommend to restart the server or at least the OpenAdmin and Cron services: `systemctl restart admin cron`.
:::
