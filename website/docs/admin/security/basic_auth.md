---
sidebar_position: 4
---

# Basic Auth

Enable [Basic Access Authentication](https://en.wikipedia.org/wiki/Basic_access_authentication) as an additional security layer for the OpenAdmin interface.

![Basic Authentication page with the Enable dropdown and the username and password fields](/img/openadmin-screenshots/security/basic_auth-page.png#gh-light-mode-only)
![Basic Authentication page with the Enable dropdown and the username and password fields](/img/openadmin-screenshots/security/basic_auth-page_dark.png#gh-dark-mode-only)

## Enable

To enable Basic Auth, select **Yes** under **Enable** and click **Save settings**.

To disable Basic Auth, select **No** under **Enable** and click **Save settings**.

While Basic Auth is off and the OpenAdmin port is open to everyone in the [firewall](/docs/admin/security/firewall) (listed in CSF's `TCP_IN`, or there's no CSF), enabling it is recommended as an extra login step in front of OpenAdmin.

## Logins

When Basic Auth is enabled, a username and password are required.

Configure the **Username** and **Password** fields, then click **Save settings** to apply. The password must be 6-30 characters and cannot contain an apostrophe (`'`).

Use the **Generate** button next to the password field to create a random strong password, and the eye icon to toggle its visibility.
