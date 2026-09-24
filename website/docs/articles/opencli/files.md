# Files

### Purge Trash

The `purge_trash` script is run periodically to remove files form users *.Trash* folder, according to the `autopurge_trash` setting.


#### Single User

To purge files for a single user:

```bash
opencli files-purge_trash --user [USERNAME]
```

To list files that would be purged (dry-run) for a single user:

```bash
opencli files-purge_trash --user [USERNAME] --dry-run
```

To purge files for a single user regardless of the `autopurge_trash` setting:

```bash
opencli files-purge_trash --user [USERNAME] --force
```

#### All Users

To purge files for all users:

```bash
opencli files-purge_trash
```

To list files that would be purged (dry-run) for all users:

```bash
opencli files-purge_trash --dry-run
```

To purge files for all users regardless of the `autopurge_trash` setting:

```bash
opencli files-purge_trash --force
```


### Fix Permissions

The `fix_permissions` script can be used to fix permissions on user files inside their container.

It performs:

Website files:
- sets the owner of all files in `/var/www/html/` to the  www-data user.
- sets the group of all files in `/var/www/html/` to the www-data group.
- sets the permissions of all files to 644.
- sets the permissions of all directories to 755.

Emails:
- sets the owner of all files in `/var/mail/` to the  user.
- sets the permissions of all files to 644.
- sets the permissions of all directories to 755.


#### Single folder

Fix permissions for specific folder for user:

```bash
opencli files-fix_permissions <USERNAME> <PATH>
```

**TIP:** add `--debug` flag to show verbose information on what is exaclty being done:
```bash
opencli files-fix_permissions <USERNAME> <PATH> --debug
```

Example:
```bash
opencli files-fix_permissions stefan stefan.pejcic.rs
```

#### Single User
Fix permissions for all folders iniside user home directory (`/var/www/html/`):

```bash
opencli files-fix_permissions <USERNAME>
```

#### All Users

Use the `--all` flag to change permissions for all active users:

```bash
opencli files-fix_permissions --all
```


### Malware Scan

Scans a user's website files (`html_data` volume) with ClamAV. Infected files are moved to the `.quarantine/` folder in the user's files and are listed on the Quarantine page in OpenPanel. Files marked as safe in OpenPanel are skipped, and so are `vendor/` and `node_modules/` folders.

Requires the ClamAV service (`clamav` container) to be running.

Scan files for a single user:
```bash
opencli files-malware_scan <USERNAME>
```

Scan files for all active users:
```bash
opencli files-malware_scan --all
```

### Resellers Storage

Calculates total disk usage for all users owned by each reseller and saves the current and maximum disk blocks to the reseller's limits file (`/etc/openpanel/openadmin/resellers/<RESELLER>.json`).

```bash
opencli files-calculate_resellers_storage
```
