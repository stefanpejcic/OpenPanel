# OpenAdmin System Crons

Cron jobs used by OpenPanel are configured in **`/etc/cron.d/openpanel`** file and should **not be edited manually**, as they will be overwritten during updates.  

If you need to customize or add system cron jobs, use the **root user's crontab**.  


## docker-collect_stats --all
Command `opencli docker-collect_stats --all` collects the docker stats for all users. THis data is visible by end-users From 'Resurce Usage' pag, and for Amdinistrators form *Users > single user*  page.

Default schedule: hourly (`0 * * * *`)

Documentation: https://dev.openpanel.com/cli/docker.html#Collect-Stats




## domains-stats
Command `opencli domains-stats` parses (Caddy) access logs for all active users and their domains/.

- **Default schedule:** daily at 02:30 (`30 2 * * *`)  
- **Documentation:** [opencli domains-stats](https://dev.openpanel.com/cli/domains.html#Parse-domain-access-logs)



## websites-pagespeed --all
Command `opencli websites-pagespeed --all` collects **Google PageSpeed** data for all websites on the server. This data is visible by end-users from Site Manager and WP Manager pages.

If user has configured pagespeed api keys, those will be used automatically for all their domains.

- **Default schedule:** daily at 04:00 (`0 4 * * *`)  
- **Documentation:** [opencli websites-pagespeed](https://dev.openpanel.com/cli/websites.html#PageSpeed)


## update
Command `opencli update` checks for OpenPanel updates and performs them if `autopatch` or `autoupdate` are enabled.

- **Default schedule:** daily at 00:15 (`15 0 * * *`)  
- **Documentation:** [opencli update](https://dev.openpanel.com/cli/update.html)



## server-ips
Command `opencli server-ips` generates a fresh list of the server’s IP addresses.

- **Default schedule:** monthly on the 12th at 00:00 (`0 0 12 * *`)  
- **Documentation:** This command is DEPRECATED.


## ftp-users
Command `opencli ftp-users` recreates the FTP users file that displays users in OpenAdmin interface.

- **Default schedule:** every 8 hours (`0 */8 * * *`)  
- **Documentation:** [opencli ftp-users](https://dev.openpanel.com/cli/ftp.html#Users)



## email-server pflogsumm
Command `opencli email-server pflogsumm` generates an **HTML email summary report**, visible in **OpenAdmin > Emails > Summary Reports**.

- **Default schedule:** daily at 01:45 (`45 1 * * *`)  
- **Documentation:** [opencli email-server pflogsumm](https://dev.openpanel.com/cli/email.html#pflogsumm)



## docker-images --all
Command `opencli docker-images --all` checks all active user's docker images for updates using [Cup 🥤](https://github.com/sergi0g/cup) and displays the data to end-users on *Containers > Image Updates* page.

- **Default schedule:** daily at 03:45 (`45 3 * * *`)  
- **Documentation:** [opencli docker-images](https://dev.openpanel.com/cli/docker.html#Images)



## files-purge_trash
Command `opencli files-purge_trash` purges **FileManager Trash folders** for all users, based on the `autopurge_trash` setting retention.

- **Default schedule:** daily at 04:45 (`45 4 * * *`)  
- **Documentation:** [opencli files-purge_trash](https://dev.openpanel.com/cli/files.html#Purge-Trash)



## waf-update
Command `opencli waf-update` updates the default **OWASP CRS WAF rules**. Logs can be viewed via: `opencli waf-update log`.

- **Default schedule:** monthly on the 1st at 00:00 (`0 0 1 * *`)  
- **Documentation:** [opencli waf-update](https://dev.openpanel.com/cli/waf.html#Update)



## docker-backup
Command `opencli docker-backup` executes a **backup for all users’ Docker containers** This is disabled by default, as each end-user can then [schedule and manage their own backups from OpenPanel UI](/docs/panel/files/backups/).

- **Default schedule:** `59 23 31 2 *`  
- **Documentation:** [Configuring OpenPanel Backups: Admin-configured](/docs/articles/backups/comfiguring-backups/#1-admin-configured)



## sentinel
Command `opencli sentinel` runs the notification script for server monitoring.

- **Default schedule:** every 5 minutes (`*/5 * * * *`)  
- **Documentation:** [opencli sentinel](https://dev.openpanel.com/cli/sentinel.html)



## sentinel --report
Command `opencli sentinel --report` generates the [**daily usage report**](/docs/admin/settings/notifications).

- **Default schedule:** daily at 11:45 (`45 11 * * *`)
- **Documentation:** [opencli sentinel report](/docs/admin/settings/notifications)



## sentinel --startup
Command `opencli sentinel --startup` sends notification after server reboot.

- **Command:** `opencli sentinel --startup`
- **Default schedule:** `@reboot`



## Redis Temp Directory
At reboot, this cron ensures the Redis temporary directory exists with proper permissions.

- **Command:** `mkdir -p /tmp/redis && chmod 777 /tmp/redis`
- **Default schedule:** `@reboot`
