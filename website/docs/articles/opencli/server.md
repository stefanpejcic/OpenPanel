# Server

Scripts for managing the server.

## Logrotate

Reads configuration: [`logrotate_enable`](/docs/articles/opencli/config/#logrotate_enable) [`logrotate_size_limit`](/docs/articles/opencli/config/#logrotate_size_limit) [`logrotate_retention`](/docs/articles/opencli/config/#logrotate_retention) [`logrotate_keep_days`](/docs/articles/opencli/config/#logrotate_keep_days) and configures logrotate for caddy, openpanel, syslog.

```bash
opencli server-logrotate
```

<details>
  <summary>Example output</summary>

```bash
# opencli server-logrotate
Caddy log rotation configured.
Syslog rotation configured.
OpenPanel log rotation configured.
```
</details>

## Migrate

Migrates all data from this server to another.

Usage: 
```bash
opencli server-migrate -h <DESTINATION_IP> --user root --password <DESTINATION_PASSWORD>
```

<details>
  <summary>Example output</summary>

```bash
# opencli server-migrate -h 203.0.113.20 --user root --password 'DestinationPass'
Checking available disk on destination server ...
There is enough disk space on destination server.
Creating system users on remote server ...
Syncing files (/home directory) ...
...
[OK] Files have been copied to the remote server.

Enabling rootless podman for all users ...
Setting linger for: stefan
Enabling rootless podman for: stefan (1/3) ...
[OK] Context stefan processed
...
Restoring user quotas ...
Syncing /var/log/openpanel ...
Syncing /var/log/caddy/ ...
Syncing /etc/csf/ ...
Syncing /etc/bind ...
Syncing /etc/openpanel ...
Syncing system cronjobs...
Syncing root_mysql Docker volume ...
Syncing /root/docker-compose.yml and /root/.env ...
Restarting services on 203.0.113.20 server ...
...
[OK] Sync complete
```
</details>

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
