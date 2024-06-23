---
sidebar_position: 4
---

# Cron jobs

*OpenAdmin > Server >  OpenPanel Cron Jobs* allows Administrators to view scheduled crons for OpenAdmin and edit their schedule or enable/disable logging to `/etc/openpanel/openadmin/cron.log` file.

:::warning We recommend against editing the cronjobs!
Editing the schedule or commands is not recommended, as it may render certain features inaccessible. This should only be done in cases where fine-tuning execution on servers with low resources is necessary, or when instructed by the OpenPanel support team. These are crons that OpenPanel uses. If you want to add custom cronjobs, you can safely [add them in crontab for the root user](https://www.google.com/search?q=linux+add+cron).
:::

Cron format is:

```bash
SCHEDULE USER COMMAND
```
