# FTP

Enable and manage FTP server shared for all OpenPanel users.


## List

List all FTP sub-users from all OpenPanel accounts:
```bash
opencli ftp-list
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-list
FTP sub-users for 'stefan':
backup@example.com | /var/www/html/example.com | 2048M
dev@example.com | /var/www/html/example.com/dev | unlimited

FTP sub-users for 'demo':
upload@demo.net | /var/www/html/demo.net | unlimited
```
</details>


List all FTP sub-users for a single OpenPanel accounts:
```bash
opencli ftp-list <OPENPANEL_USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-list stefan
FTP sub-users for 'stefan':
backup@example.com | /var/www/html/example.com | 2048M
dev@example.com | /var/www/html/example.com/dev | unlimited
```
</details>

The last column is the account's disk quota, see [Quota](#quota).

Add `--json` to either command for JSON output. `quota_mb` is `0` for unlimited accounts, `used_kb` is the disk usage from the last quota check and is empty for accounts without a quota.

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-list stefan --json
[{"openpanel_user":"stefan","username":"backup@example.com","password":"$6$rounds=656000$...","directory":"/var/www/html/example.com","uid":"1001","gid":"1001","quota_mb":"2048","used_kb":"524288"}]
```
</details>

## Add
Create FTP sub-user for openpanel user:
```bash
opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME> [--quota=<MB>] [--debug]
```

The folder must be inside `/var/www/html/` and can't point through a symlink, the same goes for `opencli ftp-path`.

`--quota` sets a disk quota right away, in MB (`500`) or with a unit (`500M`, `2G`), see [Quota](#quota).

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-add backup@example.com 'StrongPass1' /var/www/html/example.com stefan
Success: FTP user 'backup@example.com' created successfully (UID: 1001, GID: 1001).
```
</details>

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-add backup@example.com 'StrongPass1' /var/www/html/example.com stefan --quota=2G
    FTP quota for user 'backup@example.com' set to 2G.
Success: FTP user 'backup@example.com' created successfully (UID: 1001, GID: 1001).
```
</details>

## Password
Change password for FTP sub-user of openpanel user:
```bash
opencli ftp-password <username> <new_password> <openpanel_username> [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-password backup@example.com 'NewStrongPass2' stefan
Success: FTP user 'backup@example.com' password updated successfully.
```
</details>

## Path
Change FTP path for sub-user:
```bash
opencli ftp-path <username> <path> <openpanel_username> [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-path backup@example.com /var/www/html/example.com/backups stefan
Success: FTP path for user 'backup@example.com' changed successfully.
```
</details>

## Quota
Show the disk quota and usage of an FTP sub-user:
```bash
opencli ftp-quota <username> <openpanel_username>
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-quota backup@example.com stefan
FTP user 'backup@example.com' quota: 2G, used: 512M
```
</details>

Set or remove the disk quota:
```bash
opencli ftp-quota <username> <quota> <openpanel_username> [--debug]
```

`<quota>` is in MB (`500`) or with a unit (`500M`, `2G`). `0` or `unlimited` removes the quota.

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-quota backup@example.com 2G stefan
Success: FTP quota for user 'backup@example.com' set to 2G.
# opencli ftp-quota backup@example.com unlimited stefan
Success: Quota removed for FTP user 'backup@example.com'.
```
</details>

The quota is the size of the account's folder (everything under its path, also files uploaded by other means). The FTP service checks it when it starts and then every 60 seconds. Once the folder is over the quota, uploads and new folders are refused and open sessions of that account are closed, deleting and downloading files still works. Uploads are allowed again on the next check after the folder is back under the quota. `opencli ftp-quota` runs a check right away.

Quotas are stored as the 6th field in `/etc/openpanel/ftp/users/<openpanel_username>/users.list`, and the usage from the last check in `quota.usage` next to it.

## Delete
Delete FTP sub-user for openpanel user:
```bash
opencli ftp-delete <username> <openpanel_username> [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-delete backup@example.com stefan
Success: FTP user 'backup@example.com' deleted successfully.
```
</details>

## Connections
View all current connections to the FTP servise:
```bash
opencli ftp-connections
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-connections
  412 ftp       0:00 vsftpd: 203.0.113.5/backup@example.com: IDLE
  418 ftp       0:00 vsftpd: 198.51.100.7/upload@demo.net: RETR
```
</details>

View current connections for a specific OpenPanel user (FTP sub-users on the domains that user owns):
```bash
opencli ftp-connections <openpanel_username>
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-connections stefan
  412 ftp       0:00 vsftpd: 203.0.113.5/backup@example.com: IDLE
```
</details>

## Logs
View FTP service logs:
```bash
opencli ftp-logs
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-logs
Wed Sep 24 10:30:12 2026 [pid 412] CONNECT: Client "203.0.113.5"
Wed Sep 24 10:30:13 2026 [pid 411] [backup@example.com] OK LOGIN: Client "203.0.113.5"
Wed Sep 24 10:31:02 2026 [pid 413] [backup@example.com] OK UPLOAD: Client "203.0.113.5", "/var/www/html/example.com/index.php", 1824 bytes, 312.45Kbyte/sec
```
</details>

## Users
Recreate the list of all FTP users in `/etc/openpanel/ftp/all.users`, used by OpenAdmin *(internal helper, not listed in `opencli --help`)*:
```bash
opencli ftp-users
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-users
backup@example.com|$6$...|/home/stefan/docker-data/volumes/stefan_html_data/_data/example.com|1001|1001|2048 upload@demo.net|$6$...|/home/demo/docker-data/volumes/demo_html_data/_data/demo.net|1002|1002
```
</details>
