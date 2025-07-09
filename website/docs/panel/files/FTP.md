---
sidebar_position: 2
---

# FTP

OpenPanel accounts can create FTP users, which share the FTP service with all other FTP users on the server. You can connect to FTP using the server IP address and the default port, `21`.

:::info
FTP is available in the [OpenPanel Enterprise](/enterprise/) edition, guide for Administrators on [How to setup FTP in OpenPanel](/docs/articles/user-experience/how-to-setup-ftp-in-openpanel/)
:::

## View users

To view FTP connection information and manage existing users—such as changing passwords, modifying their FTP path (the directory they are limited to), or deleting accounts-navigate to `OpenPanel > Files > FTP`.

## Create user

To create a new FTP sub-user, go to `OpenPanel > Files > FTP` from the menu and click the **Add Account** button. A new section will appear, where you can set the FTP username, password, and path.

Click the **Add Account** button to save your changes.

:::info
FTP sub-users **must end** with dot (`.`) followed by the OpenPanel username - example: `ftpuser.openpaneluser`. [FTP username requirements](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Change password

To change password for an FTP sub-account, go to `OpenPanel > Files > FTP` from the menu and click the **Change Password** button located next to the user. On the next page set the new password for the user.

Click the **Change Password** button to save.

:::info
FTP User's Passwords must contain **at least one** uppercase letters (`A-Z`), lowercase letters (`a–z`), digits (`0–9`) and special symbols. [FTP passwords requirements](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Change path

To change path (directory user is restricted to) for an FTP sub-account, go to `OpenPanel > Files > FTP` from the menu and click the **Change Path** button located next to the user. On the next page set new path for the user.

Click the **Change Path** button to save.

## Delete user

To delete an FTP sub-account, go to `OpenPanel > Files > FTP` from the menu and click the **Delete** button located next to the user. The button will swich to text 'Confirm', click it again to confirm the deletion.

## View Connections

To view active connections (sessions), go to `OpenPanel > Files > FTP` from the menu, then click on 'View Connections' link on top right of the page.
