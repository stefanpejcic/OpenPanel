# Server

Scripts for managing the server.

## Logrotate

Reads configuration: [`logrotate_enable`](/docs/articles/opencli/config/#logrotate_enable) [`logrotate_size_limit`](/docs/articles/opencli/config/#logrotate_size_limit) [`logrotate_retention`](/docs/articles/opencli/config/#logrotate_retention) [`logrotate_keep_days`](/docs/articles/opencli/config/#logrotate_keep_days) and configures logrotate for caddy, openpanel, syslog.

```bash
opencli server-logrotate
```

## Migrate

Migrates all data from this server to another.

Usage: 
```bash
opencli server-migrate -h <DESTINATION_IP> --user root --password <DESTINATION_PASSWORD>
```

All available flags:
```bash
opencli server-migrate -h <remote_host> -u <remote_user> [--password <password>] [--force] [--exclude-home] [--exclude-logs] [--exclude-csf] [--exclude-mail] [--exclude-bind] [--exclude-openpanel] [--exclude-mysql] [--exclude-stack] [--exclude-postupdate] [--exclude-users] [--exclude-contexts]
```

- `-h`, `--host` - destination server IP.
- `-u`, `--user` - SSH user on the destination server.
- `--password` - SSH password for the destination server.
- `--force` - skip checking if the destination server already has users.
- `--exclude-csf` - don't sync `/etc/csf/` firewall configuration.
- `--exclude-contexts` - don't set up rootless Podman for users on the destination server.
- `--exclude-*` - skip syncing that part of the data (home directories, logs, mail, DNS zones, OpenPanel configuration, MySQL, root docker stack, post-update scripts, users).
