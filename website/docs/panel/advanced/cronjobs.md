---
sidebar_position: 1
---

# Cron Jobs

A cron job is a Linux command used to schedule tasks for future execution. It allows you to automate repetitive tasks, such as sending notifications or running scripts at specific intervals.

![cronjobs.png](/img/panel/v2/cronjobsmain.png)

On the CronJobs page you can view currently scheduled tasks, create new, edit or delete them.

:::info
The TimeZone setting is handy for running scheduled [cronjobs](/docs/panel/advanced/cronjobs) in your local timezone.
:::


## Add

To create a new cronjob click on the 'Create CronJob' button and in the new form set the script to be executed, choose a container to execute the script and the desired schedule.

![cronjobs_new.png](/img/panel/v2/cronjobs.png)

The first field allows you to choose the container which is going to be running the script.

![cronjobs_container.png](/img/panel/v2/cronjobs_container.png)

The second field allows you to set a predefined (common) schedule:

- Every 30s
- Every minute
- Every 5 minutes
- Every 30 minutes
- Hourly
- Daily
- Weekly
- Monthly
- Yearly

![cronjobs_new_predefined.png](/img/panel/v2/cronjobs_common.png)

you can also set a standard cron expression representing set of times, using **6** space-separated fields:

| Field name   | Mandatory? | Allowed values      | Allowed special characters |
| ------------ | ---------- | ----------------- | ------------------------- |
| Seconds      | Yes        | 0-59               | `*` `/` `,` `-`           |
| Minutes      | Yes        | 0-59               | `*` `/` `,` `-`           |
| Hours        | Yes        | 0-23               | `*` `/` `,` `-`           |
| Day of month | Yes        | 1-31               | `*` `/` `,` `-` `?`       |
| Month        | Yes        | 1-12 or JAN-DEC    | `*` `/` `,` `-`           |
| Day of week  | Yes        | 0-6 or SUN-SAT     | `*` `/` `,` `-` `?`       |

:::warning
There are 6 fields instead of the usual 5 found in standard Unix cron. This is because OpenPanel cron jobs also support scheduling by seconds.
:::

For more information, check [CRON_Expression_Format](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format)

## Edit

To edit an existing cronjob, click on the 'Edit' button next to it. This action will allow you to edit that specific cron job.

![cronjobs_edit.gif](/img/panel/v2/cron_edit_v2.gif)

To modify the schedule for when the script is executed you can use a tool such as https://crontab.guru/.

When you're done click on the 'Save' button to update the crontab file with your changes.

## Delete

To delete a cronjob, click on the 'Delete' button next to it. This action will open a modal asking you to confirm the deletion.

![cronjobs_delete.gif](/img/panel/v2/cron_delete.gif)

## Logs

Each cron job execution is recorded in JSON format.

To view the logs, click the *“View Logs”* button. This will open a new tab displaying all cron job logs.

To filter logs by a specific job name (comment), append the following parameter to the URL:
`?job=` followed by the job name. Example: `/cronjobs/log?job=whmcs-cron`

## File Editor

The *File Editor* option allows you to edit the file where crons are stored. File format is:

```
[job-exec "JOB_NAME"]
schedule = 
container = 
command =
```

Example:
```
[job-exec "example"]
schedule = @daily
container = nginx
command = curl https://ip.openpanel.com
```

Another Example:
```
[job-exec "whmcs-cron"]
schedule = @every 5m
container = php-fpm-8.4
command = php -q /var/www/html/whmcs-paths/crons/cron.php
```

:::info
You can also set `no-overlap` for a cronjob, to avoid running the job multiple times in parallel if the previous execution did not finish.
:::


## Import / Export

Cronjobs can be bulk-edited through a file, making it easy to edit multiple jobs at once or transfer them between servers.
Simply click the *“Switch to File Editor”* button to open the editor.

![file editor](https://i.postimg.cc/zXx0LDMm/slika.png)
