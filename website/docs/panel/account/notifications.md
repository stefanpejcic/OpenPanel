---
sidebar_position: 3
---

# Notifications

Users can set actions for which to receive email notifications:

![notifications.png](/img/panel/v1/account/notifications.png)

:::info
If you do not see the Notifications page, ask your provider to enable the [notifications module](/docs/admin/settings/openpanel/#notifications).
:::

| Value                                          | Description                                  | Default |
|------------------------------------------------|----------------------------------------------|---------|
| notify_account_login                           | Notify on new logins                         | 1       |
| notify_account_login_for_known_netblock        | Notify on new login from known netblock      | 1       |
| notify_account_login_notification_disabled     | Notify when login notifications are disabled | 1       |
| notify_autossl_expiry                          | Notify when SSL certificate is soon expiring | 0       |
| notify_autossl_expiry_coverage                 | Notify when SSL certificate is expired       | 0       |
| notify_autossl_renewal_coverage                | Notify when certificate renewal failed   | 0       |
| notify_autossl_renewal_coverage_reduced        | Notify when certificate renewal failed   | 0       |
| notify_autossl_renewal_uncovered_domains       | Notify when certificate renewal failes for custom SSL| 0   |
| notify_contact_address_change                  | Notify when email address is changed             | 1       |
| notify_contact_address_change_notification_disabled | Notify when notifications for email address change are disabled | 1   |
| notify_disk_limit                              | Notify when account reached disk limit                         | 1       |
| notify_email_quota_limit                       | Notify when email accounts reach their quota limit                  | 1       |
| notify_password_change                         | Notify when password is changed                    | 1       |
| notify_password_change_notification_disabled   | Notify when notifications for password change are disabled      | 1       |
| notify_ssl_expiry                              | Notify when SSL certificate is expired                      | 1       |
| notify_twofactorauth_change                    | Notify when two-factor authentication is disabled/enabled   | 1       |
| notify_twofactorauth_change_notification_disabled | Notify when notifications for two-factor authenticatin change are disabled | 1 |


User does not receive notifications for actions performed through the OpenAdmin interface (such as impersonating a user, changing a password, disabling 2FA, etc.) or from API calls (e.g., from WHMCS, Blesta, FOSSBilling).


## Email notifications

![new_login.png](/img/panel/v1/account/new_login.png)

![2fa_disabled.png](/img/panel/v1/account/2fa_disabled.png)

![2fa_enabled.png](/img/panel/v1/account/2fa_enabled.png)

![email_changed.png](/img/panel/v1/account/email_changed.png)

![pass_changed.png](/img/panel/v1/account/pass_changed.png)

![preferences_changed.png](/img/panel/v1/account/preferences_changed.png)
