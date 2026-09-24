---
sidebar_label: "Restart a Service with a Cron Job"
description: "How to automatically restart a service (web server, PHP, MySQL, Redis and others) on a schedule in OpenPanel, using a cron job that stops the container's main process."
---

# How to Restart a Service Automatically with a Cron Job

Sometimes you want a service restarted on a schedule - for example PHP-FPM every night, or a Node.js app that slowly leaks memory. In OpenPanel you can do this with a regular **cron job**, no terminal access needed.

---

## How It Works

Every service in your OpenPanel account (web server, PHP, MySQL, Redis, your apps, ...) runs in its own container with a **restart policy**. When the container's main process stops, the container is started again automatically.

The main process of a container always has the process ID **1**. So to restart a service, a cron job only needs to stop PID 1 inside that container:

```bash
kill 1
```

The service shuts down and comes straight back up a moment later.

---

## Create the Cron Job

1. Go to **OpenPanel → Cron Jobs** and click **Create New**.
2. **Container**: select the service you want to restart, e.g. `php-fpm-8.3`, `nginx`, `mysql` or your app's container.
3. **Schedule**: choose when it should run, e.g. **Daily**, or a cron expression like `0 0 4 * * *` for every day at 04:00 (OpenPanel cron schedules have 6 fields - the first one is seconds).
4. **Command**:

```bash
kill 1
```

5. Save.

![Create Cron Job form with the container, schedule, common schedules, command and comment fields](/img/openpanel-screenshots/advanced/cronjobs_new-form.png#gh-light-mode-only)
![Create Cron Job form with the container, schedule, common schedules, command and comment fields](/img/openpanel-screenshots/advanced/cronjobs_new-form_dark.png#gh-dark-mode-only)

Use the **Run** button next to the job to test it right away - the service should restart within a few seconds.

See [Cron Jobs](/docs/panel/cronjobs/) for all options.

---

## `kill 1` or `kill -9 1`?

| Command | What it does | When to use |
|---|---|---|
| `kill 1` | Asks the service to shut down cleanly (SIGTERM) | Default - lets databases and apps finish what they're doing |
| `kill -9 1` | Stops the process immediately (SIGKILL) | Only if the service is stuck and ignores `kill 1` |

Prefer `kill 1`, especially for **MySQL/MariaDB** and **PostgreSQL** - a forced stop can interrupt writes.

---

## Tips

- Schedule restarts at a **quiet time**, such as early morning - visitors get an error for the few seconds the service is down.
- Restarting a **web server** or **PHP** container briefly takes all sites that use it offline, not just one site.
- A service you stopped manually from the **Services** page stays stopped - the restart policy only brings back services that stopped on their own.
- If a service needs regular restarts because it runs out of memory, consider raising its limits instead - see [Services](/docs/panel/processes/services/).

---

## Related

- [Cron Jobs](/docs/panel/cronjobs/)
- [Cron troubleshooting guide](/docs/articles/websites/cron-troubleshooting-guide/)
- [How to troubleshoot services](/docs/articles/containers/how-to-troubleshoot-services-with-openpanel/)
