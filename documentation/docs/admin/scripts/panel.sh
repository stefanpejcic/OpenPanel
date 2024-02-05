---
sidebar_position: 9
---

# System

Scripts for checking OpenPanel version, running updates, etc.

## Check version

To check installed openpanel version run the following command:

```bash
bash /usr/local/admin/scripts/users/version.sh
```

## update_check

To check if update is available run:

```bash
bash /usr/local/admin/scripts/users/update_check.sh
```


## ate

This script is used by OpenPanel to check if [autoupdate](/docs/admin/scripts/openpanel_config#autoupdate) or [autopatch](/docs/admin/scripts/openpanel_config#autopatch) options are enabled by user and then run update:

```bash
bash /usr/local/admin/scripts/users/update.sh
```

To manually update OpenPanel regardless of the `autoupdate` and `autopatch` settings, use  `--force` flag.
