---
sidebar_position: 7
---

# Activity Log

The Account Activity page provides a log record of all your actions performed on the OpenPanel, along with the timestamp and IP address from which the action was executed. The primary objective of this page is to offer insights into who carried out specific actions, such as deleting a file, adding domains, resetting WordPress admin passwords, and more.

![account_activity_log.png](/img/panel/v2/activity.png)


## Recorded actions:

The OpenPanel interface records the following account activities:

* Added item to favorites
* Removed item from favorites
* Password changed
* Email address changed
* Username changed
* Forgot password requested (via email)
* Password reset (via email)
* Logged in with password
* Logged in using 2FA code
* Logged in via API call
* Logged out
* Changed notification preferences
* Terminated active session
* Enabled or disabled Two-Factor Authentication (2FA)
* Enabled or disabled **ElasticSearch**
* Enabled or disabled **Memcached**
* Enabled or disabled **OpenSearch**
* Enabled or disabled **Redis**
* Enabled or disabled **Varnish**
* Enabled or disabled **Varnish caching** for a domain
* Added new DNS record for a domain
* Updated DNS record for a domain
* Deleted DNS record for a domain
* Reset DNS zone for a domain
* Exported DNS zone for a domain
* Added, edited, or deleted a custom Docker service
* Switched MySQL type
* Switched webserver type
* Changed image tag for Docker service
* Changed CPU or Memory limit for Docker service
* Enabled or disabled Docker service
* Force-pulled image for Docker service
* Executed command via web terminal
* Edited VirtualHosts file for a domain
* Created or deleted redirect link for a domain
* Added domain
* Deleted domain
* Suspended or unsuspended domain
* Enabled or disabled WAF for a domain
* Updated WAF rules for a domain
* Added email address
* Modified email address (password, quota, suspend/unsuspend inbound or outbound traffic)
* Downloaded email configuration for Outlook/Thunderbird
* Deleted email address
* Accessed webmail for an email
* Edited sieve filter for an email
* Created file or folder
* Uploaded files
* Downloaded files from URL
* Created archive
* Extracted archive
* Renamed file or folder
* Deleted files to Trash
* Restored files from Trash
* Emptied Trash
* Permanently deleted files
* Changed file/folder permissions
* Copied or moved files/folders
* Edited file using File Editor
* Created or deleted FTP account
* Changed FTP account password
* Changed FTP account path
* Downloaded FTP configuration (FileZilla/Cyberduck)
* Initiated ClamAV scan for folder
* Created MySQL database
* Created MySQL database user
* Assigned or revoked all privileges for a user on a MySQL database
* Changed MySQL root user password
* Deleted MySQL database or user
* Changed password for MySQL database user
* Edited MySQL configuration
* Enabled or disabled remote MySQL access
* Created PostgreSQL database
* Created PostgreSQL database user
* Assigned or revoked all privileges for a user on a PostgreSQL database
* Deleted PostgreSQL database or user
* Changed password for PostgreSQL database user
* Edited PostgreSQL configuration
* Enabled or disabled remote PostgreSQL access
* Enabled or disabled **pgAdmin**
* Edited version, password, or email for **pgAdmin**
* Enabled or disabled **phpMyAdmin**
* Edited Max Execution Time, Memory Limit, Upload Limit, Version, or Absolute URI for **phpMyAdmin**
* Changed default PHP version for new domains
* Changed PHP version for a domain
* Edited limits for PHP versions
* Edited PHP configuration via PHP Selector
* Created new Python / NodeJS application
* Started, stopped, or restarted application
* Edited application
* Deleted application
* Executed `pip`, `npm`, or `pnpm install` command
* Installed, detached, or uninstalled WordPress
* Restored WordPress files or database from a backup
* Generated full/files/databases WordPress backup
* Initiated scan for WordPress installations
* Generated auto-login link for wp-admin
* Flushed WP cache
* Executed `wp-cli` commands
* Initiated PageSpeed data refresh
* Updated WordPress debug options
* Updated WordPress site information
* Edited WordPress auto-update preferences
* Installed, detached, or uninstalled Website Builder
* Terminated process using Process Manager
* Enabled or disabled a service
* Refreshed resource usage for the account


You can search for these actions by IP, username, date, or specific action.

By default only 25 actions will be shown per page, you can navigate pages using the pagination links or click on 'Show all activity' to display all actions.
