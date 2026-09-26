---
sidebar_position: 11
---

# Cron Jobs

A cron job is a Linux command used to schedule tasks for future execution. It allows you to automate repetitive tasks, such as sending notifications or running scripts at specific intervals.

![Cron Jobs page listing scheduled jobs with their schedule, container, command and comment](/img/openpanel-screenshots/advanced/cronjobs-list.png#gh-light-mode-only)
![Cron Jobs page listing scheduled jobs with their schedule, container, command and comment](/img/openpanel-screenshots/advanced/cronjobs-list_dark.png#gh-dark-mode-only)

The Cron Jobs page has three tabs:

- **Cron Jobs**: the table of scheduled jobs, where you can create new jobs and edit, delete, run or view the logs of existing ones.
- **File Editor**: edit the file the jobs are stored in directly, see [File Editor](#file-editor).
- **Logs**: the output of your cron jobs, see [Logs](#logs).


## Add

To create a new cronjob click on the 'Create New' button and in the new form set the script to be executed, choose a container to execute the script and the desired schedule.

![Create Cron Job form with the container, schedule, common schedules, command and comment fields](/img/openpanel-screenshots/advanced/cronjobs_new-form.png#gh-light-mode-only)
![Create Cron Job form with the container, schedule, common schedules, command and comment fields](/img/openpanel-screenshots/advanced/cronjobs_new-form_dark.png#gh-dark-mode-only)

The first field allows you to choose the container which is going to be running the script.

![Select Container dropdown of the Create Cron Job form](/img/openpanel-screenshots/advanced/cronjobs_new-container.png#gh-light-mode-only)
![Select Container dropdown of the Create Cron Job form](/img/openpanel-screenshots/advanced/cronjobs_new-container_dark.png#gh-dark-mode-only)

The second field allows you to set a predefined (common) schedule:

- Every 30 Seconds
- Every Minute
- Every 5 Minutes
- Every 30 Minutes
- Hourly
- Daily
- Weekly
- Monthly
- Yearly

![Common schedules dropdown set to Hourly, which fills in @hourly as the schedule](/img/openpanel-screenshots/advanced/cronjobs_new-common.png#gh-light-mode-only)
![Common schedules dropdown set to Hourly, which fills in @hourly as the schedule](/img/openpanel-screenshots/advanced/cronjobs_new-common_dark.png#gh-dark-mode-only)

you can also set a standard cron expression representing set of times, using **6** space-separated fields:

| Field name   | Mandatory? | Allowed values      | Allowed special characters |
| ------------ | ---------- | ----------------- | ------------------------- |
| Seconds      | Yes        | 0-59               | `*` `/` `,` `-`           |
| Minutes      | Yes        | 0-59               | `*` `/` `,` `-`           |
| Hours        | Yes        | 0-23               | `*` `/` `,` `-`           |
| Day of month | Yes        | 1-31               | `*` `/` `,` `-` `?`       |
| Month        | Yes        | 1-12 or JAN-DEC    | `*` `/` `,` `-`           |
| Day of week  | Yes        | 0-6 or SUN-SAT     | `*` `/` `,` `-` `?`       |

:::info
There are 6 fields instead of the usual 5 found in standard Unix cron. This is because OpenPanel cron jobs also support scheduling by seconds.
:::

For more information, check [CRON_Expression_Format](https://pkg.go.dev/github.com/robfig/cron#hdr-CRON_Expression_Format)

## Edit

To edit an existing cronjob, click on the 'Edit' button next to it. This action will allow you to edit that specific cron job.

![A cron job row in edit mode with editable schedule, container, command and comment fields](/img/openpanel-screenshots/advanced/cronjobs-edit.png#gh-light-mode-only)
![A cron job row in edit mode with editable schedule, container, command and comment fields](/img/openpanel-screenshots/advanced/cronjobs-edit_dark.png#gh-dark-mode-only)

To modify the schedule for when the script is executed you can use a tool such as https://crontab.guru/.

When you're done click on the 'Save' button to update the crons file with your changes.

## Delete

To delete a cronjob, click on the 'Delete' button next to it. The button then switches to a 'Confirm' state with a 5 second countdown - click it again within that window to actually remove the job. If you don't click again, it reverts back to 'Delete' and nothing is removed.

![Delete button of a cron job turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/advanced/cronjobs-delete.png#gh-light-mode-only)
![Delete button of a cron job turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/advanced/cronjobs-delete_dark.png#gh-dark-mode-only)

## Run Now

To test a cron job without waiting for its schedule, click the 'Run' button next to it. This opens a modal that executes the job's command inside its configured container right away and streams the output live as it runs, followed by the exit code once it finishes.

![Run now dialog of a cron job showing the command output and Finished successfully](/img/openpanel-screenshots/advanced/cronjobs-run.png#gh-light-mode-only)
![Run now dialog of a cron job showing the command output and Finished successfully](/img/openpanel-screenshots/advanced/cronjobs-run_dark.png#gh-dark-mode-only)

This is useful for quickly checking that a job is configured correctly before relying on its schedule. Closing the modal disconnects the stream and stops the command if it's still running.

## Logs

Each cron job execution is recorded in JSON format.

![Logs tab of the Cron Jobs page with the Job and Lines filters, Refresh button and log entries of recent runs](/img/openpanel-screenshots/advanced/cronjobs-logs.png#gh-light-mode-only)
![Logs tab of the Cron Jobs page with the Job and Lines filters, Refresh button and log entries of recent runs](/img/openpanel-screenshots/advanced/cronjobs-logs_dark.png#gh-dark-mode-only)

To view the logs, open the **Logs** tab, or click the **Logs** button in a job's row to open it already filtered to that job. You can filter by job name, choose how many lines to display and refresh the results. A job's logs can also be linked directly, for example `/cronjobs/logs?job=whmcs-cron`.

To filter logs by a specific job name (comment) directly through the API endpoint, append the following parameter to the URL:
`?job=` followed by the job name. Example: `/cronjobs/log?job=whmcs-cron`

## File Editor

The **File Editor** tab allows you to edit the file where crons are stored. File format is:

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
Open the **File Editor** tab, edit the jobs and click **Save Changes**.

![Cron jobs File Editor showing the jobs in crons.ini format](/img/openpanel-screenshots/advanced/cronjobs_editor-editor.png#gh-light-mode-only)
![Cron jobs File Editor showing the jobs in crons.ini format](/img/openpanel-screenshots/advanced/cronjobs_editor-editor_dark.png#gh-dark-mode-only)

See also: [Restart a service automatically with a cron job](/docs/articles/containers/restart-service-with-cron/)

## Bulk Actions

Tick the checkbox of one or more cron jobs, or the checkbox in the table header to select every cron job shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![Two cron jobs selected with the bulk actions bar offering Run now, Change schedule, Change container, No overlap on, No overlap off and Delete](/img/openpanel-screenshots/advanced/cronjobs-bulk.png#gh-light-mode-only)
![Two cron jobs selected with the bulk actions bar offering Run now, Change schedule, Change container, No overlap on, No overlap off and Delete](/img/openpanel-screenshots/advanced/cronjobs-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Run now** | Runs the selected jobs one after another and reports each exit code. |
| **Change schedule** | Sets the schedule you enter, with 6 fields (seconds first) or a descriptor like `@hourly`. |
| **Change container** | Runs the selected jobs in the container you pick. |
| **No overlap on** | Skips a run while the previous run of the same job is still going. |
| **No overlap off** | Lets a job run even while its previous run is still going. |
| **Delete** | Deletes the selected jobs. |

Click an action, confirm it, and it runs on the selected cron jobs one after another. When it's done the page reloads with a notice listing the cron jobs it worked for, or which ones failed and why.
