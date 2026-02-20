---
sidebar_position: 2
---

# Feature Manager

The Feature Manager allows administrators to enable or disable specific features (pages) within the OpenPanel UI. This is useful for customizing the control panel experience based on user roles, security policies, or hosting plans.

Each feature can be toggled individually via the interface. Once activated, the feature becomes visible and accessible to all users. Deactivating a feature hides it from the panel and disables its functionality.

![openadmin features](/img/admin/tremor/features.png)

![openadmin features](/img/admin/tremor/features_edit.png)

---

## Available Features
Here is the rewritten information in a table format with the requested columns:

| **Name**                   | **Link**                 | **Description**                                          | **Note**                                       |
| -------------------------- | ------------------------ | -------------------------------------------------------- | ---------------------------------------------- |
| Email Notifications        | `/account/notifications` | Manage email notification preferences.                   | Emails are sent based on selected preferences. |
| Locales (Language Change)  | `/account/languages`     | Change the panel interface language.                     |                                                |
| Favorites (Bookmarks)      | `/account/favorites`     | Bookmark frequently used pages.                          |                                                |
| Varnish Caching            | `/cache/varnish`         | Manage Varnish caching per domain.                       |                                                |
| Docker (Containers)        | `/containers`            | Allocate resources and manage container lifecycles.      |                                                |
| FTP Accounts               | `/ftp`                   | Create and manage FTP accounts.                          | Requires separate FTP server configuration.    |
| Email Accounts             | `/emails`                | Manage email accounts.                                   | Requires separate mail server configuration.   |
| Remote MySQL               | `/mysql/remote-mysql`    | Allow or block remote MySQL connections.                 |                                                |
| MySQL Configuration        | `/mysql/configuration`   | Modify MySQL settings from the panel.                    |                                                |
| PHP Options                | `/php/options`           | Edit PHP directives via a user-friendly page.            |                                                |
| PHP.INI Editor             | `/php/ini`               | Directly edit the `php.ini` file.                        | Applies to any configured PHP version.         |
| phpMyAdmin                 | `/mysql/phpmyadmin`      | Manage databases with phpMyAdmin.                        |                                                |
| Cronjobs                   | `/cronjobs`              | Create, edit, and schedule cron jobs.                    |                                                |
| WordPress                  | `/wordpress`             | Install and manage WordPress sites.                      | Managed via WP Manager.                        |
| Disk Usage Explorer        | `/disk-usage`            | Visually explore disk usage across directories.          |                                                |
| Inodes Explorer            | `/inodes-explorer`       | View inode usage per directory.                          |                                                |
| Resources Usage            | `/usage`                 | View Docker container resource usage.                    |                                                |
| Server Info                | `/server/info`           | View hosting limits and server details.                  |                                                |
| Apache/Nginx Configuration | `/server/webserver_conf` | Modify webserver (Apache/Nginx) container configuration. |                                                |
| Change Timezone            | `/server/timezone`       | Update system timezone settings for containers.          |                                                |
| Coraza WAF                 | `/waf`                   | Manage Coraza WAF per domain.                            | Enabled by default for new domains.            |
| Fix Permissions            | `/fix-permissions`       | Fix file ownership and permissions for websites.         |                                                |
| DNS                        | `/dns`                   | Manage DNS records with a zone editor.                   | Requires BIND9 server.                         |
| Domain Redirects           | `/domains/redirects`     | Manage domain-level redirects.                           |                                                |
| Malware Scanner            | `/malware-scanner`       | Scan for malware using ClamAV.                           | Directory exclusions can be configured.        |
| GoAccess                   | `/domains/logs`          | View GoAccess-generated log reports.                     |                                                |
| Process Manager            | `/process-manager`       | View and terminate system processes.                     |                                                |
| Redis                      | `/cache/redis`           | Configure Redis per user.                                |                                                |
| Memcached                  | `/cache/memcached`       | Configure Memcached per user.                            |                                                |
| Elasticsearch              | `/cache/elasticsearch`   | Configure Elasticsearch from the panel.                  |                                                |
| Opensearch                 | `/cache/opensearch`      | Configure Opensearch from the panel.                     |                                                |
| Temporary Links            | `/websites`              | Test websites using temporary OpenPanel subdomains.      | Links expire after 15 minutes.                 |
| Login History              | `/account/loginlog`      | View history of the last 20 IP logins.                   |                                                |
| 2FA                        | `/account/2fa`           | Enable Two-Factor Authentication.                        |                                                |
| Activity Log               | `/account/activity`      | Review all recorded account actions.                     |                                                |

## Use Cases

**Feature Sets** are used to control which UI features users can access based on their assigned hosting package. This allows for clear separation between user types and service levels.

### Example 1: Database-Only Plans

Create a feature set named **"MySQL Only"** and enable only MySQL-related features within it.
Assign this feature set to all database-focused hosting packages. For instance:

* One package allows up to **10 databases**.
* Another package allows **unlimited databases** (`0` for no limit).

Despite the difference in limits, all users under these plans will see **only MySQL-related pages** in the UI.

### Example 2: Beginner vs. Advanced Users

Create two separate feature sets:

* **Advanced Users Set**:
  Enable features like **Docker** and **PHP.INI Editor** to give experienced users full control—such as setting custom resource limits, restarting services, etc.

* **Beginner Users Set**:
  Do **not** enable advanced features. Instead, allow access to a **PHP selector** with limited options. This keeps the UI simple and safe for users with minimal technical experience.


## Feature not showing?

Features are accessible to users [only if the corresponding **Module** is active](/docs/admin/settings/modules/). Modules control which OpenPanel features are available, while **Feature Sets** determine access based on the user's hosting package.

For example, adding the "Docker" feature to a plan does **not** grant access to the Docker (Containers) pages in the UI unless the **Docker module** is also activated under **OpenAdmin > Settings > Modules**.

