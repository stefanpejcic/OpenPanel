---
sidebar_position: 4
---


# Modules

Modules (also known as plugins) extend the OpenPanel UI by adding new features and pages. To make a feature available to a user or plan, it must first be activated as a module.

Modules are loaded when the OpenPanel UI starts. A service restart is required to apply any newly added modules.

Available Modules:

| Name                | Title                       | Description                                                                                                | Link                                                           | Type       |
| ------------------- | --------------------------- | ---------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- | ---------- |
| notifications       | Email Notifications         | Allows users to configure notification preferences and receive emails.                                     | [/account/notifications](/account/notifications)               | community  |
| account             | Settings (Email & Password) | Enables users to change their username, password, and email address within their user panel.               | [/account/settings](/account/settings)                         | community  |
| sessions            | Active Sessions             | Lets users view all active sessions for their account.                                                     | [/account/sessions](/account/sessions)                         | community  |
| locale              | Locales (Language Change)   | Allows users to switch languages in their user panel.                                                      | [/account/languages](/account/languages)                       | community  |
| favorites           | Favorites (Bookmarks)       | Enables users to add and manage panel pages as favorites.                                                  | [/account/favorites](/account/favorites)                       | community  |
| varnish             | Varnish Caching             | Provides Varnish caching, allowing users to enable or disable caching per domain.                          | [/cache/varnish](/cache/varnish)                               | community  |
| docker              | Docker (Containers)         | Users can allocate resources and control lifecycle (start/stop) of containers.                             | [/containers](/containers)                                     | enterprise |
| ftp                 | FTP accounts                | Allows users to create and manage FTP accounts. Requires separate FTP server setup.                        | [/ftp](/ftp)                                                   | enterprise |
| emails              | Email accounts              | Enables users to create and manage email accounts. Requires separate mail server configuration.            | [/emails](/emails)                                             | enterprise |
| mysql               | MySQL                       | Users can manage MySQL databases and users.                                                                | [/mysql](/mysql)                                               | community  |
| remote\_mysql       | Remote MySQL                | Allows users to enable or disable remote access to MySQL databases.                                        | [/mysql/remote-mysql](/mysql/remote-mysql)                     | community  |
| mysql\_import       | Import MySQL Databases      | Enables importing .sql files into MySQL databases from the user panel.                                     | [/mysql/import](/mysql/import)                                 | community  |
| mysql\_conf         | MySQL Configuration         | Users can edit MySQL configuration settings via their panel.                                               | [/mysql/configuration](/mysql/configuration)                   | community  |
| postgresql          | PostgreSQL                  | Manage PostgreSQL databases and users.                                                                     | [/postgresql](/postgresql)                                     | beta       |
| remote\_postgresql  | Remote PostgreSQL           | Allows users to enable or disable remote access to PostgreSQL databases.                                   | [/postgresql/remote-postgresql](/postgresql/remote-postgresql) | beta       |
| postgresql\_import  | Import PostgreSQL Databases | Enables importing .sql files into PostgreSQL databases from the user panel.                                | [/postgresql/import](/postgresql/import)                       | beta       |
| postgresql\_conf    | PostgreSQL Configuration    | Users can edit PostgreSQL configuration settings via their panel.                                          | [/postgresql/configuration](/postgresql/configuration)         | beta       |
| pgamin              | pgAdmin                     | Provides access to pgAdmin for managing PostgreSQL databases.                                              | [/postgresql/pgamin](/postgresql/pgamin)                       | beta       |
| php                 | PHP                         | Users can change PHP versions per domain, set limits, and default versions for new domains.                | [/php](/php)                                                   | community  |
| php\_options        | PHP Options                 | Allows editing of PHP.INI directives using a user-friendly options page.                                   | [/php/options](/php/options)                                   | community  |
| php\_ini            | PHP.INI Editor              | Enables direct editing of PHP.INI files for any PHP version from OpenPanel.                                | [/php/ini](/php/ini)                                           | community  |
| phpmyadmin          | phpMyAdmin                  | Provides phpMyAdmin access to manage MySQL databases.                                                      | [/mysql/phpmyadmin](/mysql/phpmyadmin)                         | community  |
| crons               | Cronjobs                    | Allows users to configure scheduled cronjobs.                                                              | [/cronjobs](/cronjobs)                                         | community  |
| backups             | Backups                     | Enables users to configure backups for their accounts or services.                                         | [/backups](/backups)                                           | enterprise |
| wordpress           | WordPress                   | Supports WordPress installation via AutoInstaller and management through WP Manager.                       | [/wordpress](/wordpress)                                       | community  |
| website\_builder    | Website Builder             | Provides a drag-and-drop HTML website builder using GrapeJS editor.                                        | [/website-builder](/website-builder)                           | community  |
| pm2                 | Applications                | Allows setup of containerized NodeJS and Python apps via AutoInstaller.                                    | [/pm2](/pm2)                                                   | enterprise |
| autoinstaller       | AutoInstaller               | Enables managing websites and installing applications via an Auto-Installer page.                          | [/auto-installer](/auto-installer)                             | community  |
| disk\_usage         | Disk Usage Explorer         | Lets users browse disk usage statistics by directories.                                                    | [/disk-usage](/disk-usage)                                     | enterprise |
| inodes              | Inodes Explorer             | Provides browsing of inode usage statistics per directory.                                                 | [/inodes-explorer](/inodes-explorer)                           | enterprise |
| usage               | Resources Usage             | Collects docker stats for containers, viewable on Resource Usage page.                                     | [/usage](/usage)                                               | community  |
| info                | Server Info                 | Displays hosting plan limits and server info on Server Information page.                                   | [/server/info](/server/info)                                   | community  |
| webserver\_conf     | Apache/Nginx Configuration  | Allows editing of Apache/Nginx configurations for webserver containers.                                    | [/server/webserver\_conf](/server/webserver_conf)              | community  |
| waf                 | Coraza WAF                  | CorazaWAF is enabled automatically for new domains; users can toggle per domain.                           | [/waf](/waf)                                                   | enterprise |
| filemanager         | File Manager                | Provides access to manage website files within `/var/www/html/`. Disable for email-only or DNS-only plans. | [/files](/files)                                               | community  |
| fix\_permissions    | Fix Permissions             | Allows users to fix file ownership and permissions for website files.                                      | [/fix-permissions](/fix-permissions)                           | community  |
| dns                 | DNS                         | Runs local BIND9 server, creating DNS zones for new domains with editable records via Zone Editor.         | [/dns](/dns)                                                   | community  |
| redirects           | Domain Redirects            | Enables users to create and manage domain redirects.                                                       | [/domains/redirects](/domains/redirects)                       | enterprise |
| domains             | Domains                     | Allows users to add and manage domains. Disable for email-only or DB-only plans.                           | [/domains](/domains)                                           | community  |
| capitalize\_domains | Capitalize Domains          | Enables users to capitalize letters in their domains.                                                      | [/domains/capitalize](/domains/capitalize)                     | enterprise |
| malware\_scan       | Malware Scanner             | Runs ClamAV to scan domain docroots for malware; users can exclude directories and view quarantined files. | [/malware-scanner](/malware-scanner)                           | enterprise |
| domain\_logs        | Domain Logs                 | Allows users to view raw webserver access logs for their domains.                                          | [/domains/log](/domains/log)                                   | community  |
| goaccess            | GoAccess                    | Runs GoAccess to generate HTML reports from domain logs, accessible from Domain Logs page.                 | [/domains/stats](/domains/stats)                               | community  |
| process\_manager    | Process Manager             | Users can view and terminate running processes via the Process Manager page.                               | [/process-manager](/process-manager)                           | community  |
| redis               | Redis                       | Enables configuration of Redis caches from user panel.                                                     | [/cache/redis](/cache/redis)                                   | community  |
| memcached           | Memcached                   | Allows configuration of Memcached caches from user panel.                                                  | [/cache/memcached](/cache/memcached)                           | community  |
| elasticsearch       | Elasticsearch               | Enables users to configure Elasticsearch services from their panel.                                        | [/cache/elasticsearch](/cache/elasticsearch)                   | community  |
| opensearch          | Opensearch                  | Allows users to configure Opensearch services from their panel.                                            | [/cache/opensearch](/cache/opensearch)                         | community  |
| temporary\_links    | Temporary Links             | Allows users to test websites using temporary openpanel.org subdomains with 15-minute expiry links.        | [/websites](/websites)                                         | enterprise |
| login\_history      | Login History               | Displays the last 20 IP addresses that logged into the user's account.                                     | [/account/loginlog](/account/loginlog)                         | community  |
| twofa               | 2FA                         | Enables Two-Factor Authentication (2FA) for enhanced account security.                                     | [/account/2fa](/account/2fa)                                   | community  |
| activity            | Activity Log                | Allows users to browse all recorded actions and events related to their account.                           | [/account/activity](/account/activity)                         | community  |





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
