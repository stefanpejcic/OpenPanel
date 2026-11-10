---
sidebar_position: 1
---

# Server time

Use the **System > Server time** page to update the system-wide timezone for both the host operating system and the OpenPanel user interface.

Changing the timezone ensures that logs, scheduled tasks, and UI timestamps align with your desired regional time settings.

The page shows the current timezone. To change it, type a city or region in **Find a timezone** (for example *Tbilisi*) to narrow down the **Timezone** list, pick one and click **Change timezone**. A message confirms the change, and errors are shown on the page.

![Server time page with the current timezone, the timezone search and the timezone list](/img/openadmin-screenshots/advanced/timezone-page.png#gh-light-mode-only)
![Server time page with the current timezone, the timezone search and the timezone list](/img/openadmin-screenshots/advanced/timezone-page_dark.png#gh-dark-mode-only)

:::info
After changing timezone, we recommend to restart the server or at least the OpenAdmin and Cron services: `systemctl restart admin cron`.
:::
