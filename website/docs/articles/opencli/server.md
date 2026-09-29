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
[OK] Files have been copied to the remote server.

Enabling rootless podman for all users ...
Setting linger for: stefan
Enabling rootless podman for: stefan (1/3) ...
[OK] Context stefan processed
...
Syncing /var/log/openpanel ...
Syncing /var/log/caddy/ ...
Syncing /usr/local/mail/openmail ...
Syncing mailboxes from /var/mail/ ...
Syncing /etc/csf/ ...
Syncing /etc/bind ...
Updated DNS zone for domain example.com - /etc/bind/zones/example.com.zone
Syncing /etc/openpanel ...
Replacing 198.51.100.10 with 203.0.113.20 in stack and Caddy configuration ...
Syncing system cronjobs...
Exporting panel databases ...
Importing panel databases on 203.0.113.20 ...
[OK] Panel databases imported.
Syncing /root/docker-compose.yml and /root/.env ...
Replacing 198.51.100.10 with 203.0.113.20 in stack and Caddy configuration ...
Restoring user quotas ...
...
Checking podman context for user: stefan
Starting containers for context: stefan (1/3)...
...
Restarting services on 203.0.113.20 server ...
Starting mailserver and webmail on 203.0.113.20 server ...
Recalculating disk and inodes usage for all users on 203.0.113.20 ...
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

The destination server needs a fresh OpenPanel install with no users. The panel database is copied with a dump instead of its data files, containers are recreated on the destination from each user's volumes, and the old server IP is replaced with the new one in DNS zones, `/root/docker-compose.yml`, `/root/.env`, the Caddy configuration and `openpanel.config`. The same containers that are running on this server are started on the destination.
