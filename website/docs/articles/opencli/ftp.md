# FTP

Enable and manage FTP server shared for all OpenPanel users.


## List

List all FTP sub-users from all OpenPanel accounts:
```bash
opencli ftp-list
```


List all FTP sub-users for a single OpenPanel accounts:
```bash
opencli ftp-list <OPENPANEL_USERNAME>
```

Add `--json` to either command for JSON output.

## Add
Create FTP sub-user for openpanel user:
```bash
opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME> [--debug]
```

## Password
Change password for FTP sub-user of openpanel user:
```bash
opencli ftp-password <username> <new_password> <openpanel_username> [--debug]
```

## Path
Change FTP path for sub-user:
```bash
opencli ftp-path <username> <path> <openpanel_username> [--debug]
```

## Delete
Delete FTP sub-user for openpanel user:
```bash
opencli ftp-delete <username> <openpanel_username> [--debug]
```

## Connections
View all current connections to the FTP servise:
```bash
opencli ftp-connections
```

View current connections for a specific OpenPanel user:
```bash
opencli ftp-connections <openpanel_username>
```

## Logs
View FTP service logs:
```bash
opencli ftp-logs
```

## Users
Recreate list of all ftp users *(internal helper, not available as an `opencli` command)*:
```bash
bash /usr/local/opencli/ftp/users.sh
```
