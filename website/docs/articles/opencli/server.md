# Server

Scripts for managing the server.

## Logrotate

Reads configuration: [`logrotate_enable`](/cli/config.html#logrotate-enable) [`logrotate_size_limit`](/cli/config.html#logrotate-size-limit) [`logrotate_retention`](/cli/config.html#logrotate-retention) [`logrotate_keep_days`](/cli/config.html#logrotate-keep-days) and configures logrotate for caddy, openpanel, syslog.

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
opencli server-migrate -h <remote_host> -u <remote_user> [--password <password>] [--exclude-home] [--exclude-logs] [--exclude-mail] [--exclude-bind] [--exclude-openpanel] [--exclude-mysql] [--exclude-stack] [--exclude-postupdate] [--exclude-users]
```

## IPs

Updates IPs for users - used on OpenPanel UI to dispaly dedicated IP addresses to user.

For all users:
```bash
opencli server-logrotate
```

For a simgle user:
```bash
opencli server-ips <USERNAME>
```
