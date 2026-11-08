---
sidebar_position: 8
---

# Updates

The **Settings > Updates** page lets you manage and control how OpenPanel updates are applied. It is organized into sections: Current version, Auto updates, Update logs and Rollback.

## Current version
This section displays information about the version of OpenPanel currently installed on your system, as well as the latest available version.

![Current version section with the installed and latest version and the changelog link](/img/openadmin-screenshots/settings/updates-current.png#gh-light-mode-only)
![Current version section with the installed and latest version and the changelog link](/img/openadmin-screenshots/settings/updates-current_dark.png#gh-dark-mode-only)

- **Installed version:** Shows the version of OpenPanel currently running.

- **Latest version:** Shows the most recent version available for update.

If the installed version matches the latest version, OpenPanel is fully up to date. If not, click **Update now** to upgrade to the latest version.

A notice above the sections says when the installed version needs attention, with a link to the changelog:

- **Older than the minimum supported version** (error): versions below 2.0.22 miss a lot of security improvements, update as soon as you can.
- **A major version behind** (error): older major versions stop getting fixes, plan the update.
- **Behind the latest minor version** (warning): a newer minor version is out with fixes and new features.

## Auto updates
The Auto updates section lets you configure how OpenPanel should handle updates automatically.
You can select the update policy that best fits your needs:

- **Minor versions only:** Automatically apply updates that include only bug fixes and security patches. Major feature updates must be installed manually.

- **Both minor and major versions:** Automatically apply all updates, including new features, bug fixes, and security enhancements.

- **Major versions only:** Automatically apply updates that introduce major new features. Bug fixes and security patches will require manual updates.

- **Never:** Disable automatic updates. You will need to manually check for and install all updates.

![Auto Updates section with the update preference dropdown](/img/openadmin-screenshots/settings/updates-auto.png#gh-light-mode-only)
![Auto Updates section with the update preference dropdown](/img/openadmin-screenshots/settings/updates-auto_dark.png#gh-dark-mode-only)

Choosing the appropriate auto-update setting ensures that OpenPanel remains as stable, secure, and feature-rich as you need it to be.

A recommendation is shown under the dropdown for the two options that skip security patches:

- **Never:** update manually at least once a month.
- **Major versions only:** update manually at least once a quarter, to get the minor versions with bug fixes and security patches.

Both link to [how to update manually](/docs/articles/opencli/update/), and also say how many days ago the last update ran when that's over a month (Never) or 90 days (Major versions only).

## Update logs
The Update logs section provides a record of all updates that have been applied to OpenPanel.

![Update Logs section listing the update log files](/img/openadmin-screenshots/settings/updates-logs.png#gh-light-mode-only)
![Update Logs section listing the update log files](/img/openadmin-screenshots/settings/updates-logs_dark.png#gh-dark-mode-only)

## Rollback
The Rollback section lets you revert the OpenPanel UI to an earlier version. Pick a version from the list, newest first, and click **Roll back**. OpenAdmin asks you to confirm, and turns automatic updates off so OpenPanel doesn't update again on its own.

![Rollback section for returning to a previous version](/img/openadmin-screenshots/settings/updates-rollback.png#gh-light-mode-only)
![Rollback section for returning to a previous version](/img/openadmin-screenshots/settings/updates-rollback_dark.png#gh-dark-mode-only)

## Control updates
For more information about updates, check:
- [How to Enable Automatic Updates for OpenPanel](/docs/articles/install-update/how-to-enable-auto-update-openpanel/)
- [How to Disable Automatic Updates for OpenPanel](/docs/articles/install-update/how-to-disable-auto-update-openpanel/)
- [How to Manually Update OpenPanel](/docs/articles/install-update/how-to-enable-auto-update-openpanel/)
- [How to Skip a Specific OpenPanel Version Update](/docs/articles/install-update/skip-specific-openpanel-version-update)
- [How to Check OpenPanel Version from Terminal](/docs/articles/install-update/how-to-check-openpanel-version)
