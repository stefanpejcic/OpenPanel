---
sidebar_label: "OpenAdmin System Crons"
---

# System Cron Jobs Explained

Cron jobs used by OpenPanel are configured in the **`/etc/cron.d/openpanel`** file and should **not be edited manually**. Schedules and logging can be changed from [OpenAdmin > Server > Scheduled Actions](/docs/admin/advanced/crons), and the whole file is replaced with the default one when running [`opencli update --cron`](/docs/articles/opencli/update/).

Each job logs a line to `/var/log/openpanel/admin/cron.log` when it runs.

If you need to customize or add system cron jobs, use the **root user's crontab**.


## docker-collect_stats --all
Command `opencli docker-collect_stats --all` collects the container stats for all users. This data is visible to end-users from the *Resource Usage* page, and to Administrators from the *Users > single user* page.

- **Default schedule:** hourly (`0 * * * *`)
- **Documentation:** [opencli docker-collect_stats](/docs/articles/opencli/docker/#collect-stats)


## user-quota
Command `opencli user-quota` generates the disk and inodes usage report for all users and saves it to `/etc/openpanel/openpanel/quota_report.json`.

- **Default schedule:** every 5 minutes (`*/5 * * * *`)
- **Documentation:** [opencli user-quota](/docs/articles/opencli/user/#quota)


## domains-stats
Command `opencli domains-stats` parses (Caddy) access logs for all active users and their domains.

- **Default schedule:** daily at 02:30 (`30 2 * * *`)
- **Documentation:** [opencli domains-stats](/docs/articles/opencli/domains/#stats)


## websites-pagespeed --all
Command `opencli websites-pagespeed --all` collects **Google PageSpeed** data for all websites on the server. This data is visible to end-users from Site Manager and WP Manager pages.

If a user has configured PageSpeed API keys, those will be used automatically for all their domains.

- **Default schedule:** daily at 04:00 (`0 4 * * *`)
- **Documentation:** [opencli websites-pagespeed](/docs/articles/opencli/websites/#pagespeed)


## websites-vulnerability --all
Command `opencli websites-vulnerability --all` scans all WordPress websites on the server for known vulnerabilities in their plugins and themes.

- **Default schedule:** at 05:00 every second day (`0 5 */2 * *`)
- **Documentation:** [opencli websites-vulnerability](/docs/articles/opencli/websites/#vulnerability)


## files-calculate_resellers_storage
Command `opencli files-calculate_resellers_storage` calculates the total disk usage of each reseller's users and saves it to the reseller's limits file `/etc/openpanel/openadmin/resellers/<reseller>.json`.

- **Default schedule:** every 4 hours (`0 */4 * * *`)
- **Documentation:** [opencli files-calculate_resellers_storage](/docs/articles/opencli/files/#resellers-storage)


## update
Command `opencli update` checks for OpenPanel updates and performs them if `autopatch` or `autoupdate` are enabled.

- **Default schedule:** daily at 00:15 (`15 0 * * *`)
- **Documentation:** [opencli update](/docs/articles/opencli/update/)


## ftp-users
Command `opencli ftp-users` recreates the FTP users file that displays users in OpenAdmin interface.

- **Default schedule:** every 8 hours (`0 */8 * * *`)
- **Documentation:** [opencli ftp-users](/docs/articles/opencli/ftp/#users)


## email-server pflogsumm
Command `opencli email-server pflogsumm` generates an **HTML email summary report**, visible in **OpenAdmin > Emails > Summary Reports**.

- **Default schedule:** daily at 01:45 (`45 1 * * *`)
- **Documentation:** [opencli email-server pflogsumm](/docs/articles/opencli/email/#pflogsumm)


## files-purge_trash
Command `opencli files-purge_trash` purges **FileManager Trash folders** for all users, based on the `autopurge_trash` setting retention.

- **Default schedule:** daily at 04:45 (`45 4 * * *`)
- **Documentation:** [opencli files-purge_trash](/docs/articles/opencli/files/#purge-trash)


## waf update
Command `opencli waf update` updates the default **OWASP CRS WAF rules**. Update history can be viewed with: `opencli waf update log`.

- **Default schedule:** monthly on the 1st at 00:00 (`0 0 1 * *`)
- **Documentation:** [opencli waf update](/docs/articles/opencli/waf/#update)


## backup
Command `opencli backup` backs up the OpenPanel and OpenAdmin system configuration. User data (website files, databases, mailboxes) is not included. Runs can be viewed from **OpenAdmin > Backups > System > Runs**.

- **Default schedule:** weekly on Sunday at 01:00 (`0 1 * * 0`)
- **Documentation:** [opencli backup](/docs/articles/opencli/backup/)


## files-malware_scan --all
Command `opencli files-malware_scan --all` scans website files of all users with ClamAV and quarantines any infected file found. Requires ClamAV to be enabled on the server.

- **Default schedule:** weekly on Sunday at 00:00 (`0 0 * * 0`)
- **Documentation:** [opencli files-malware_scan](/docs/articles/opencli/files/#malware-scan)


## docker-backup
Command `opencli docker-backup` executes a **backup for all users' containers**. This is disabled by default, as each end-user can then [schedule and manage their own backups from OpenPanel UI](/docs/panel/backups/).

- **Default schedule:** `59 23 31 2 *` (never runs)
- **Documentation:** [opencli docker-backup](/docs/articles/opencli/docker/#backup), [Configuring OpenPanel Backups: Admin-configured](/docs/articles/backups/configuring-backups/#1-admin-configured)


## sentinel
Command `opencli sentinel` checks services, resource usage, traffic, logins and DNS, and sends alerts to OpenAdmin > Notifications (and email/webhook). Notifications for issues that are resolved on a later run are marked as read.

- **Default schedule:** every 5 minutes (`*/5 * * * *`)
- **Documentation:** [opencli sentinel](/docs/articles/opencli/sentinel/)


## sentinel --report
Command `opencli sentinel --report` sends the [**daily usage report**](/docs/admin/settings/notifications) to the Administrator email address.

- **Default schedule:** daily at 11:45 (`45 11 * * *`)
- **Documentation:** [opencli sentinel](/docs/articles/opencli/sentinel/)


## sentinel --startup
Command `opencli sentinel --startup` runs after a server reboot: it prepares the Redis temporary directory, starts stopped containers for root and all users, and sends the reboot notification.

- **Default schedule:** `@reboot`
- **Documentation:** [opencli sentinel](/docs/articles/opencli/sentinel/)
