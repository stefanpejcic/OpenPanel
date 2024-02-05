---
sidebar_position: 2
---

# SSH

The SSH Access page provides the capability to check the current status of remote SSH access, which is by default disabled. Moreover, it allows you to toggle the enable/disable setting as needed.

![ssh_enabled.png](/img/panel/v1/advanced/ssh_enabled.png)

To enable SSH access click on the 'Enable SSH Access' button.

![ssh_disabled.png](/img/panel/v1/advanced/ssh_disabled.png)

When SSH access is enabled, it permits SSH connections from any location to the random SSH port unique to your account.

To modify the SSH password for your account, click on the 'Change SSH Password' button, and within the modal, set a new password.

This alteration will promptly update the password; however, all existing ssh connections will remain active.

To terminate existing SSH connections, use the [Process Manager](/docs/panel/advanced/process_manager) interface and kill the appropriate processes.
