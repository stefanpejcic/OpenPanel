---
sidebar_position: 4
---

# Activity Log

OpenAdmin records every change made by Administrators and Resellers in a separate activity log for each account. Only requests that change something are recorded, opening or browsing pages is not.

Each entry has the date and time, the IP address it was made from and a short description of the action, for example:

```
2026-09-28 08:02:51  203.0.113.10  Added domain 'shop.bakery.rs' for user 'john'
```

- Actions that failed end with `(failed)`, for example a domain that could not be added.
- Actions made through the [OpenAdmin API](/docs/articles/dev-experience/openadmin-api) are recorded in the log of the administrator the API token belongs to and end with `(via API)`.
- Passwords and other values typed into forms are never written to the log.
- Logins are not part of the activity log, they are shown in the **Activity** column and kept in `/var/log/openpanel/admin/login.log`.

## Viewing the Activity Log

![Activity Log page of an administrator with the date, IP address and action of each change](/img/openadmin-screenshots/accounts/administrators-activity.png#gh-light-mode-only)
![Activity Log page of an administrator with the date, IP address and action of each change](/img/openadmin-screenshots/accounts/administrators-activity_dark.png#gh-dark-mode-only)

The **Activity** column on the [Administrators](/docs/admin/accounts/administrators) and [Resellers](/docs/admin/accounts/resellers) pages shows the last action of each account, with the IP address and time of their last login below it. Click the action, or pick **Activity Log** from the Edit menu, to open the full log:

- `/administrators/activity/<username>` for administrators
- `/resellers/activity/<username>` for resellers

The newest actions are shown first, 50 per page. Use the search to filter by action, IP address or date, and the arrows next to the column names to sort.

| Role | Can view |
|---|---|
| **Super Admin** | The activity log of every account. |
| **Admin** | The activity log of every account. |
| **Reseller** | Only their own, from **My Activity Log** on their [Reseller Account](/docs/admin/accounts/resellers#reseller-account) page. |

## Log files

The logs are stored in `/var/log/openpanel/admin/activity/<username>.log`, one line per action. Once a log grows past 2 MB, the oldest entries are removed so that about the newest 1 MB is kept.

When an account is renamed with `opencli admin rename` (or from OpenAdmin), its log file is renamed too, and when an account is deleted with `opencli admin delete` its log file is deleted.

## Actions recorded for Administrators

Administrators can use every page in OpenAdmin, so every change below can appear in their log. Values in `<>` are replaced with the actual names.

**Administrators**

- `Created administrator '<username>'`
- `Renamed administrator '<username>' to '<new name>'`
- `Changed password for administrator '<username>'`
- `Suspended / Unsuspended administrator '<username>'`
- `Deleted administrator '<username>'`
- `Disabled 2FA / Removed passkeys for administrator '<username>'`

**Resellers**

- `Created reseller '<username>'`
- `Renamed reseller '<username>' to '<new name>'`
- `Changed password for reseller '<username>'`
- `Updated limits for reseller '<username>'`
- `Updated branding for reseller '<username>'`
- `Suspended / Unsuspended / Deleted reseller '<username>'`
- `Disabled 2FA / Removed passkeys for reseller '<username>'`

**Own account**

- `Disabled 2FA for own account`
- `Enabled 2FA for own account`
- `Deleted a passkey`
- `Added a passkey`
- `Renamed a passkey`

**Users**

- `Start / Stop / Restart container '<name>' of user '<username>'`
- `Edited custom message for user '<username>'`
- `Edited notes for user '<username>'`
- `Started an account transfer`
- `Imported an account from <panel>`
- `Changed default PHP version for user '<username>'`
- `Exported user '<username>'`
- `Deleted an export of user '<username>'`
- `Created user '<username>' on plan '<plan>'`
- `Suspended / Unsuspended / Deleted user '<username>'`
- `Edited user '<username>'`
- `Changed / Reset feature permissions for user '<username>'`
- `Changed <setting> for user '<username>'`

**Plans & Features**

- `Updated feature sets`
- `Updated feature set '<plan>'`
- `Deleted plan '<plan>'`
- `Updated plans`
- `Created plan '<plan>'`
- `Edited plan '<plan>'`

**Domains**

- `Started bulk add of <count> domains: <domain>, <domain>`
- `Added domain '<domain>' for user '<username>'`
- `Changed DNS cluster settings`
- `Edited domain file templates`
- `Edited VirtualHost of domain '<domain>'`
- `Edited DNS zone templates`
- `Turned WAF / HSTS on or off for domain '<domain>'`
- `Changed DNS status for domain '<domain>'`
- `Edited DNS zone of domain '<domain>'`
- `Edited Caddyfile of domain '<domain>'`
- `Edited webserver config of user '<username>'`
- `Changed SSL of domain '<domain>'`

**Emails**

- `Changed email accounts`
- `Deleted an email account`
- `Removed quota of an email account`
- `Set quota of an email account`
- `Changed sending/receiving restrictions of an email account`
- `Changed password of an email account`
- `Changed email rate limits`
- `Edited email rate limits`
- `Changed the mail queue`
- `Mail queue action <action>`
- `Changed email settings`

**Services**

- `Emptied a crash log`
- `Emptied a log file`
- `Emptied update log`
- `Start / Stop / Restart service '<name>'`
- `Changed services`
- `Changed crash logs`
- `Edited a crash log`
- `Edited a service`
- `Changed FTP accounts`
- `Refreshed FTP accounts`
- `Changed FTP settings`
- `Changed service resource limits`
- `Changed log settings`
- `Edited a log file`
- `Container images: <action>`
- `Pull / Delete container image '<image>'`
- `Edited update log`

**Backups**

- `Changed system backup settings`
- `Deleted system backup '<file>'`
- `Restored system backup '<file>'`
- `Started a system backup`
- `Changed user backup configuration`
- `Started a user backup`
- `Changed user backup settings`

**Server**

- `Changed scheduled actions`
- `Changed demo mode`
- `Dropped memory caches`
- `Cleared swap`
- `Started a server migration`
- `Changed cluster nodes`
- `Kill process <pid>`
- `Rebooted the server`
- `Changed root password`
- `Changed SSH access`
- `Edited SSH configuration`
- `Swap: <action>`
- `Changed server timezone`

**Security**

- `Changed ConfigServer Firewall`
- `Changed ImunifyAV`
- `Changed Basic Authentication`
- `Edited blocked user agents`
- `Disabled OpenAdmin`
- `Changed WAF settings`
- `Changed WAF rules`

**Notifications**

- `Deleted a notification`
- `Marked a notification as read`
- `Paused notifications`
- `Resumed notifications`
- `Snoozed a notification`
- `Unsnoozed a notification`
- `Changed notification settings`

**Settings**

- `Removed license key`
- `Deleted a default file for new users`
- `Changed license key`
- `Verified license`
- `Changed PHP <setting>`
- `Changed API access`
- `Changed Caddy settings`
- `Edited custom code`
- `Changed default settings for new users`
- `Added a default file for new users`
- `Copied default files to user '<username>'`
- `Changed general settings`
- `Changed locales`
- `Changed enabled modules`
- `Changed OpenPanel settings`
- `Changed PHP settings`
- `Changed update settings`
- `Changed update log`
- `Started an OpenPanel update`
- `Edited a default file for new users`

**Bulk actions**

- `Bulk <action> on <page>: <item>, <item>` - one line for the whole run, listing the selected rows. For example `Bulk Suspend on users: john, mary`.

Any other change that is not in this list is recorded as the request method and path, for example `POST /settings/example`.

## Actions recorded for Resellers

Resellers can only use the pages for their own users, plans and emails, so these are the actions that can appear in their log:

**Own reseller account**

- `Changed password for reseller '<username>'`
- `Updated branding for reseller '<username>'`

**Own account**

- `Disabled 2FA for own account`
- `Enabled 2FA for own account`
- `Deleted a passkey`
- `Added a passkey`
- `Renamed a passkey`

**Users**

- `Start / Stop / Restart container '<name>' of user '<username>'`
- `Edited custom message for user '<username>'`
- `Edited notes for user '<username>'`
- `Changed default PHP version for user '<username>'`
- `Exported user '<username>'`
- `Deleted an export of user '<username>'`
- `Created user '<username>' on plan '<plan>'`
- `Suspended / Unsuspended / Deleted user '<username>'`
- `Edited user '<username>'`
- `Changed / Reset feature permissions for user '<username>'`
- `Changed <setting> for user '<username>'`

**Plans & Features**

- `Updated feature sets`
- `Updated feature set '<plan>'`
- `Deleted plan '<plan>'`
- `Updated plans`
- `Created plan '<plan>'`
- `Edited plan '<plan>'`

**Emails**

- `Changed email accounts`
- `Deleted an email account`
- `Removed quota of an email account`
- `Set quota of an email account`
- `Changed sending/receiving restrictions of an email account`
- `Changed password of an email account`
- `Changed email rate limits`
- `Edited email rate limits`
- `Changed the mail queue`
- `Mail queue action <action>`
- `Changed email settings`

**Bulk actions**

- `Bulk <action> on <page>: <item>, <item>` - on the Users, Plans, Emails and container pages they have access to.
