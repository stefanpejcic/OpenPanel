# OpenPanel --- Cron Troubleshooting Guide

This is a **generic troubleshooting guide** for OpenPanel cron jobs.

------------------------------------------------------------------------

## **1) How cron works in OpenPanel**

In OpenPanel:

-   Cron **does NOT run directly on the host** like traditional Linux
    crontab.
-   Each **user has their own cron container**.
-   Jobs run **inside a specified container** (e.g. `openlitespeed`,
    `nginx`, `php`, etc.).
-   Cron jobs are defined in **`/home/USERNAME/crons.ini`**, not in
    `/etc/crontab`.

------------------------------------------------------------------------

## **2) First diagnostic step --- run the command manually (is it even cron-related?)**

Before troubleshooting cron, always verify that the command itself
works.

Access the user container terminal:
https://openpanel.com/docs/panel/containers/terminal/#accessing-the-terminal

Inside the container, run your command manually, for example:

``` bash
php -q /var/www/html/example.com/apps/console/console.php example-task
```

### If this fails:

-   The issue is **NOT cron**
-   Fix the error first (missing PHP extensions, wrong path,
    permissions, app config, etc.)

Only continue with cron troubleshooting **after this works manually.**

------------------------------------------------------------------------

## **3) Check if the cron service is running**

Each user must have their own cron service active.

Check in OpenPanel: https://openpanel.com/docs/panel/advanced/services/

You should see a **cron service running for the user**.

### If it is NOT running, start it manually (terminal):

``` bash
cd /home/USERNAME && docker --context=USERNAME compose up -d cron
```


------------------------------------------------------------------------

## **4) Verify the cron file format (`crons.ini`)**

Check the user's cron file:

``` bash
cat /home/USERNAME/crons.ini
```

A correct job looks like this:

``` ini
[job-exec "example-job"]
schedule = @every 1m
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php example-task
```

### Common mistakes

❌ Using host PHP path:

``` bash
/usr/local/lsws/lsphp84/bin/php
```

✅ Use container PHP instead:

``` bash
php
```

❌ Wrong container:

``` ini
container = web
```

✅ Must match the actual container:

``` ini
container = openlitespeed
```

------------------------------------------------------------------------

## **5) Check cron execution logs (very important)**

OpenPanel provides per-job execution logs.

Cron job logs: https://openpanel.com/docs/panel/advanced/cronjobs/#logs

Look for: - Job started - Job failed - PHP errors - Permission denied -
"Container not found"

------------------------------------------------------------------------

## **6) Verify paths inside the container**

Paths inside Docker may differ from host paths.

Check inside the container:

``` bash
docker --context USERNAME exec -it CONTAINER bash
ls /var/www/html
```

Confirm that:

    /var/www/html/example.com/

actually exists inside the container.

------------------------------------------------------------------------

## **7) Inspect the cron container logs**

If jobs still don't run, inspect the cron container itself:

``` bash
docker --context USERNAME logs cron
```

Look for:
- Scheduler errors
- Container startup issues
- Permission problems

------------------------------------------------------------------------

## **8) Validate correct container selection**

Make sure the job uses the correct container:

  App type              Likely container
  --------------------- ------------------
  OpenLiteSpeed + PHP   `openlitespeed` or `php`
  Nginx + PHP       `nginx` or `php`
  Standalone container     `CONTAINER_NAME`

If you pick the wrong one → job fails.

------------------------------------------------------------------------

## **9) Generic working cron template**

For an app hosted at `example.com`:

``` ini
[job-exec "task-1"]
schedule = @every 1m
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php task-1

[job-exec "task-2"]
schedule = @every 5m
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php task-2

[job-exec "task-hourly"]
schedule = @hourly
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php hourly

[job-exec "task-daily"]
schedule = @daily
container = openlitespeed
command = php -q /var/www/html/example.com/apps/console/console.php daily
```

------------------------------------------------------------------------

## **10) Quick troubleshooting decision tree**

### ❓ "Cron doesn't run at all"

Check: - Is cron service running? → Services page\
- Is there a cron container? → `docker --context USERNAME  ps | grep cron`\
- Does `/home/USERNAME/crons.ini` exist?

### ❓ "Cron runs but command fails"

Check: - Run manually inside container first\
- Check cron logs in OpenPanel\
- Check container logs

### ❓ "Cron runs but the app still complains"

Check: - App logs inside `/var/www/html/example.com/.../logs/`\
- File permissions inside the container
