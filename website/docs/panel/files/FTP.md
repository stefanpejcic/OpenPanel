---
sidebar_position: 3
---

# FTP

Users can create FTP sub-accounts, which share the FTP service among all users on the server. You can connect to FTP using the server IP address and the default port, `21`.

:::info
By default, FTP is **not enabled** in OpenPanel. If you don’t see the option in your OpenPanel interface, it means the server administrator has not enabled it]. You’ll need to request them to [enable FTP module](/docs/admin/settings/openpanel/#ftp) before you can use it.
:::

## View users

To view FTP connection information and manage existing users—such as changing passwords, modifying their FTP path (the directory they are limited to), or deleting accounts—navigate to `OpenPanel > Files > FTP`.

## Create user

To create a new FTP sub-account, go to `OpenPanel > Files > FTP` from the menu and click the **Add Account** button located in the bottom right corner of the screen. A new section will appear above the existing users table, where you can set the FTP username, password, and path.

![add ftp user](/img/panel/v1/files/add_ftp_acc.png)

Click the **Add Account** button to save your changes.

:::info
FTP sub-users **must end** with dot (`.`) followed by the OpenPanel username - example: `ftpuser.openpaneluser`. [FTP username requirements](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Change password

To change password for an FTP sub-account, go to `OpenPanel > Files > FTP` from the menu and click the **Change Password** button located next to the user. A modal will appear in which you can set new password for the user.

![change ftp user password](/img/panel/v1/files/change_ftp_pass.png)

Click the **Change Password** button to save.

:::info
FTP User's Passwords must contain **at least one** uppercase letters (`A-Z`), lowercase letters (`a–z`), digits (`0–9`) and special symbols. [FTP passwords requirements](/docs/articles/accounts/forbidden-usernames/#ftp)
:::

## Change path

To change path (directory user is restricted to) for an FTP sub-account, go to `OpenPanel > Files > FTP` from the menu and click the **Change Path** button located next to the user. A modal will appear in which you can set new path for the user.

![change ftp user password](/img/panel/v1/files/change_ftp_pass.png)

Click the **Change Path** button to save.

## Delete user

To delete an FTP sub-account, go to `OpenPanel > Files > FTP` from the menu and click the **Delete** button located next to the user. The button will swich to text 'Confirm', click it again to confirm the deletion.

