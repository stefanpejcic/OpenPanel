---
sidebar_position: 1
---

# CronJobs

A cron job is a Linux command used to schedule tasks for future execution. It allows you to automate repetitive tasks, such as sending notifications or running scripts at specific intervals.

![cronjobs.png](/img/panel/v2/cronjobsmain.png)

On the CronJobs page you can view currently scheduled tasks, create new, edit or delete them.

:::info
The TimeZone setting is handy for running scheduled [cronjobs](/docs/panel/advanced/cronjobs) in your local timezone.
:::


## Add a CronJob

To create a new cronjob click on the 'Create CronJob' button and in the new form set the script to be executed, choose a container to execute the script and the desired schedule.

![cronjobs_new.png](/img/panel/v2/cronjobs.png)

The first field allows you to choose the container which is going to be running the script.

![cronjobs_container.png](/img/panel/v2/cronjobs_container.png)

The second field allows you to set a predefined (common) schedule:

- Once per minute
- Once per 5 minutes
- Twice per hour 
- Once per hour
- Twice per day
- Once per day
- Once per week
- Twice per month (every 1st and 15th of the month)
- Once per month
- Once per year

![cronjobs_new_predefined.png](/img/panel/v2/cronjobs_common.png)


## Edit a CronJob

To edit an existing cronjob, click on the 'Edit' button next to it. This action will allow you to edit that specific cron job.

![cronjobs_edit.gif](/img/panel/v2/cron_edit_v2.gif)

To modify the schedule for when the script is executed you can use a tool such as https://crontab.guru/.

When you're done click on the 'Save' button to update the crontab file with your changes.

## Delete a CronJob

To delete a cronjob, click on the 'Delete' button next to it. This action will open a modal asking you to confirm the deletion.

![cronjobs_delete.png](/img/panel/v1/advanced/cronjobs_delete.png)
