---
sidebar_position: 4
---

# Cron jobs

**OpenAdmin > Advanced > System Cron Jobs** lets Administrators view the scheduled cron tasks used by OpenPanel, modify their schedules, or enable/disable logging to the file `/etc/openpanel/openadmin/cron.log`.

![screenshot](/img/admin/openadmin_cronjobs.png)


Changing schedules may cause some OpenPanel features to stop working correctly. Only adjust these settings if:

- You need to fine-tune execution on servers with limited resources, or
- You have been advised to do so by OpenPanel support.

These cron jobs are essential for OpenPanel’s internal operations.

If you want to add your own custom cron jobs, it’s safer to add them directly in the root user’s crontab.
