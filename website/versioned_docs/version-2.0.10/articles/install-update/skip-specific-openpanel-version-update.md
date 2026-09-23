# Skip Update

OpenPanel can be updated either **automatically** or **manually**, depending on your preferences.

The update process performs the following actions:

* Pulls the latest OpenPanel UI Docker image from Docker Hub
* Updates the OpenAdmin UI from the latest GitHub release
* Updates OpenCLI commands from GitHub
* Executes any update scripts provided for configuration changes
* Checks and installs updates via the system package manager
* Updates docker compose to latest version
* Verifies if any kernel updates require a reboot
* Executes post-update scripts if provided by the administrator

If both **autoupdate** and **autopatch** options are enabled, updates will be applied automatically when a new version is released.

However, you can **skip specific versions** by listing them in the following file:

```
/etc/openpanel/upgrade/skip_versions
```

### When to Use This

This is useful, for example, if a changelog announces the removal of a feature you rely on. By skipping that version, you can ensure your current setup remains unaffected.

### How to Skip a Version

1. Create the directory (if it doesn't already exist):

```bash
mkdir -p /etc/openpanel/upgrade/
```

2. Add the version number you want to skip:

```bash
echo 1.4.9 >> /etc/openpanel/upgrade/skip_versions
```

Now, when version `1.4.9` is released, it will be ignored even if autoupdates are enabled.

You can add as many versions as needed:

```bash
echo 1.5.0 >> /etc/openpanel/upgrade/skip_versions
echo 1.5.1 >> /etc/openpanel/upgrade/skip_versions
```

That's it - a simple and effective control over your OpenPanel update process.
