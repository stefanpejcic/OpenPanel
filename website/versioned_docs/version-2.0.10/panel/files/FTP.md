---
sidebar_position: 2
---

# FTP Accounts

OpenPanel accounts can create FTP sub-users that share the FTP service with all other FTP users on the server. Connect using the server IP address and port `21`.

![FTP Accounts page with the FTP server and port, and a table of accounts with their path and configuration downloads](/img/openpanel-screenshots/files/ftp-list.png#gh-light-mode-only)
![FTP Accounts page with the FTP server and port, and a table of accounts with their path and configuration downloads](/img/openpanel-screenshots/files/ftp-list_dark.png#gh-dark-mode-only)

:::info
FTP is available in the [OpenPanel Enterprise](/enterprise/) edition. Administrator setup guide: [How to setup FTP in OpenPanel](/docs/articles/user-experience/how-to-setup-ftp-in-openpanel/)
:::

---

## View Users

Go to **OpenPanel > Files > FTP Accounts** to see all FTP accounts, the directory each account is restricted to, and connection details including the server IP.

## Create User

1. Go to **OpenPanel > Files > FTP Accounts** and click **New Account**.
2. Fill in the form:

   | Field | Description |
   |---|---|
   | **Username** | The local part of the FTP username: the full username will be `username@domain.com` |
   | **Domain** | Select one of your domains |
   | **Password** | Password for the FTP account |
   | **Path (folder)** | Directory the account is restricted to: must start with `/var/www/html/` |

3. Click **Create Account** to create the user.

![New FTP Account form with the username, domain, password and path fields](/img/openpanel-screenshots/files/ftp_new-form.png#gh-light-mode-only)
![New FTP Account form with the username, domain, password and path fields](/img/openpanel-screenshots/files/ftp_new-form_dark.png#gh-dark-mode-only)

:::info
FTP usernames may only contain letters, numbers, and the characters `-`, `_`, `@`, `.`. [FTP username requirements](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Change Password

1. Go to **OpenPanel > Files > FTP Accounts** and click **Change Password** next to the account.
2. Enter the new password, confirm it, and click **Update**.

![Change FTP password form with a new password field](/img/openpanel-screenshots/files/ftp_password-form.png#gh-light-mode-only)
![Change FTP password form with a new password field](/img/openpanel-screenshots/files/ftp_password-form_dark.png#gh-dark-mode-only)

:::info
FTP passwords must be 8-30 characters long and contain at least one uppercase letter (`A-Z`), lowercase letter (`a-z`), digit (`0-9`), and special character. [FTP password requirements](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Change Path

1. Go to **OpenPanel > Files > FTP Accounts** and click **Change Path** next to the account.
2. Enter the new directory path and click **Update**. The path must start with `/var/www/html/`.

![Change FTP path form with the folder field](/img/openpanel-screenshots/files/ftp_path-form.png#gh-light-mode-only)
![Change FTP path form with the folder field](/img/openpanel-screenshots/files/ftp_path-form_dark.png#gh-dark-mode-only)

## Delete User

Go to **OpenPanel > Files > FTP Accounts** and click **Delete** next to the account. The button turns into **Confirm** with a 5-second countdown; click it again before the countdown ends to delete the account.

![Delete button of an FTP account turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/files/ftp_actions-delete.png#gh-light-mode-only)
![Delete button of an FTP account turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/files/ftp_actions-delete_dark.png#gh-dark-mode-only)

## Download Client Configuration

On the FTP accounts page you can download a ready-to-import configuration file for:

- **FileZilla**: downloads an `.xml` connection file
- **Cyberduck**: downloads an `.ftpbookmark` file

Both files are pre-filled with the correct server IP, port, username, and remote directory.

![FileZilla and Cyberduck buttons in the Configuration column of an FTP account](/img/openpanel-screenshots/files/ftp_actions-config.png#gh-light-mode-only)
![FileZilla and Cyberduck buttons in the Configuration column of an FTP account](/img/openpanel-screenshots/files/ftp_actions-config_dark.png#gh-dark-mode-only)

## View Active Connections

Go to **OpenPanel > Files > FTP Accounts** and click **View Connections** to see all currently active FTP sessions for your account.

![FTP Connections page listing current FTP sessions](/img/openpanel-screenshots/files/ftp_connections-list.png#gh-light-mode-only)
![FTP Connections page listing current FTP sessions](/img/openpanel-screenshots/files/ftp_connections-list_dark.png#gh-dark-mode-only)
