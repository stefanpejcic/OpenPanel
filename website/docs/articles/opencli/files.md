# Files

### Purge Trash

The `purge_trash` script is run periodically to remove files form users *.Trash* folder, according to the `autopurge_trash` setting.


#### Single User

To purge files for a single user:

```bash
opencli files-purge_trash --user [USERNAME]
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-purge_trash --user stefan

🧹 Trash Cleanup Report:
  - stefan: 128MB freed
----------------------------------------
✅ Total freed across all users: 128MB
```
</details>

To list files that would be purged (dry-run) for a single user:

```bash
opencli files-purge_trash --user [USERNAME] --dry-run
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-purge_trash --user stefan --dry-run
[INFO] DRY-RUN mode: No files will be deleted.
[DRY-RUN] Would delete: /home/stefan/docker-data/volumes/stefan_html_data/_data/.Trash/files/old-backup.zip (user: stefan)

🧹 Trash Cleanup Report:
  - stefan: 128MB (would be) freed
----------------------------------------
🔍 Total that *would* be freed: 128MB
```
</details>

To purge files for a single user regardless of the `autopurge_trash` setting:

```bash
opencli files-purge_trash --user [USERNAME] --force
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-purge_trash --user stefan --force
[INFO] Running in FORCE mode: All trash will be purged.

🧹 Trash Cleanup Report:
  - stefan: 128MB freed
----------------------------------------
✅ Total freed across all users: 128MB
```
</details>

#### All Users

To purge files for all users:

```bash
opencli files-purge_trash
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-purge_trash

🧹 Trash Cleanup Report:
  - demo: 12MB freed
  - stefan: 128MB freed
----------------------------------------
✅ Total freed across all users: 140MB
```
</details>

To list files that would be purged (dry-run) for all users:

```bash
opencli files-purge_trash --dry-run
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-purge_trash --dry-run
[INFO] DRY-RUN mode: No files will be deleted.
[DRY-RUN] Would delete: /home/stefan/docker-data/volumes/stefan_html_data/_data/.Trash/files/old-backup.zip (user: stefan)
...

🧹 Trash Cleanup Report:
  - demo: 12MB (would be) freed
  - stefan: 128MB (would be) freed
----------------------------------------
🔍 Total that *would* be freed: 140MB
```
</details>

To purge files for all users regardless of the `autopurge_trash` setting:

```bash
opencli files-purge_trash --force
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-purge_trash --force
[INFO] Running in FORCE mode: All trash will be purged.

🧹 Trash Cleanup Report:
  - demo: 12MB freed
  - stefan: 128MB freed
----------------------------------------
✅ Total freed across all users: 140MB
```
</details>


### Fix Permissions

The `fix_permissions` script can be used to fix permissions on user files inside their container.

For website files in `/var/www/html/` (or the given path) it:
- sets the owner and group of all files and directories to the user.
- sets the permissions of all files to 664.
- sets the permissions of all directories to 775.
- fixes the owner of the user's FTP sub-user folders.


#### Single folder

Fix permissions for specific folder for user:

```bash
opencli files-fix_permissions <USERNAME> <PATH>
```

**TIP:** add `--debug` flag to show verbose information on what is exaclty being done:
```bash
opencli files-fix_permissions <USERNAME> <PATH> --debug
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-fix_permissions stefan stefan.pejcic.rs --debug
changed ownership of '/home/stefan/docker-data/volumes/stefan_html_data/_data/stefan.pejcic.rs/index.php' from root:root to 1001:1001
mode of '/home/stefan/docker-data/volumes/stefan_html_data/_data/stefan.pejcic.rs/index.php' changed from 0644 (rw-r--r--) to 0664 (rw-rw-r--)
...
Permissions applied successfully to /var/www/html/stefan.pejcic.rs
```
</details>

<details>
  <summary>Example output</summary>

```bash
# opencli files-fix_permissions stefan stefan.pejcic.rs
Permissions applied successfully to /var/www/html/stefan.pejcic.rs
```
</details>

#### Single User
Fix permissions for all folders iniside user home directory (`/var/www/html/`):

```bash
opencli files-fix_permissions <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-fix_permissions stefan
[*] Fixing ownership for FTP path: /home/stefan/docker-data/volumes/stefan_html_data/_data/example.com
Permissions applied successfully to /var/www/html/
```
</details>

#### All Users

Use the `--all` flag to change permissions for all active users:

```bash
opencli files-fix_permissions --all
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-fix_permissions --all
Permissions applied successfully to /var/www/html/
Permissions applied successfully to /var/www/html/
...
```
</details>


### Malware Scan

Scans a user's website files (`html_data` volume) with ClamAV. Infected files are moved to the `.quarantine/` folder in the user's files and are listed on the Quarantine page in OpenPanel. Files marked as safe in OpenPanel are skipped, and so are `vendor/` and `node_modules/` folders.

Requires the ClamAV service (`clamav` container) to be running.

Scan files for a single user:
```bash
opencli files-malware_scan <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-malware_scan stefan
Scanning files for user: stefan (4812 files, vendor/ and node_modules/ excluded)
  QUARANTINED: /var/www/html/example.com/wp-content/uploads/shell.php (Php.Webshell.Generic)
  1 infected file(s) quarantined for stefan.
```
</details>

Scan files for all active users:
```bash
opencli files-malware_scan --all
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-malware_scan --all
Processing user: stefan (1/2)
Scanning files for user: stefan (4812 files, vendor/ and node_modules/ excluded)
  No malware found for stefan.
------------------------------
Processing user: demo (2/2)
Scanning files for user: demo (312 files, vendor/ and node_modules/ excluded)
  No malware found for demo.
------------------------------
DONE.
```
</details>

### Resellers Storage

Calculates total disk usage for all users owned by each reseller and saves the current and maximum disk blocks to the reseller's limits file (`/etc/openpanel/openadmin/resellers/<RESELLER>.json`).

```bash
opencli files-calculate_resellers_storage
```

<details>
  <summary>Example output</summary>

```bash
# opencli files-calculate_resellers_storage
Total resellers: 2

reseller1 has (3) accounts - total disk blocks usage:
5242880 / 31457280 (blocks)

------------------------------------------------
reseller2 has no user accounts - skipping..
```
</details>

