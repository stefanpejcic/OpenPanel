---
sidebar_position: 4
---

# Update Settings


The Update Settings page in OpenPanel allows you to manage and control how updates are applied. It is organized into three sections: Current Version, Auto Updates, and Update Logs.

## Current Version
This section displays information about the version of OpenPanel currently installed on your system, as well as the latest available version.
You can manually check for updates and install them if a newer version is available.

Installed Version: Shows the version of OpenPanel currently running.

Latest Version: Shows the most recent version available for update.

If the installed version matches the latest version, OpenPanel is fully up to date. If not, you can initiate a manual update to upgrade to the latest version.


## Auto Updates
The Auto Updates section lets you configure how OpenPanel should handle updates automatically.
You can select the update policy that best fits your needs:

- Minor Versions Only: Automatically apply updates that include only bug fixes and security patches. Major feature updates must be installed manually.

- Both Minor and Major Versions: Automatically apply all updates, including new features, bug fixes, and security enhancements.

- Major Versions Only: Automatically apply updates that introduce major new features. Bug fixes and security patches will require manual updates.

- Never: Disable automatic updates. You will need to manually check for and install all updates.

Choosing the appropriate auto-update setting ensures that OpenPanel remains as stable, secure, and feature-rich as you need it to be.

## Update Logs
The Update Logs section provides a record of all updates that have been applied to OpenPanel.

![openadmin set update preferences](/img/admin/openpanel_settings_updates.png)

Examples:
- Autoupdate: 1.0.2 will **NOT** be updated to 1.0.3 BUT 1.0.2 will be updated to 2.0.0
- Autopatch:  1.0.2 will be updated to 1.0.3 BUT 1.0.2 will **NOT** be updated to 2.0.0
