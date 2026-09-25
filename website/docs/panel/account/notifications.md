---
sidebar_position: 3
---

# Notifications

Users can set actions for which to receive email notifications:

![Email Notifications page with a checkbox for each event and the Save Preferences button](/img/openpanel-screenshots/account/notifications-form.png#gh-light-mode-only)
![Email Notifications page with a checkbox for each event and the Save Preferences button](/img/openpanel-screenshots/account/notifications-form_dark.png#gh-dark-mode-only)

:::info
If you do not see the Notifications page, ask your provider to enable the [notifications module](/docs/admin/settings/modules/#notifications).
:::

| Value                                          | Description                                  | Default |
|------------------------------------------------|----------------------------------------------|---------|
| notify_account_login                           | Notify on new logins                         | 1       |
| notify_account_login_for_known_netblock        | Also notify on logins from an IP address that's already in your login history. Turn it off to only get emails for logins from new IP addresses | 1       |
| notify_account_login_notification_disabled     | Notify when login notifications are disabled | 1       |
| notify_autossl_expiry                          | Notify when SSL certificate is soon expiring | 0       |
| notify_autossl_expiry_coverage                 | Notify when SSL certificate is expired       | 0       |
| notify_autossl_renewal_coverage                | Notify when certificate renewal failed   | 0       |
| notify_autossl_renewal_coverage_reduced        | Notify when certificate renewal failed   | 0       |
| notify_autossl_renewal_uncovered_domains       | Notify when certificate renewal fails for custom SSL| 0   |
| notify_contact_address_change                  | Notify when email address is changed             | 1       |
| notify_contact_address_change_notification_disabled | Notify when notifications for email address change are disabled | 1   |
| notify_disk_limit                              | Notify when the account uses 85% of its disk space or 95% of its inodes. Sent once, and again only after usage drops below and crosses the limit again | 1       |
| notify_email_quota_limit                       | Notify when an email account uses 90% of its quota. Sent once, and again only after usage drops below and crosses the limit again | 1       |
| notify_password_change                         | Notify when password is changed                    | 1       |
| notify_password_change_notification_disabled   | Notify when notifications for password change are disabled      | 1       |
| notify_ssl_expiry                              | Notify when SSL certificate is expired                      | 1       |
| notify_twofactorauth_change                    | Notify when two-factor authentication is disabled/enabled   | 1       |
| notify_twofactorauth_change_notification_disabled | Notify when notifications for two-factor authenticatin change are disabled | 1 |


User does not receive notifications for actions performed through the OpenAdmin interface (such as impersonating a user, changing a password, disabling 2FA, etc.). Logins, password, email, username and 2FA changes made through the [OpenPanel API](/docs/panel/api/) do send the same notifications as when they're made from the panel.

When a security notification (login, password, email address or 2FA change) is turned off and its matching `_notification_disabled` option is on, an email is sent to let you know the notification was disabled.

:::note
Emails are currently sent only for logins, password, email address, username and two-factor authentication changes, disk usage and email quota limits. The SSL options are saved but not sent yet.
:::


## Email notifications

![new_login.png](/img/panel/v1/account/new_login.png)

![2fa_disabled.png](/img/panel/v1/account/2fa_disabled.png)

![2fa_enabled.png](/img/panel/v1/account/2fa_enabled.png)

![email_changed.png](/img/panel/v1/account/email_changed.png)

![pass_changed.png](/img/panel/v1/account/pass_changed.png)

![preferences_changed.png](/img/panel/v1/account/preferences_changed.png)
