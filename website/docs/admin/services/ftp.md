---
sidebar_position: 5
---

# FTP Accounts

:::info
FTP management in OpenAdmin is only available on [OpenPanel Enterprise edition](/enterprise)
:::

The *Services > FTP* page lets you review the FTP sub-accounts associated with OpenPanel users, and edit the FTP server configuration. It has two tabs: **Accounts** (`/services/ftp`) and **Configuration** (`/services/ftp/settings`).

![FTP page on the Accounts tab with the FTP service status and the table of FTP accounts with their owner, path and quota](/img/openadmin-screenshots/services/ftp-accounts.png#gh-light-mode-only)
![FTP page on the Accounts tab with the FTP service status and the table of FTP accounts with their owner, path and quota](/img/openadmin-screenshots/services/ftp-accounts_dark.png#gh-dark-mode-only)

The FTP service must be running for the accounts list to be available.

The **Accounts** tab table includes the following details:
- **Account** – Username of the FTP account.
- **Owner** – The OpenPanel user that owns the FTP account, linking to that user's page. For accounts named `ftpuser@domain.com` this is the owner of that domain, the same user shown for it on the **Domains** page. If the domain no longer exists, the owner is shown as `-`.
- **Path** – The file system path to which the FTP account has access.
- **Quota** – Space used and the disk quota of the account, or **Unlimited**. Accounts over their quota are marked **Over quota**.
- **Actions** – **Set quota** opens a field next to the account: enter the quota in MB (`500`) or with a unit (`500M`, `2G`) and click **Save**. Enter `0` to remove the quota.

![FTP Accounts tab with the quota field of an account open](/img/openadmin-screenshots/services/ftp-quota.png#gh-light-mode-only)
![FTP Accounts tab with the quota field of an account open](/img/openadmin-screenshots/services/ftp-quota_dark.png#gh-dark-mode-only)

To change many accounts at once, tick their checkboxes (or the one in the table header) and use the bar at the bottom of the page: **Set quota** sets the same quota on all of them, **Remove quota** makes them unlimited.

![Two FTP accounts selected with the bulk actions bar offering Set quota and Remove quota](/img/openadmin-screenshots/services/ftp-bulk.png#gh-light-mode-only)
![Two FTP accounts selected with the bulk actions bar offering Set quota and Remove quota](/img/openadmin-screenshots/services/ftp-bulk_dark.png#gh-dark-mode-only)

The quota is the size of the account's folder, checked by the FTP service every 60 seconds. Once the folder is over the quota, uploads and new folders are refused and the account's open sessions are closed, downloads and deletes still work. Quotas can also be set with [`opencli ftp-quota`](/docs/articles/opencli/ftp/#quota).

Accounts themselves are created and removed from the OpenPanel user interface, not from here. The data shown is populated by the `opencli ftp-users` command; if no accounts are shown yet, click **Click to refresh data** to run it manually, or wait for its periodic cronjob run.

The **Configuration** tab exposes the raw FTP server configuration file in an editable text area. Edit the file contents and click **Save Configuration** to apply changes.

![FTP Configuration tab with the FTP server settings](/img/openadmin-screenshots/services/ftp-configuration.png#gh-light-mode-only)
![FTP Configuration tab with the FTP server settings](/img/openadmin-screenshots/services/ftp-configuration_dark.png#gh-dark-mode-only)
