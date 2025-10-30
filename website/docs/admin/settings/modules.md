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
* To **edit the default.vcl file for Varnish** use the [*OpenAdmin > Doamins > Edit Domain Templates* page](/docs/admin/settings/defaults/) or edit file: [`/etc/openpanel/varnish/default.vcl`](https://github.com/stefanpejcic/openpanel-configuration/blob/main/varnish/default.vcl).
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

