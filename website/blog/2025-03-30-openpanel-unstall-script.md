---
title: Introducing the OpenPanel Uninstall Script
author: OpenPanel Team
author_url: https://openpanel.com
author_image_url: https://openpanel.com/static/img/logo.png
tags: [OpenPanel, uninstall, script, updates]
description: Learn about the new `unstall.sh` script for easily removing OpenPanel and its components.
date: 2025-03-30
---

We are excited to announce the release of the `unstall.sh` script, a simple and efficient way to completely remove OpenPanel and its components from your system.

### Why Do You Need `unstall.sh`?

While OpenPanel provides a seamless installation experience, there may be situations where you need to uninstall it, such as:
- Migrating to a new server.
- Reinstalling OpenPanel for troubleshooting.
- Switching to a different hosting control panel.

The `unstall.sh` script automates the uninstallation process, ensuring that all OpenPanel components, services, and configurations are removed cleanly.

---

### What Does `unstall.sh` Do?

The script performs the following actions:
1. Stops and disables OpenPanel services.
2. Removes Docker containers and volumes used by OpenPanel.
3. Deletes OpenPanel configuration files and directories.
4. Removes firewall rules (CSF or UFW) configured during installation.
5. Deletes the swap file created during installation (if applicable).
6. Removes cron jobs and logrotate configurations.
7. Restores `/etc/fstab` to remove disk quotas.
8. Uninstalls packages installed during the OpenPanel setup.

---

### How to Use `unstall.sh`

1. Download the script from the [OpenPanel GitHub repository](https://github.com/stefanpejcic/OpenPanel).
2. Make the script executable:
   ```bash
   chmod +x unstall.sh

   run sh unstall.sh or  ./unstall.sh
