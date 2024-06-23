---
sidebar_position: 3
---

# SSH Access

*OpenAdmin > Server >  SSH Access* allows Administrators to view and modify current server SSH configuration.

### Basic SSH Settings

This page displays:

- **Port** - current SSH port
- **PermitRootLogin** - allow login for *root* user
- **PasswordAuthentication** - enable usage of passwords for ssh
- **PubkeyAuthentication** - enable usage of ssh keys

You can change any value and click on the save button to apply.

### Advanced SSH Settings

Here you can edit the SSH configuration file: `/etc/ssh/sshd_config`

### Authorized SSH Keys

Here you can view current authorized ssh keys, remove them or add new key.
