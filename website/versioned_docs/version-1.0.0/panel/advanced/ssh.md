---
sidebar_position: 2
---

# SSH

The SSH Access page provides the capability to check the current status of remote SSH access, which is by default disabled. Moreover, it allows you to toggle the enable/disable setting as needed.

![ssh_enabled.png](/img/panel/v1/advanced/ssh_enabled.png)

To enable SSH access click on the 'Enable SSH Access' button.

![ssh_disabled.png](/img/panel/v1/advanced/ssh_disabled.png)

When SSH access is enabled, it permits SSH connections from any location to the random SSH port unique to your account.

## Changing Your SSH Password from terminal

1. Log in to your server over SSH.
2. Enter the command:
   ```shell
   passwd
   ```
3. Type your password, then press Enter.
4. When prompted for your current password, enter your SSH password, then press Enter.
5. Retype your new password, then press Enter. If successful, you will see the output:
   ```shell
   passwd: all authentication tokens updated successfully
   ```


## Changing Your SSH Password from WebTerminal


1. Navigate to *OpenPanel > WebTerminal*
2. Enter the command:
   ```shell
   passwd
   ```
3. Type your password, then press Enter.
4. When prompted for your current password, enter your SSH password, then press Enter.
5. Retype your new password, then press Enter. If successful, you will see the output:
   ```shell
   passwd: all authentication tokens updated successfully
   ```
   
## Changing Your SSH Password from OpenPanel interface

To modify the SSH password for your account from the OpenPanel interface, navigate to *SSH* and click on the 'Change SSH Password' button. Set the new password in the modal and confirm.

The password will immediately be changed; however, all existing ssh connections will remain active.

To terminate existing SSH connections, use the [Process Manager](/docs/panel/advanced/process_manager) interface and kill the appropriate processes.
