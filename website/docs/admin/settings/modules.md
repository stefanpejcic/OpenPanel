---
sidebar_position: 4
---

# Modules

Modules extend the OpenPanel UI by adding new features and pages. To make a feature available to a user or plan, it must first be activated as a module.

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
* Users can access [**Docker > Terminal**](/docs/panel/containers/terminal/) page to run docker exec commands.
* Users can access [**Docker > Image Updates**](/docs/panel/containers/image/) page to check for available image updates.
* Users can access [**Docker > Logs**](/docs/panel/containers/logs/) page to view service logs.
* Users can access [**Docker > Change Image Tag**](/docs/panel/containers/change/) page to change images tag.
* Users can access [**Docker > Switch Web Server**](/docs/panel/containers/webserver/) page to switch webservers.
* Users can access [**Docker > Switch MySQL Type**](/docs/panel/containers/mysql/) page to switch mysql/mariadb.

When disabled:
* Users can not access any of the *Docker* pages.

Customize options:
* None

## FTP

The **`ftp`** module allows users to create and manage FTP sub-accounts.

When enabled:
* Users can access the [**Files > FTP** page](/docs/panel/caching/varnish/) to manage FTP accounts.

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
* Users can access the [**Webmail** page](/docs/panel/emails/webmail/).

When disabled:
* Users can not create and manage Email accounts.

Customize options:
* To **configure email server** refer to [*How-to Guides > Configure Email Server](/docs/articles/user-experience/how-to-setup-email-in-openpanel/).
* To **configure email client** refer to [*How-to Guides > How to setup your email client](/docs/articles/email/how-to-setup-your-email-client/).
* To **view all email accounts on a server** use the [*OpenAdmin > Emails > Email Accounts* page](/docs/admin/emails/).
* To **set webmail domain or relay hosts** use the [*OpenAdmin > Emails > Email Settings* page](/docs/admin/emails/settings/).
* To **set up fail2ban** refer to [*How-to Guides > Setup Fail2ban](/docs/articles/email/how-to-setup-fail2ban-mailserver-openpanel/).
* To **set up Rspamd** refer to [*How-to Guides > RSPAMD GUI](/docs/articles/email/rspamd-gui-port-11334/).
* To **set up DKIM for a domain** refer to [*How-to Guides > Setup DKIM](/docs/articles/email/how-to-setup-dkim-for-mailserver/).
* To **limit number of email accounts per user** edit the email accounts limit when creating/editing hosting packages.


## MySQL

The **`mysql`** module allows users to create and manage mysql databases.

When enabled:
* MySQL/MariaDB auto-starts when user accesses Databases section, opens phpMyAdmin or installs WordPress.
* Users can access the [**MySQL > Databases** page](/docs/panel/mysql/databases/) to manage databases.
* Users can access the [**MySQL > New Database** page](/docs/panel/mysql/new_db/) to create databases.
* Users can access the [**MySQL > Database Wizard** page](/docs/panel/mysql/wizard/) to create database, user and assign privileges.
* Users can access the [**MySQL > Root Password** page](#) to change root user password.
* Users can access the [**MySQL > Process List** page](/docs/panel/mysql/processlist/) to view all active processes.
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

How-to guides:
* To **connect to a database** refer to [*How-to Guides > Connecting to MySQL Server from Applications in OpenPanel](/docs/articles/databases/how-to-connect-to-mysql-from-php-applications-in-openpanel/).
* To **troubleshoot errors** refer to [*How-to Guides > How to troubleshoot: Error establishing a database connection](/docs/articles/databases/how-to-troubleshoot-error-establishing-a-database-connection/).

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


## MySQL Import

The **`mysql_import`** module allows users to import files into their databases.

When enabled:
* Users can access the [**MySQL > Import Database** page](/docs/panel/mysql/import/) to import files into a database.

When disabled:
* Users can not access the *MySQL > Import Database* page.

Customize options:
* None

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








## Process Manager

The **`process_manager`** module allows users to view and terminate processes from all running services.

When enabled:
* Users can access the [**Advanced > Process Manager** page](/docs/panel/advanced/process_manager/).

When disabled:
* Users can not access the *Advanced > Process Manager* page.

Customize options:
* None



## 2FA

The **`twofa`** module allows users to enable 2 factor authentication for their account.

When enabled:
* Users can access the [**Account > Two-Factor Authentication** page](/docs/panel/account/2fa).

When disabled:
* Users can not access the *Advanced > Two-Factor Authentication* page nor manage 2FA.

Customize options:
* To **enable 2FA widget** use [*OpenAdmin > Settings > OpenPanel* page and *Display 2FA widget* option](/docs/admin/settings/openpanel/).
* To **check 2FA status for a user** refer to [How to check if 2FA is active for OpenPanel user account?](https://community.openpanel.org/d/220-how-to-check-if-2fa-is-active-for-openpanel-user-account).

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

