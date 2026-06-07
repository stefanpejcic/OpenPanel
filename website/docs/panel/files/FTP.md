---
sidebar_position: 2
---

# FTP

OpenPanel accounts can create FTP sub-users that share the FTP service with all other FTP users on the server. Connect using the server IP address and port `21`.

:::info
FTP is available in the [OpenPanel Enterprise](/enterprise/) edition. Administrator setup guide: [How to setup FTP in OpenPanel](/docs/articles/user-experience/how-to-setup-ftp-in-openpanel/)
:::

---

## View Users

Go to **OpenPanel > Files > FTP** to see all FTP accounts, the directory each account is restricted to, and connection details including the server IP.

## Create User

1. Go to **OpenPanel > Files > FTP** and click **Add Account**.
2. Fill in the form:

   | Field | Description |
   |---|---|
   | **Domain** | Select one of your domains |
   | **Username** | The local part of the FTP username: the full username will be `username@domain.com` |
   | **Password** | Password for the FTP account |
   | **Path** | Directory the account is restricted to: must start with `/var/www/html/` |

3. Click **Add Account** to create the user.

:::info
FTP usernames may only contain letters, numbers, and the characters `-`, `_`, `@`, `.`. [FTP username requirements](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Change Password

1. Go to **OpenPanel > Files > FTP** and click **Change Password** next to the account.
2. Enter the new password and click **Change Password**.

:::info
FTP passwords must contain at least one uppercase letter (`A-Z`), lowercase letter (`a-z`), digit (`0-9`), and special character. [FTP password requirements](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Change Path

1. Go to **OpenPanel > Files > FTP** and click **Change Path** next to the account.
2. Enter the new directory path and click **Change Path**. The path must start with `/var/www/html/`.

## Delete User

Go to **OpenPanel > Files > FTP** and click **Delete** next to the account, then confirm the deletion.

## Download Client Configuration

On the FTP accounts page you can download a ready-to-import configuration file for:

- **FileZilla**: downloads an `.xml` connection file
- **Cyberduck**: downloads an `.ftpbookmark` file

Both files are pre-filled with the correct server IP, port, username, and remote directory.

## View Active Connections

Go to **OpenPanel > Files > FTP** and click **View Connections** to see all currently active FTP sessions for your account.
