---
sidebar_position: 4
---

# Modules

Modules extend the OpenPanel UI by adding new features and pages. To make a feature available to a user or plan, it must first be activated as a module.

![openadmin_modules_settings](/img/admin/2.0/openadmin_modules_settings.png)

- Modules are **core features** that are already available on installation and are developed by OpenPanel.
- Plugins are custom features that need to be installed and are developed by third-party developers.

Available Modules:

## Notifications

The **`notifications`** module is required to send email notifications to users.

When enabled:
* Emails are sent according to each user’s notification preferences.
* Users can manage their preferences through the OpenPanel UI at: [**Accounts > Email Notifications**](/docs/panel/account/notifications/).

When disabled:
* No emails will be sent, regardless of user preferences.

Customize email notifications:
* To **set default preferences for new users** edit the [`/etc/openpanel/skeleton/notifications.yaml`](https://github.com/stefanpejcic/openpanel-configuration/blob/main/skeleton/notifications.yaml) file.
* To **customize email templates** refer to [Customizing OpenPanel Email Templates](https://community.openpanel.org/d/214-customizing-openpanel-email-templates).
* To **configure custom SMTP** use [OpenAdmin > Settings > Notifications page](/docs/admin/settings/notifications/).



## Account

The **`account`** module is required for users to change their email, password or username.

When enabled:
* Users can change their email, password and username through the OpenPanel UI at: [**Accounts > Settings**](/docs/panel/account/).

When disabled:
* Users can not change their passwords from OpenPanel UI, only from 'Password Reset' on login form, if this option is enabled.

Customize password and username changes:
* To **enable or disable password reset on login forms** edit 'Enable password reset on login' setting from [OpenAdmin > Settings > OpenPanel](/docs/admin/settings/openpanel/).
* To **prevent users from changing their username** edit 'Allow users to change username' setting from  [OpenAdmin > Settings > OpenPanel](/docs/admin/settings/openpanel/).


## Sessions

The **`sessions`** module allows users to view and manage their active sessions.

When enabled:
* Users can view all their active sessions, logs and terminate any session through the OpenPanel UI at: [**Accounts > Active Sessions**](/docs/panel/account/active_sessions/).

When disabled:
* Users can not access the *Accounts > Active Sessions* page.

Customize sessions duration:
* To **control session duration** edit 'Session duration' setting from [OpenAdmin > Settings > OpenPanel](/docs/admin/settings/openpanel/#Statistics).
* To **control session lifetime** edit 'Session lifetime' setting from [OpenAdmin > Settings > OpenPanel](/docs/admin/settings/openpanel/#Statistics).

## Locale

The **`locale`** (Languages) module allows users to change panel language.

When enabled:
* Users can change their preferred language for OpenPanel UI from the login page and [**Accounts > Change Language** page](/docs/panel/account/language/).

When disabled:
* Users can not access the *Accounts > Change Language* page to change their locale.
* Users are forced to the Admin defined default locale.

Customize locales:
* To **set the default locale** use [OpenAdmin > Settings > Locales](/docs/articles/accounts/default-user-locales/).
* To **install new locales for users** use the [OpenAdmin > Settings > Locales](/docs/admin/settings/locales/#install-locale).
* To **create a new translation** please see [How to Create a New Locale](/docs/admin/settings/locales/#edit-locale)


## Favorites

The **`favorites`** module allows users to *pin* items in their sidebar menu for quick navigation.

When enabled:
* Users can add pages to favorites with **left-click** on ⭐ icon in top-right corner of the page.
* Users can remove pages from favorites with **right-click** on ⭐ icon in top-right corner of the page.
* Users can access favorites from sidebar menu.
* Users can access the [**Accounts > Favorites** page](/docs/panel/account/favorites/).

When disabled:
* Users can not access the *Accounts > Favorites* page to manage favorites.
* Users are not see favorites in the sidebar nor the ⭐ icon in top-right corner of pages.

Customize favorites:
* To **control the total number of favorites for user** (default is 10) use [`favorites-items` config](https://dev.openpanel.com/cli/config.html#favorites-items).
* To **edit user's favorites from terminal** edit their: `/etc/openpanel/openpanel/core/users/{current_username}/favorites.json` file.


## Varnish

The **`varnish`** module allows users to control varnish caching for their domains.

When enabled:
* Varnish server starts for user and proxies traffic back to their webserver. 
* Users can access the [**Caching > Varnish** page](/docs/panel/caching/varnish/).
* Users can enable/disable Varnish service.
* Users can enable/disable Varnish caching per domain.
* Users can view logs for the Varnish service.

When disabled:
* Users do not have access to the *Caching > Varnish* page.
* Varnish is used only if Administrator enabled it for user when creating the account. 

Customize options:
* To **enable/disable Varnish for all new users** use [*OpenAdmin > Settings > User Defaults* page and *Enable Varnish Proxy* option](/docs/admin/settings/defaults/).
* To **enable/disable Varnish for a single user** when creating their account use the [**Enable Varnish Cache** option](/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/).
* To **change default CPU/RAM for service** use the [*OpenAdmin > Settings > User Defaults* page](/docs/admin/settings/defaults/).
* To **edit the default.vcl file for Varnish** use the [*OpenAdmin > Domains > Edit Domain Templates* page](/docs/admin/settings/defaults/) or edit file: [`/etc/openpanel/varnish/default.vcl`](https://github.com/stefanpejcic/openpanel-configuration/blob/main/varnish/default.vcl).
* To **purge Varnish cache** refer to [How-to Guides > Purging Varnish Cache](/docs/articles/websites/purge-varnish-cache-from-terminal/)
* To **check if Varnish is enabled for domain** refer to [How to check if Varnish Caching is enabled for a domain in OpenPanel?](https://community.openpanel.org/d/207-how-to-check-if-varnish-caching-is-enabled-for-a-domain-in-openpanel)


## Docker

The **`docker`** module allows users to manage and add new docker containers.

When enabled:
* Users can access [**Docker > Containers**](/docs/panel/containers/) page to view and manage services.
* Users can access [**Docker > Containers > New**](/docs/panel/containers/#adding-new-services) page to add new services.
* Users can access [**Docker > Image Updates**](/docs/panel/containers/image/) page to check for available image updates.
* Users can access [**Docker > Logs**](/docs/panel/containers/logs/) page to view service logs.

When disabled:
* Users can not access any of the *Docker* pages.

The web terminal, changing a service's image tag, and switching the webserver/MySQL type each live behind their own module below rather than `docker` - see **Web Terminal**, **Change Container Image Tag**, **Switch WebServer**, and **Switch MySQL Type**.


## Web Terminal

The **`terminal`** module allows users to run docker exec commands from an interactive shell inside their containers.

When enabled:
* Users can access [**Docker > Terminal**](/docs/panel/containers/terminal/) page to run docker exec commands.

When disabled:
* Users do not have access to the *Terminal* page.

Customize options:
* [Disable the Terminal](/docs/articles/dev-experience/disable-openpanel-web-terminal/)


## Change Container Image Tag

The **`change_image`** module allows users to change the image tag used by a container.

When enabled:
* Users can access [**Docker > Change Image Tag**](/docs/panel/containers/change/) page to change images tag.

When disabled:
* Users do not have access to the *Change Image Tag* page.


## Switch WebServer

The **`change_ws`** module allows users to switch the webserver used for their account.

When enabled:
* Users can access [**Docker > Switch Web Server**](/docs/panel/containers/webserver/) page to switch webservers.

When disabled:
* Users do not have access to the *Switch Web Server* page.


## Switch MySQL Type

The **`change_db`** module allows users to switch between MySQL and MariaDB for their account.

When enabled:
* Users can access [**Docker > Switch MySQL Type**](/docs/panel/containers/mysql/) page to switch mysql/mariadb.

When disabled:
* Users do not have access to the *Switch MySQL Type* page.


## Fix Permissions

The **`fix_permissions`** module allows users to reset file/folder permissions.

When enabled:
* Users can access the [**Files > Fix Permissions** page](/docs/panel/files/fix_permissions/).

When disabled:
* Users can not access the *Files > Fix Permissions* page.


## FTP

The **`ftp`** module allows users to create and manage FTP sub-accounts.

When enabled:
* Users can access the [**Files > FTP** page](/docs/panel/files/FTP/) to manage FTP accounts.

When disabled:
* Users can not create and manage FTP accounts.

Customize options:
* To **configure FTP server** refer to [*How-to Guides > Setup FTP](/docs/articles/user-experience/how-to-setup-ftp-in-openpanel/).
* To **edit VSFTPD configuration** edit the [`/etc/openpanel/ftp/vsftpd.conf` file](https://github.com/stefanpejcic/openpanel-configuration/blob/main/ftp/vsftpd.conf).
* To **view all ftp accounts on a server** use the [*OpenAdmin > Services > FTP* page](/docs/admin/services/ftp/).
* To **limit number of ftp accounts per user** edit the ftp accounts limit when creating/editing hosting packages.

## Emails

The **`emails`** module allows users to create and manage Email accounts.

When enabled:
* Users can access the [**Emails** pages](/docs/panel/emails/) to manage Email accounts.

When disabled:
* Users can not create and manage Email accounts.

Customize options:
* To **configure email server** refer to [*How-to Guides > Configure Email Server*](/docs/articles/user-experience/how-to-setup-email-in-openpanel/).
* To **configure email client** refer to [*How-to Guides > How to setup your email client*](/docs/articles/email/how-to-setup-your-email-client/).
* To **view all email accounts on a server** use the [*OpenAdmin > Emails > Email Accounts* page](/docs/admin/emails/).
* To **set up fail2ban** refer to [*How-to Guides > Setup Fail2ban](/docs/articles/email/how-to-setup-fail2ban-mailserver-openpanel/).
* To **set up Rspamd** refer to [*How-to Guides > RSPAMD GUI](/docs/articles/email/rspamd-gui-port-11334/).
* To **set up DKIM for a domain** refer to [*How-to Guides > Setup DKIM](/docs/articles/email/how-to-setup-dkim-for-mailserver/).
* To **limit number of email accounts per user** edit the email accounts limit when creating/editing hosting packages.


## Webmail

The **`webmail`** module allows users to auto-login to Roundcube webmail for their Email accounts.

When enabled:
* Users can access the [**Webmail**](/docs/panel/emails/webmail/) button.

When disabled:
* Users can not auto-login to webmail.

Customize options:
* To **set webmail domain or relay hosts** use the [*OpenAdmin > Emails > Email Settings* page](/docs/admin/emails/settings/).


## Email Filters

The **`email_filters`** module allows users to configure filters and forwarders for their emails.

When enabled:
* Users can access the [**Email Filters** page](/docs/panel/emails/filters/)

When disabled:
* Users can not access the Email Filters page.



## Email Aliases

The **`email_aliases`** module allows users to forwards mail from a non-existing address to one or more destinations.

When enabled:
* Users can access the [**Email Aliases** page](/docs/panel/emails/aliases/)

When disabled:
* Users can not access the Email Aliases page.


## Address Importer

The **`email_import`** module allows users to import email addresses from a CSV file.

When enabled:
* Users can access the [**Address Importer** page](/docs/panel/emails/import/)

When disabled:
* Users can not access the Address Importer page.


## Address Exporter

The **`email_export`** module allows users to export to CSV file all their email addresses.

When enabled:
* Users can export email addresses by visiting: `/emails/export`

When disabled:
* Users can not access the Address Exporter page.



## Default Address

The **`email_default`** module allows users to create default (catch-all) email addresses.

When enabled:
* Users can access the [**Default Email Address** page](/docs/panel/emails/default_address/)

When disabled:
* Users can not access the Default Email Address page.



## Email Deliverability

The **`email_deliverability`** module allows users to check and validate SPF, DKIM and DMARC records for their domains.

When enabled:
* Users can access the [**Email Deliverability** page](/docs/panel/emails/email_deliverability/) to view DNS records.

When disabled:
* Users can not access the Email Deliverability page.


## MySQL

The **`mysql`** module allows users to create and manage mysql databases.

When enabled:
* MySQL/MariaDB auto-starts when user accesses Databases section, opens phpMyAdmin or installs WordPress.
* Users can access the [**MySQL > Databases** page](/docs/panel/mysql/databases/) to manage databases.
* Users can access the [**MySQL > New Database** page](/docs/panel/mysql/new_db/) to create databases.
* Users can access the [**MySQL > Database Wizard** page](/docs/panel/mysql/wizard/) to create database, user and assign privileges.
* Users can access the [**MySQL > Users** page](/docs/panel/mysql/users/) to manage users.
* Users can access the [**MySQL > New User** page](/docs/panel/mysql/new_user/) to create users.
* Users can access the [**MySQL > Change Password** page](#) to change password for a user.
* Users can access the [**MySQL > Assign User to DB** page](/docs/panel/mysql/assign/) to assign all privileges to user over a database.
* Users can access the [**MySQL > Remove User from DB** page](/docs/panel/mysql/remove/) to revoke all privileges to user over a database.

When disabled:
* Users do not have access to the *MySQL* section.

Customize options:
* To **set mysql or mariadb for all new users** use [*OpenAdmin > Settings > User Defaults* page and *MySQL type* option](/docs/admin/settings/defaults/).
* To **set mysql, percona or mariadb for a single user** when creating their account use the [**MySQL Type** option](/docs/articles/docker/how-to-set-mysql-mariadb-per-user-in-openpanel/).
* To **change default CPU/RAM for service** use the [*OpenAdmin > Settings > User Defaults* page](/docs/admin/settings/defaults/).
* To **restrict access to system users** edit the [`mysql_restricted_usernames`](https://dev.openpanel.com/cli/config.html#mysql-restricted-usernames) setting.
* To **restrict access to system databases** edit the [`mysql_restricted_databases`](https://dev.openpanel.com/cli/config.html#mysql-restricted-databases) setting.
* To **increase the startup time allowed for waiting MySQL to initalize** increase [`mysql_startup_time`](https://dev.openpanel.com/cli/config.html#mysql-startup-time).


How-to guides:
* To **connect to a database** refer to [*How-to Guides > Connecting to MySQL Server from Applications in OpenPanel](/docs/articles/databases/how-to-connect-to-mysql-from-php-applications-in-openpanel/).
* To **troubleshoot errors** refer to [*How-to Guides > How to troubleshoot: Error establishing a database connection](/docs/articles/databases/how-to-troubleshoot-error-establishing-a-database-connection/).


## MySQL Root Password

The **`mysql_root_password`** module allows users to change password for their MySQL root user from the user panel.

When enabled:
* Users can access the [**MySQL > Root Password** page](#) to change root user password.

When disabled:
* Users do not have access to the *MySQL Root Password* page.



## MySQL Show Processes

The **`mysql_processlist`** module allows users to view the MySQL process list.

When enabled:
* Users can access the [**MySQL > Process List** page](/docs/panel/mysql/processlist/) to view all active processes.

When disabled:
* Users do not have access to the *MySQL Process List* page.


## Remote MySQL

The **`remote_mysql`** module allows users to enable/disable remote access to mysql.

When enabled:
* Remote access is disabled by default.
* Random port is allocated per user for their mysql instances.
* Users can access the [**MySQL > Remote Access** page](/docs/panel/mysql/remote/) to enable/disable remote access.
* Users can connect to any database from remote location once the option is enabled.

When disabled:
* Remote access is disabled.

Customize options:
* None


## phpMyAdmin

The **`phpmyadmin`** module allows users to manage phpMyAdmin service.

When enabled:
* phpMyAdmin can be accessed by the user.
* phpMyAdmin is available on a `https://DOMAIN:2053` when domain is set, or `http://IP:8888` when IP is set for panel access.

When disabled:
* Users do not have access to the *phpMyAdmin* service.

Customize options:
* To **change CPU/RAM for phpMyAdmin service** or values: **php_max_execution_time, php_memory_limit, php_upload_limit** use [*OpenAdmin > Services > Service Limits](/docs/admin/services/limits/).

How-to guides:
* To **import tables into a database** refer to [**the Documentation**](/docs/panel/mysql/phpmyadmin/#import-sql-files).

## MySQL Import

The **`mysql_import`** module allows users to import files into their databases.

When enabled:
* Users can access the [**MySQL > Import Database** page](/docs/panel/mysql/import/) to import files into a database.

When disabled:
* Users can not access the *MySQL > Import Database* page.

Customize options:
* To **set the max file size allowed for import** increase [`mysql_import_max_size_gb`](https://dev.openpanel.com/cli/config.html#mysql-import-max-size-gb) value.

How-to guides:
* To **import into a database** refer to [*How-to Guides > Importing a Database](/docs/articles/docker/import-database/).


## MySQL Conf

The **`mysql_conf`** module allows users to edit mysql server configuration.

When enabled:
* Users can access the [**MySQL > Edit Configuration** page](#) to edit service .cnf file.

When disabled:
* Users can not access the *MySQL > Edit Configuration* page.

Customize options:
* To **set available options for configuration** edit file: [`/etc/openpanel/mysql/keys.txt`](https://github.com/stefanpejcic/openpanel-configuration/blob/main/mysql/keys.txt).
* To **edit the mysql.cnf file for a single user** edit file: `/home/${username}/custom.cnf`.
* To **edit the mysql.cnf file for all new users** edit file: [`/etc/openpanel/mysql/user.cnf`](https://github.com/stefanpejcic/openpanel-configuration/blob/main/mysql/user.cnf).




## PostgreSQL

The **`postgresql`** module allows users to create and manage PostgreSQL databases and users.

When enabled:
* Users can access the *PostgreSQL* section.

When disabled:
* Users do not have access to the *PostgreSQL* section.

## Remote PostgreSQL

The **`remote_postgresql`** module allows users to enable/disable remote access to PostgreSQL.

When enabled:
* Remote access is disabled by default.
* Random port is allocated per user for their PostgreSQL instances.
* Users can access the [**PostgreSQL > Remote Access** page](#) to enable/disable remote access.
* Users can connect to any database from remote location once the option is enabled.

When disabled:
* Remote access is disabled.

Customize options:
* None


## pgAdmin

The **`pgadmin`** module allows users to manage pgAdmin service.

When enabled:
* pgAdmin can be managed by the user.
* Users have access to the *pgAdmin* section.
* pgAdmin is available on a custom per-user port.

When disabled:
* Users do not have access to the *pgAdmin* section.

Customize options:
* To **change default CPU/RAM for pgAdmin** use the 'manage' button in top-rgiht corner.


## PostgreSQL Import

The **`postgresql_import`** module allows users to import files into their databases.

When enabled:
* Users can access the [**PostgreSQL > Import Database** page](#) to import files into a database.

When disabled:
* Users can not access the *PostgreSQL > Import Database* page.

Customize options:
* None


## PostgreSQL Conf

The **`postgresql_conf`** module allows users to edit PostgreSQL server configuration.

When enabled:
* Users can access the [**PostgreSQL > Edit Configuration** page](#) to edit service .cnf file.

When disabled:
* Users can not access the *PostgreSQL > Edit Configuration* page.


## Crons

The **`crons`** module allows users to schedule [Ofelia](https://hub.docker.com/r/mcuadros/ofelia) cron jobs.

When enabled:
* Users can access the [**Advanced > Cron Jobs** page](/docs/panel/advanced/cronjobs/).
* Users can [add cronjobs](/docs/panel/advanced/cronjobs/#add)
* Users can [edit cronjobs](/docs/panel/advanced/cronjobs/#edit)
* Users can [view logs for cronjobs](/docs/panel/advanced/cronjobs/#logs)
* Users can [edit crons file](/docs/panel/advanced/cronjobs/#file-editor)
* Users can [import and export cronjobs](/docs/panel/advanced/cronjobs/#import--export)

When disabled:
* Users can not access the *Advanced > Cron Jobs* page nor modify crons.

Customize options:
* To **pre-set cronjobs for new users** edit the `/etc/openpanel/ofelia/users.ini` file.
* To **set max file size for the cron file to be editable via OpenPanel UI** set the [`cron_max_file_size_kb`](https://dev.openpanel.com/cli/config.html#cron-max-file-size-kb) value.



## Process Manager

The **`process_manager`** module allows users to view and terminate processes from all running services.

When enabled:
* Users can access the [**Advanced > Process Manager** page](/docs/panel/advanced/process_manager/).

When disabled:
* Users can not access the *Advanced > Process Manager* page.

Customize options:
* None


## Server Info

The **`info`** module allows users to view server information, hosting plan information and OpenPanel information.

When enabled:
* Users can access the [**Advanced > Server Information** page](/docs/panel/advanced/server_info/).

When disabled:
* Users can not access the *Advanced > Server Information* page.

Customize options:
* None



## Temporary Links

The **`temporary_links`** module allows users to test their websites using temporary subdomains (links are valid for 15 minutes).

When enabled:
* Users can access the [**Live Preview** button on the Site Manager](/docs/panel/applications/wordpress/#temporary-link).

When disabled:
* Users can not access the *Live Preview* button on the Site Manager page.

Customize options:
* To **self-host a proxy service** - refer to [How-to Guides > Temporary Links API](/docs/articles/dev-experience/selfhosted-temporary-links-api/).
* To **configure a custom domain** - update the [`temporary_links` option](https://dev.openpanel.com/cli/config.html#temporary-links).

## Login History

The **`login_history`** module allows users to view login history for their account.

When enabled:
* Users can access the [**Account > Login History** page](/docs/panel/account/login_history/).

When disabled:
* Users can not access the *Account > Login History* page.

Customize options:
* To **control number of logins stored per user** edit 'Login records to keep per user' setting from [OpenAdmin > Settings > OpenPanel](/docs/admin/settings/openpanel/#Statistics).


## 2FA

The **`twofa`** module allows users to enable 2 factor authentication for their account.

When enabled:
* Users can access the [**Account > Two-Factor Authentication** page](/docs/panel/account/2fa).
* 2FA is required on login page if account has enabled it.

When disabled:
* Users can not access the *Advanced > Two-Factor Authentication* page nor manage 2FA.

Customize options:
* To **enable 2FA widget** use [*OpenAdmin > Settings > OpenPanel* page and *Display 2FA widget* option](/docs/admin/settings/openpanel/).
* To **enforce 2FA for all users** use [*OpenAdmin > Settings > OpenPanel* page and *Enforce 2FA* option](/docs/admin/settings/openpanel/).
* To **check 2FA status for a user** refer to [How to check if 2FA is active for OpenPanel user account?](https://community.openpanel.org/d/220-how-to-check-if-2fa-is-active-for-openpanel-user-account).

## Passkeys

The **`passkeys`** module allows users to seup [Passkeys](https://safety.google/intl/en_in/safety/authentication/passkey/) for their account.

NOTE: Passkeys require that a domain name is used for panel access.

When enabled:
* Users can access the [**Account > Passkeys** page](/docs/panel/account/passkeys).
* 'Sign up with Passkey' is shown on login page.

When disabled:
* Users can not login using Passkeys.

## Activity

The **`activity`** module allows users to view their activity logs.

When enabled:
* Users can access the [**Account > Activity Log** page](/docs/panel/account/account_activity).

When disabled:
* Users can not access the *Account > Activity Log* page.

Customize options:
* To **edit activity log from terminal** open file: `/etc/openpanel/openpanel/core/users/{username}/activity.log`.
* To **set total number of lines per user** edit `activity_lines_retention` setting.
* To **set total size of log per user** edit `activity_max_size_bytes` setting.
* To **log actions from 3rd-party plugin** refer to: [*How to log actions from Custom Plugins in user Activity Log*](https://community.openpanel.org/d/218-how-to-log-actions-from-custom-plugins-in-user-activity-log)


## Backups

The **`backups`** module allows users to configure their own backups: what to backup, destination, retention, schedule, etc.

When enabled:
* Users can access the [**Files > Backups** page](/docs/panel/files/backups/).
* Users can configure backup schedule, encryption, retention and destination.

When disabled:
* Users do not have access to the *Files > Backups* page.
* [Administrators need to configure backups for the user](/docs/articles/backups/configuring-backups/#1-admin-configured).

## Backup Wizard

The **`backup_wizard`** module allows users to generate and download a full account backup.

When enabled:
* Users can access the *Backup Wizard* to generate and download a full backup of their account.

When disabled:
* Users can not access the Backup Wizard.



## Services

The **`services`** module allows users to enable/disable services without the Docker module.

When enabled:
* Users can access the [**Advanced > Services** page](/docs/panel/advanced/services/).
* Users can enable/disable services.
* User view current service status, resource usage (CPU%, Memory%, Disk I/O, PIDs..), container name (to be used to connect to service from other containers).
* Users can view logs for services.

When disabled:
* Users do not have access to the *Advanced > Services* page.


## Memcached

The **`memcached`** module allows users to enable/disable Memcached service.

When enabled:
* Users can access the [**Caching > Memcached** page](/docs/panel/caching/memcached/).
* Users can enable/disable Memcached service.
* User can connect to the instance from other containers using: `elasticsearch:11211`
* Users can view logs for the Memcached service.

When disabled:
* Users do not have access to the *Caching > Memcached* page.

## Redis

The **`redis`** module allows users to enable/disable Redis service.

When enabled:
* Users can access the [**Caching > Redis** page](/docs/panel/caching/redis/).
* Users can enable/disable Redis service.
* User can connect to the instance from other containers using: `redis:6379`
* Users can view logs for the Redis service.

When disabled:
* Users do not have access to the *Caching > Redis* page.

## Valkey

The **`valkey`** module allows users to enable/disable Valkey service.

When enabled:
* Users can access the [**Caching > Valkey** page](/docs/panel/caching/valkey/).
* Users can enable/disable Valkey service.
* User can connect to the instance from other containers using: `valkey:6379`
* Users can view logs for the Valkey service.

When disabled:
* Users do not have access to the *Caching > Valkey* page.


## ElasticSearch

The **`elasticsearch`** module allows users to enable/disable ElasticSearch service.

When enabled:
* Users can access the [**Caching > ElasticSearch** page](/docs/panel/caching/elasticsearch/).
* Users can enable/disable ElasticSearch service.
* User can connect to the instance from other containers using: `elasticsearch:9200`
* Users can view logs for the ElasticSearch service.

When disabled:
* Users do not have access to the *Caching > ElasticSearch* page.



## OpenSearch

The **`opensearch`** module allows users to enable/disable OpenSearch service.

When enabled:
* Users can access the [**Caching > OpenSearch** page](/docs/panel/caching/opensearch/).
* Users can enable/disable OpenSearch service.
* User can connect to the instance from other containers using: `opensearch:9200`
* Users can view logs for the OpenSearch service.

When disabled:
* Users do not have access to the *Caching > OpenSearch* page.


## Disk Usage Explorer

The **`disk_usage`** module allows users to view disk usage per-directory.

When enabled:
* Users can access the [**Files > Disk Usage** page](/docs/panel/files/disk_usage/).

When disabled:
* Users do not have access to the *Files > Disk Usage* page.


## Inodes Explorer

The **`inodes`** module allows users to view inode usage per-directory.

When enabled:
* Users can access the [**Files > Inodes Explorer** page](/docs/panel/files/inodes/).

When disabled:
* Users do not have access to the *Files > Inodes Explorer* page.





## AutoInstaller
The **`autoinstaller`** module allows users to autoinstall WordPress, website Builder, Mautic, Python/NodeJS applications, etc.

When enabled:
* Users can access the [**Websites > Auto Installer** page](/docs/panel/applications/autoinstaller/).

When disabled:
* Users do not have access to the *Websites > Auto Installer* page.


## Drupal

The **`drupal`** module allows users to install and manage Drupal websites.

When enabled:
* Drupal is available on the Autoinstaller page.
* Users can install Drupal using Auto Installer.
* Users can manage Drupal websites using Site Manager: clone a site, run/restore backups, generate a one-time admin login link, rebuild the cache, and view watchdog logs.
* Users can update Drupal core in place using the *Update* tab (drush-based, one click).

When disabled:
* Drupal is not available on the Autoinstaller page.
* Drupal websites can not be managed via Openpanel.

## Joomla

The **`joomla`** module allows users to install and manage Joomla websites.

When enabled:
* Joomla is available on the Autoinstaller page.
* Users can install Joomla using Auto Installer.
* Users can manage Joomla websites using Site Manager: clone a site and run/restore backups.
* The *Update* tab shows whether a newer Joomla version is available and links directly into the site's own Joomla Admin to run the update (Joomla has no safe unattended CLI updater).

When disabled:
* Joomla is not available on the Autoinstaller page.
* Joomla websites can not be managed via Openpanel.

## OpenCart

The **`opencart`** module allows users to install and manage OpenCart websites.

When enabled:
* OpenCart is available on the Autoinstaller page.
* Users can install OpenCart using Auto Installer.
* Users can manage OpenCart websites using Site Manager: clone a site and run/restore backups.
* The *Update* tab shows whether a newer OpenCart version is available and links directly into the site's own admin panel to run the update.

When disabled:
* OpenCart is not available on the Autoinstaller page.
* OpenCart websites can not be managed via Openpanel.

## Nextcloud

The **`nextcloud`** module allows users to install and manage Nextcloud websites.

When enabled:
* Nextcloud is available on the Autoinstaller page.
* Users can install Nextcloud using Auto Installer.
* Users can manage Nextcloud websites using Site Manager: clone a site, run/restore backups, and view logs.
* Users can update Nextcloud in place using the *Update* tab (one click).

When disabled:
* Nextcloud is not available on the Autoinstaller page.
* Nextcloud websites can not be managed via Openpanel.

## PrestaShop

The **`prestashop`** module allows users to install and manage PrestaShop websites.

When enabled:
* PrestaShop is available on the Autoinstaller page.
* Users can install PrestaShop using Auto Installer.
* Users can manage PrestaShop websites using Site Manager: clone a site and run/restore backups.
* The *Update* tab shows whether a newer PrestaShop version is available and links directly into the site's own admin panel to run the update.

When disabled:
* PrestaShop is not available on the Autoinstaller page.
* PrestaShop websites can not be managed via Openpanel.

## Matomo

The **`matomo`** module allows users to install and manage Matomo websites.

When enabled:
* Matomo is available on the Autoinstaller page.
* Users can install Matomo using Auto Installer.
* Users can manage Matomo websites using Site Manager: clone a site and run/restore backups.
* Users can update Matomo in place using the *Update* tab (one click).

When disabled:
* Matomo is not available on the Autoinstaller page.
* Matomo websites can not be managed via Openpanel.

## Moodle

The **`moodle`** module allows users to install and manage Moodle websites.

When enabled:
* Moodle is available on the Autoinstaller page.
* Users can install Moodle using Auto Installer.
* Users can manage Moodle websites using Site Manager: clone a site and run/restore backups.
* Users can update Moodle in place using the *Update* tab (one click).

When disabled:
* Moodle is not available on the Autoinstaller page.
* Moodle websites can not be managed via Openpanel.

## MediaWiki

The **`mediawiki`** module allows users to install and manage MediaWiki websites.

When enabled:
* MediaWiki is available on the Autoinstaller page.
* Users can install MediaWiki using Auto Installer, choosing a specific release from the version dropdown.
* Users can manage MediaWiki websites using Site Manager: clone a site and run/restore backups.
* Users can update MediaWiki in place using the *Update* tab (one click).

When disabled:
* MediaWiki is not available on the Autoinstaller page.
* MediaWiki websites can not be managed via Openpanel.

## Flarum

The **`flarum`** module allows users to install and manage Flarum forums.

When enabled:
* Flarum is available on the Autoinstaller page.
* Users can install Flarum using Auto Installer, choosing a specific version from the version dropdown.
* Users can manage Flarum forums using Site Manager: clone a forum, run/restore backups, and view logs.
* Users can update Flarum in place using the *Update* tab (Composer-based, one click).

When disabled:
* Flarum is not available on the Autoinstaller page.
* Flarum forums can not be managed via Openpanel.

## SofaWiki

The **`sofawiki`** module allows users to install and manage SofaWiki wikis.

When enabled:
* SofaWiki is available on the Autoinstaller page.
* Users can install SofaWiki using Auto Installer.
* Users can manage SofaWiki wikis using Site Manager: clone a wiki and run/restore backups.

When disabled:
* SofaWiki is not available on the Autoinstaller page.
* SofaWiki wikis can not be managed via Openpanel.

> **NOTE:** SofaWiki has no database, no tagged releases, and no CLI installer - the *Update* tab is not available for this type; the site owner completes SofaWiki's own setup wizard in the browser after install.

## TinyPhotoGallery

The **`tinyphotogallery`** module allows users to install and manage TinyPhotoGallery photo galleries.

When enabled:
* TinyPhotoGallery is available on the Autoinstaller page.
* Users can install TinyPhotoGallery using Auto Installer.
* Users can remove TinyPhotoGallery installations using Site Manager.

When disabled:
* TinyPhotoGallery is not available on the Autoinstaller page.
* TinyPhotoGallery installations can not be managed via Openpanel.

> **NOTE:** TinyPhotoGallery has no database, no admin account, no tagged releases, and no CLI installer - install is just downloading a single `index.php` file and creating an empty `photos/` folder next to it. Files-only backup/restore is supported (whole-docroot tar/untar); there is no clone support and no *Update* tab.

## TinyFileManager

The **`tinyfilemanager`** module allows users to install and manage TinyFileManager instances.

When enabled:
* TinyFileManager is available on the Autoinstaller page.
* Users can install TinyFileManager using Auto Installer.
* Users can remove TinyFileManager installations using Site Manager.

When disabled:
* TinyFileManager is not available on the Autoinstaller page.
* TinyFileManager installations can not be managed via Openpanel.

> **NOTE:** TinyFileManager has no database, no tagged releases, and no CLI installer - install downloads a single `tinyfilemanager.php` file and writes the admin username/password provided on the install form directly into that file's `$auth_users` array (bcrypt-hashed). Files-only backup/restore is supported; there is no clone support and no *Update* tab.

## DokuWiki

The **`dokuwiki`** module allows users to install and manage DokuWiki wikis.

When enabled:
* DokuWiki is available on the Autoinstaller page.
* Users can install DokuWiki using Auto Installer - the admin account and site configuration are created automatically, no browser setup wizard needed.
* Users can manage DokuWiki wikis using Site Manager: clone a wiki and run/restore backups.
* Users can update DokuWiki in place using the *Update* tab (one click) - `conf/`, `data/` and `lib/plugins/` are always preserved.

When disabled:
* DokuWiki is not available on the Autoinstaller page.
* DokuWiki wikis can not be managed via Openpanel.

## phpBB

The **`phpbb`** module allows users to install and manage phpBB forums.

When enabled:
* phpBB is available on the Autoinstaller page.
* Users can install phpBB using Auto Installer, choosing a specific version from the version dropdown (populated from phpBB's GitHub releases).
* Users can manage phpBB forums using Site Manager: clone a forum and run/restore backups.
* The *Update* tab links directly into the site's own Admin Control Panel to run phpBB's Automatic Update Package (phpBB has no safe unattended CLI updater).

When disabled:
* phpBB is not available on the Autoinstaller page.
* phpBB forums can not be managed via Openpanel.


## PHP.INI Editor
The **`php_ini`** module allows users to edit the PNP.INI files using a text editor.

When enabled:
* Users can access the [**PHP > PHP.INI Editor** page](/docs/panel/php/php_ini_editor/).

When disabled:
* Users do not have access to the *PHP > PHP.INI Editor* page.


## WordPress

The **`wordpress`** module allows users to install and manage WordPress websites.

When enabled:
* Users can access the [**Websites > WP Manager** page](/docs/panel/applications/wordpress/).
* Users can [manage WordPress websites using WP Manager](/docs/panel/applications/wordpress/#site-manager).
* WordPress is available on the Autoinstaller page.
* Users can [install WordPress using Auto Installer](/docs/panel/applications/wordpress/#install-wordpress).
* Users can [scan and import existing installations](/docs/panel/applications/wordpress/#scanning-importing-installations).
* Users can [set themes and plugins to auto-install](/docs/panel/applications/wordpress/#themes-and-plugins-sets).

When disabled:
* Users can not access the *Websites > WP Manager* page.
* WordPress is not available in Autoinstaller.
* WordPress websites can not be managed via Openpanel.

Customize options:
* To **auto install themes or plugins on new installations** refer to: [*WordPress Themes and Plugins Sets*](/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/)
* To **add a custom Google PageSpeed Insights API Key** refer to: [*How-to Guides > Google PageSpeed Insights API Key*](/docs/articles/websites/google-pagespeed-insights-api-key/)
* To **setup a mu-plugin on all new websites** edit `/etc/openpanel/wordpress/mu-plugin.php` file.
* To **set a custom WP-CLI for all websites** replace the `/etc/openpanel/wordpress/wp-cli.phar` file.
* To **customize .htaccess files used for new websites** edit files in `/etc/openpanel/wordpress/htaccess/` folder.

## Website Builder

The **`website_builder`** module allows users to create simple websites using the HTML Drag & Drop Website Builder.

When enabled:
* Users can access the [**Websites > Website Builder** page](/docs/panel/applications/builder/).
* Users can [manage static websites using Site Manager](/docs/panel/applications/builder/#edit-website).
* Website Builder is available on the Autoinstaller page.
* Users can [create static websites using Auto Installer](/docs/panel/applications/builder/#create-a-website).

When disabled:
* Users can not access the *Websites > Website Builder* page.
* Website Builder is not available in Autoinstaller.
* Static websites can not be managed via Openpanel.

## ClamAV

The **`malware_scan`** module starts a ClamAV service and allows users to scan files.

> **NOTE:** This module is tagged *BETA*.

When enabled:
* Users can access the [**Files > Malware Scanner** page](/docs/panel/files/malware-scanner/).
* ClamAV service is started on the server.

When disabled:
* Users can not access the *Files > Malware Scanner* page.
* ClamAV service is not started on the server.

Customize options:
* To **customize the cpu/memory limits for the ClamAV service** refer to: [*OpenAdmin > Services > Service Limits*](/docs/admin/services/limits/).



## Files

The **`filemanager`** module allows users to manage files and folders using the File Manager.

When enabled:
* Users can access the [**Files > File Manager** page](/docs/panel/files/).
* File Manager links are available on other pages: Domains, WP Manager, etc.

When disabled:
* Users can not access the *Files > File Manager* page.
* No links to manage files are shown on other pages.


## Trash

The **`trash`** module allows users to manage their Trash folder.

When enabled:
* Users can access the [**Files > Trash** page](/docs/panel/files/).
* User can delete to trash instead of permanently deleting files.

When disabled:
* Users can not access the *Files > Trash* page.


## Domains

The **`domains`** module allows users to add and manage domains.

When enabled:
* Users can access the [**Domains** page](/docs/panel/domains/).
* Users can manage domains.
* Users can access the 'Domains' sub-pages in the menu.

When disabled:
* Users can not access the *Domains* page.
* Users can not manage domains.

Customize options:
* To **enable HSTS for a domain** refer to:  [*How-to Guides > How to Enable HSTS on a Domain in OpenPanel*](/docs/articles/domains/how-to-enable-hsts-on-a-domain-in-openpanel/)
* To **customize default pages** refer to: [*OpenAdmin > Domains > Edit Domain Templates*](/docs/admin/domains/file_templates/)

## SSL

The **`ssl`** module allows users to view SSL configuration for their domains and add custom certificates.

When enabled:
* Users can access the [**Domains > SSL** page](/docs/panel/domains/ssl/) to view SSL status and add custom certificates.

When disabled:
* Users can not access the *Domains > SSL* page.


## Suspend Domains

The **`domain_suspend`** module allows users to suspend/unsuspend website access.

When enabled:
* Users can access the [**Suspend a Domain** page](/docs/panel/domains/suspend/).
* Users can access the [**Unsuspend a Domain** page](/docs/panel/domains/unsuspend/).

When disabled:
* Users can not suspend/unsuspend domains.

Customize options:
* To **customize the suspended domain template** use:  [*OpenAdmin > Domains > Edit Domain Templates*](/docs/admin/domains/file_templates/#suspended-website)

## Raw Access Logs

The **`domain_logs`** module allows users to view the raw access log for their domains.

When enabled:
* Users can access the [**Domains > Raw Access Logs** page](/docs/panel/domains/docroot/).

When disabled:
* Users can not access the *Domains > Raw Access Logs* page.



## GoAccess

The **`goaccess`** module runs the GoAccess service on a scheduled basis to process raw domain logs and produce HTML reports accessible through the OpenPanel UI.

When enabled:
* GoAccess service is run on the server.
* Users can access the [**Domains > GoAccess** page](/docs/panel/domains/goaccess/).

When disabled:
* Users can not access the *Domains > GoAccess* page.

Customize options:
* To **disable GoACcess report generation** update the: [*`goaccess_enable` value*](https://dev.openpanel.com/cli/config.html#goaccess-enable)
* To **change how often the reports are generated (default = @daily)** edit the schedule for `domains-stats` cron and the [`goaccess_schedule` value](https://dev.openpanel.com/cli/config.html#goaccess-schedule).
* To **generate the data manually** execute `domains-stats` cron.
* To **force regeneration of the reports* refer to: [*OpenCLI Documentation > Parse domain access logs*](https://dev.openpanel.com/cli/domains.html#Parse-domain-access-logs).


## Docroot

The **`docroot`** module allows users to set a custom docroot (folder) when adding domains, and later change the path.

When enabled:
* Users can access the [**Domains > Change Docroot** page](/docs/panel/domains/docroot/).
* Users can set a custom docroot when adding a domain.

When disabled:
* Users can not set a custom docroot when adding a domaina, and can not later change the docroot.

## Redirects

The **`redirects`** module allows users to create redirects for domains.

When enabled:
* Users can access the [**Domains > Redirects** page](/docs/panel/domains/redirects/).

When disabled:
* Users can not access the *Domains > Redirects* page.



## Capitalize Domains

The **`capitalize_domains`** module allows users to set a capitalized version fo the domain for dispaly in the OpenPanel.

When enabled:
* Users can access the [**Domains > Capitalize Domains** page](/docs/panel/domains/capitalize/).

When disabled:
* Users can not access the *Domains > Capitalize Domains* page.


## Edit VirtualHosts

The **`edit_vhost`** module allows users to edit the VirtualHosts files for their domains.

When enabled:
* Users can access the [**Domains > Edit VHosts File** page](/docs/panel/domains/vhosts/).

When disabled:
* Users can not access the *Domains > Edit VHosts File* page.

Customize options:
* To **customize the vhost files for Apache/Nginx/OpenLiteSpeed** refer to: [*OpenAdmin > Domains > Edit Domain Templates*](/docs/admin/domains/file_templates/#apache-virtualhost)


## Webserver

The **`webserver_conf`** module allows users to edit the main configuration files for their webservers.

When enabled:
* Users can access the [**Advanced > WebServer Settings** page](/docs/panel/advanced/webserver_settings/).
* Users can edit the `httpd.conf` file for Apache.
* Users can edit the `nginx.conf` file for Nginx/OpenResty.
* Users can edit the `openlitespeed.conf` file for OpenLiteSpeed.

When disabled:
* Users can not access the *Advanced > WebServer Settings* page.

Customize options:
* To **customize the default `httpd.conf` file for Apache** edit `/etc/openpanel/apache/httpd.conf` file.
* To **customize the default `nginx.conf` file for Nginx** edit `/etc/openpanel/nginx/nginx.conf` file.
* To **customize the default `openlitespeed.conf` file for OpenLiteSpeed** edit `/etc/openpanel/openlitespeed/httpd_config.conf` file.
* To **customize the default `nginx.conf` file for OpenResty** edit `/etc/openpanel/openresty/nginx.conf` file.

## DNS

The **`dns`** module runs a local BIND9 service, creates zone files for domains and allows users to manage DNS records.

When enabled:
* BIND9 service is run on the server.
* Users can access the [**Domains > DNS Zone Editor** page](/docs/panel/domains/dns/).
* DNS zone files are created for new domains.
* Users can manage DNS records.
* 'Edit Zone' links are available for domains under the *OpenPanel > Domains* page.
* Administrators can access the [**OpenAdmin > Domains > DNS Cluster** page](/docs/admin/domains/dns-cluster/).
* Administrators can access the [**OpenAdmin > Domains > Edit Zone Templates** page](/docs/admin/domains/dns_templates/).
* Administrators can access the [**OpenAdmin > Domains > DNS Zone Editor** page](/docs/admin/domains/dns/).

When disabled:
* Users can not access the *Domains > DNS Zone Editor* page.
* Administrators can not access the *DNS Zone Editor*, *Edit Zone Templates*, and *DNS Cluster* pages in OpenAdmin.

Customize options:
* To **configure nameservers** refer to: [*How-to Guides > Configure Nameservers*](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)
* To **customize DNS zone templates** refer to: [OpenAdmin > Domains > Edit Zone Templates](/docs/admin/domains/dns_templates/)
* To **configure a DNS cluster** refer to:  [*How-to Guides > DNS Clustering*](/docs/articles/domains/how-to-setup-dns-cluster-in-openpanel/)


## Dynamic DNS

The **`dynamic_dns`** module allows users to create subdomains that will be updated via webhooks.

When enabled:
* Users can access the [**Domains > Dynamic DNS** page](/docs/panel/domains/dynamic-dns/).

When disabled:
* Users can not access the *Domains > Dynamic DNS* page.


## IP Blocker

The **`ip_blocker`** module allows users to block IP addresses from accessing their websites.

When enabled:
* Users can access the [**Advanced > IP Blocker** page](/docs/panel/advanced/ip-blocker/).

When disabled:
* Users can not access the *Advanced > IP Blocker* page.

## WAF

The **`waf`** module runs a custom Caddy image with CorazaWAF and allows users to manage WAF rules and on/off protection per domain.

When enabled:
* `SecRuleEngine On` is set for new domains.
* Users can access the [**Advanced > WAF** page](/docs/panel/advanced/waf/).
* [OWASP CRS](https://github.com/coreruleset/coreruleset) is setup on installation.
* Users can edit WAF rules and enable/disable protection per domain.
* ['Firewall' widget is displayed in Site Manager](/docs/panel/applications/wordpress/#firewall).

When disabled:
* `SecRuleEngine Off` is set for new domains.
* `SecRuleEngine On` is replaced with `SecRuleEngine Off` for all existing domains.
* Users can not access the *Advanced > WAF* page.
* 'Firewall' widget is not displayed in Site Manager.

Customize options:
* [**WAF commmands**](https://dev.openpanel.com/cli/waf.html#CorazaWAF)

## PHP

The **`php`** module allows users to manage PHP versions and settings.

When enabled:
* Users can access the [**PHP > Select PHP Version** page](/docs/panel/php/domains/).
* Users can access the [**PHP > Default Version** page](/docs/panel/php/default/).
* Users can access the [**PHP > Extensions** page](/docs/panel/php/extensions/).
* Users can set PHP version per domain, set default version for new domains, edit options and view installed extensions.

When disabled:
* Users can not access the *Select PHP Version*, *Default Version*, *Options*, *Extensions* pages.
* Users can not set PHP version per domain, set default version for new domains, edit options and view installed extensions.

Customize options:
* To **set the default PHP version to be used for new users** refer to: [*OpenAdmin > Settings > Edit User Defaults > Default PHP version*](/docs/panel/php/options/#available-options)
* To **set default cpu/memory limits for PHP versions and additional PHP options** refer to: [*OpenAdmin > Settings > Edit User Defaults > Services*](/docs/panel/php/options/#available-options)
* To **install a PHP extension** refer to: [*How-to Guides > How to install a PHP extension in OpenPanel*](/docs/articles/websites/how-to-install-php-extensions-in-openpanel/).
* To **increase PHP INI memory_limit** refer to: [*How-to Guides > How to set or increase PHP INI memory_limit or other values?*](/docs/articles/websites/how-to-set-or-increase-PHP-INI-memory-limit-or-other-values/).
* To **set PHP settings per website** refer to: [*How-to Guides > PHP settings per website (folder)*](/docs/articles/websites/php-user-ini-files/).
* To **edit default .INI files** refer to: **OpenAdmin > Settings > PHP Settings > Default PHP.INI Files** or edit files in `/etc/openpanel/php/ini` folder.

## PHP Options

The **`php_options`** module allows users to manage options (limits) for their PHP versions.

When enabled:
* Users can access the [**PHP > Options** page](/docs/panel/php/options/).

When disabled:
* Users can not access the *PHP Options* page.

Customize options:
* To **customize PHP options available to users** refer to: **OpenAdmin > Settings > PHP Settings > Available Options** or edit */etc/openpanel/php/options.txt* file.

## PHP Extensions

> **NOTE:** This module is tagged *BETA*.

The **`php_extensions`** module allows users to manage extensions for their PHP versions.

When enabled:
* Users can access the [**PHP > Extensions** page](/docs/panel/php/extensions/).

When disabled:
* Users can not access the *PHP > Extensions* page.



## NodeJS

The **`nodejs`** module allows users to setup and manage containerized NodeJS applications. It is toggled independently from the [Python](#python) module below.

When enabled:
* Users can [manage NodeJS applications using Site Manager](/docs/panel/applications/pm2/#manage-applications).
* NodeJS is available on the Autoinstaller page.
* Users can [setup NodeJS applications using Auto Installer](/docs/panel/applications/pm2/#create-an-application).

When disabled:
* NodeJS is not available on the Autoinstaller page.
* NodeJS applications can not be managed via Openpanel.

Customize options:
* To **customize docker service template for new Node.JS applications** edit `/etc/openpanel/docker/compose/nodejs.yml` file.
* To **customize headers for Nginx proxy of new NodeJS applications** edit `/etc/openpanel/nginx/vhosts/1.1/nginx_proxy_headers.txt` file.
* To **add a custom Google PageSpeed Insights API Key** refer to: [*How-to Guides > Google PageSpeed Insights API Key*](/docs/articles/websites/google-pagespeed-insights-api-key/)


## Python

The **`python`** module allows users to setup and manage containerized Python applications. It is toggled independently from the [NodeJS](#nodejs) module above.

When enabled:
* Users can [manage Python applications using Site Manager](/docs/panel/applications/pm2/#manage-applications).
* Python is available on the Autoinstaller page.
* Users can [setup Python applications using Auto Installer](/docs/panel/applications/pm2/#create-an-application).

When disabled:
* Python is not available on the Autoinstaller page.
* Python applications can not be managed via Openpanel.

Customize options:
* To **customize docker service template for new Python applications** edit `/etc/openpanel/docker/compose/python.yml` file.
* To **customize headers for Nginx proxy of new Python applications** edit `/etc/openpanel/nginx/vhosts/1.1/nginx_proxy_headers.txt` file.
* To **add a custom Google PageSpeed Insights API Key** refer to: [*How-to Guides > Google PageSpeed Insights API Key*](/docs/articles/websites/google-pagespeed-insights-api-key/)


## Ruby

The **`ruby`** module allows users to setup and manage containerized Ruby applications. It is toggled independently from the [NodeJS](#nodejs) and [Python](#python) modules above.

When enabled:
* Users can [manage Ruby applications using Site Manager](/docs/panel/applications/pm2/#manage-applications).
* Ruby is available on the Autoinstaller page.
* Users can [setup Ruby applications using Auto Installer](/docs/panel/applications/pm2/#create-an-application), choosing a Ruby version from Docker Hub's official `ruby` image tags.

When disabled:
* Ruby is not available on the Autoinstaller page.
* Ruby applications can not be managed via Openpanel.

Customize options:
* To **customize docker service template for new Ruby applications** edit `/etc/openpanel/docker/compose/ruby.yml` file.
* To **customize headers for Nginx proxy of new Ruby applications** edit `/etc/openpanel/nginx/vhosts/1.1/nginx_proxy_headers.txt` file.


## Java

The **`java`** module allows users to setup and manage containerized Java applications. It is toggled independently from the [NodeJS](#nodejs), [Python](#python) and [Ruby](#ruby) modules above.

When enabled:
* Users can [manage Java applications using Site Manager](/docs/panel/applications/pm2/#manage-applications).
* Java is available on the Autoinstaller page.
* Users can [setup Java applications using Auto Installer](/docs/panel/applications/pm2/#create-an-application), choosing a JDK version (Docker Hub's official `eclipse-temurin` LTS tags).

When disabled:
* Java is not available on the Autoinstaller page.
* Java applications can not be managed via Openpanel.

Customize options:
* To **customize docker service template for new Java applications** edit `/etc/openpanel/docker/compose/java.yml` file.
* To **customize headers for Nginx proxy of new Java applications** edit `/etc/openpanel/nginx/vhosts/1.1/nginx_proxy_headers.txt` file.

> **NOTE:** Java apps run via Java 11+'s single-file source-code launch (`java Main.java`) by default - no build step needed for a simple app. Projects with a `pom.xml` can enable the "Run Maven install before starting the app" option to run `mvn install` first.


## Resources Usage

The **`usage`** module allows users to view resource usage for their services.

When enabled:
* Users can access the [**Advanced > Resource Usage** page](/docs/panel/advanced/resource_usage/).

When disabled:
* Users can not access the *Advanced > Resource Usage* page.

Customize options:
* To **edit page settings** refer to: [**OpenAdmin > Settings > OpenPanel > Statistics** page](/docs/admin/settings/openpanel/#statistics).
* To **change how often the stats are collected (default = @hourly)** edit the schedule for `docker-collect_stats --all` cron.
* To **display one combined or separate charts for cpu/ram** edit [`resource_usage_charts_mode` value](https://dev.openpanel.com/cli/config.html#resource-usage-charts-mode).
* To **change the number of items per page** edit [`resource_usage_items_per_page` value](https://dev.openpanel.com/cli/config.html#resource-usage-items-per-page).
* To **rotate the** edit [`resource_usage_retention` value](https://dev.openpanel.com/cli/config.html#resource-usage-retention).


## API

The **`api`** module allows users to access the OpenPanel API using JWT tokens.

When enabled:
* Users can access the [*Account > API Reference* page](/docs/panel/account/api/).
* Users can use the [OpenPanel API](/docs/panel/api/).

When disabled:
* Users can not use the OpenPanel API.

## MCP

The **`mcp`** module allows users to use the Model Context Protocol (MCP) to perform panel actions.

When enabled:
* Users can access the [**Account > MCP** page](/docs/panel/account/mcp/).

When disabled:
* Users can not access the *Account > MCP* page.
