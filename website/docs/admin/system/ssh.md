---
sidebar_position: 3
---

# SSH Access

*OpenAdmin > System > SSH Access* allows Administrators to view and modify current server SSH configuration. The current SSH service status (active/inactive) is shown in the top-right of the page.

### Basic SSH Settings

![SSH Access page on the Basic tab with the SSH port, root login, password and public key authentication settings](/img/openadmin-screenshots/advanced/ssh-basic.png#gh-light-mode-only)
![SSH Access page on the Basic tab with the SSH port, root login, password and public key authentication settings](/img/openadmin-screenshots/advanced/ssh-basic_dark.png#gh-dark-mode-only)

This tab displays:

- **SSH Port** - current SSH port
- **Permit Root Login** - allow login for *root* user (`PermitRootLogin`)
- **Password Authentication** - enable usage of passwords for ssh (`PasswordAuthentication`)
- **Public Key Authentication** - enable usage of ssh keys (`PubkeyAuthentication`)

You can change any value and click on the save button to apply.

### Authorized Keys

This tab is only shown when **Public Key Authentication** is enabled. Here you can view current authorized ssh keys, remove them, or add a new key.

![SSH Access Authorized Keys tab listing an authorized public key with its Remove button and the field to add a new key](/img/openadmin-screenshots/advanced/ssh_server-keys.png#gh-light-mode-only)
![SSH Access Authorized Keys tab listing an authorized public key with its Remove button and the field to add a new key](/img/openadmin-screenshots/advanced/ssh_server-keys_dark.png#gh-dark-mode-only)

### Advanced

Here you can edit the raw SSH configuration file: `/etc/ssh/sshd_config`

![SSH Access Advanced tab with the sshd configuration editor](/img/openadmin-screenshots/advanced/ssh-advanced.png#gh-light-mode-only)
![SSH Access Advanced tab with the sshd configuration editor](/img/openadmin-screenshots/advanced/ssh-advanced_dark.png#gh-dark-mode-only)
