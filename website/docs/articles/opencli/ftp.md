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
backup@example.com | /var/www/html/example.com
dev@example.com | /var/www/html/example.com/dev

FTP sub-users for 'demo':
upload@demo.net | /var/www/html/demo.net
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
backup@example.com | /var/www/html/example.com
dev@example.com | /var/www/html/example.com/dev
```
</details>

Add `--json` to either command for JSON output.

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-list stefan --json
[{"openpanel_user":"stefan","username":"backup@example.com","password":"$6$rounds=656000$...","directory":"/var/www/html/example.com","uid":"1001","gid":"1001"}]
```
</details>

## Add
Create FTP sub-user for openpanel user:
```bash
opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME> [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli ftp-add backup@example.com 'StrongPass1' /var/www/html/example.com stefan
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
Recreate list of all ftp users *(internal helper, not available as an `opencli` command)*:
```bash
bash /usr/local/opencli/ftp/users.sh
```
