---
sidebar_position: 7
---

# Activity Log

The Account Activity page provides a log record of all your actions performed on the OpenPanel, along with the timestamp and IP address from which the action was executed. The primary objective of this page is to offer insights into who carried out specific actions, such as deleting a file, adding domains, resetting WordPress admin passwords, and more.

![account_activity_log.png](/img/panel/v1/analytics/account_activity_log.png)


Logged Activities Include:

- Account: Logging in and logging out.
- Domains: Adding and deleting domains.
- File Manager: Deleting, moving, copying, and creating files.
- Services: Enabling or disabling 2FA, Memcached, REDIS, PHP, and more.
- WordPress: Installing, deleting, changing passwords, auto-login, and more.

You can search for these actions by IP, username, date, or specific action.

By default only 25 actions will be shown per page, you can navigate pages using the pagination links or click on 'Show all activity' to display all actions.

:::info
The default configuration retains a log of the most recent 100 actions, but [the administrator can change this setting](/docs/admin/scripts/openpanel_config#activity_items_per_page).
:::
