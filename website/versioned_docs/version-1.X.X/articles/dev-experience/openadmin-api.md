# OpenAdmin API Documentation

OpenAdmin API can be used to manage options available from the OpenAdmin interface: hosting plans, accounts, domains, services, DNS, email, security, server administration, and more. All API endpoints require **Bearer JWT authentication**.

:::info
OpenAdmin API is available only on [OpenPanel Enterprise edition](/enterprise/).
To use the API, enable it first from the [OpenAdmin > Settings > API Access](/docs/admin/settings/api/) page or from terminal with [`opencli api on` command](/docs/articles/opencli/api/#enable).
:::

To generate a token for use with the endpoints, send a **POST request** to `/api/` using your OpenAdmin username and password:

```bash
curl -X POST "https://OPENADMIN_DOMAIN_OR_IP:2087/api/" && \
  -H "Content-Type: application/json" -d '{"username":"OPENADMIN_USERNAME","password":"OPENADMIN_PASSWORD"}'
```

We recommend creating a separate admin user dedicated solely to API usage. This way, you can quickly suspend or delete that account if its credentials are ever compromised.

Example response with the token is:

```json
{"access_token":"eyJhbGciOiJIUzI1NiIsInRs4CI6IkpXVCJ9.eyJmcmVzaCI6ZmFsc2UsImlhdCI6MTc4MDc0OTkwOCwianRpIjoiZDcwZmY2NmEtNDNiOC00OGZkLThiMWEtZDIwYWUwMDNhNjEyIiwidHlwZSI6ImFjY2VzcyIsInN1YiI6Im9wc3VwcG9ydCIsIm5iZiI6MTc4MDc0OTkwOCwiY3NyZiI6IjZlOGRmNGVmLTA0MjktNGNkYy1hNTg5LTJkYzUwMDk2ZDI2NSIsI3V4cCI6MTc4M33c1MDgwOH0.WLQ8rsVOF6pQg07i1fI1lXfd8_cCmryiOjGqg8zgOmw"}
```

This token must be sent in header for all other API endpoints:

```
-H "Authorization: Bearer YOUR_JWT_TOKEN" -H "Content-Type: application/json"
```

💡 Hint: Use `opencli api list` command to view usage examples for all the endpoints available on your OpenPanel version.

:::danger
Some endpoints below are destructive or high-privilege (rebooting the server, changing the root SSH password, disabling OpenAdmin, migrating a server, WAF/firewall management). They run with the same privilege as the admin panel itself — scope and rotate API tokens accordingly.
:::

---

## Users

### List All Accounts

`GET /api/users`

<details>
  <summary>Example response</summary>

```json
[
  {
    "username": "stefan",
    "email": "stefan@pejcic.rs",
    "plan_name": "default_plan_nginx"
  }
]
```
</details>

### Create Account

`POST /api/users`
**Request Body:**

```json
{
  "email": "stefan@pejcic.rs",
  "username": "stefan",
  "password": "s32dsasaq",
  "plan_name": "default_plan_nginx"
}
```

**Response:**

```json
{
  "response": {
    "message": "Successfully added user stefan password: s32dsasaq"
  },
  "success": true
}
```

### Get Single Account

`GET /api/users/{username}`

### Suspend / Unsuspend / Change Password

`PATCH /api/users/{username}`
**Request Body Example:**

```json
{
  "action": "suspend",
  "password": "NEW_PASSWORD_HERE"
}
```

### Change Plan

`PUT /api/users/{username}`
**Request Body:**

```json
{
  "plan_name": "default_plan_nginx"
}
```

### Delete Account

`DELETE /api/users/{username}`

### Autologin to User Account

`CONNECT /api/users/{username}`

### List Users with Dedicated IPs

`GET /api/ips`

### List a User's Containers

`GET /api/users/{username}/containers`

Add `?stats=1` to get live CPU/memory/network stats for that user's containers instead of the compose service definitions.

### Manage a User's Container

`POST /api/users/{username}/containers/{action}/{container_name}`

`{action}` is one of `start`, `stop`, `restart`, `cpu`, `ram`.

**Request Body (for `cpu`/`ram` only):**

```json
{
  "value": "2"
}
```

---

## Plans

### List All Plans

`GET /api/plans`
**Response:**

```json
{
  "plans": [
    {
      "id": 1,
      "name": "ubuntu_nginx_mysql",
      "description": "Unlimited disk space and Nginx",
      "bandwidth": 100,
      "cpu": "1",
      "ram": "1g",
      "disk_limit": "10 GB",
      "storage_file": "0 GB",
      "inodes_limit": 1000000,
      "db_limit": 0,
      "domains_limit": 0,
      "websites_limit": 10,
      "docker_image": "openpanel/nginx"
    }
  ]
}
```

### Create Plan

`POST /api/plans`
**Request Body:**

```json
{
  "name": "Starter",
  "description": "Starter hosting plan",
  "email_limit": "10",
  "max_email_quota": "0",
  "max_hourly_email": "0",
  "ftp_limit": "5",
  "domains_limit": "5",
  "websites_limit": "5",
  "disk_limit": "5000",
  "inodes_limit": "100000",
  "db_limit": "5",
  "cpu": "1",
  "ram": "1",
  "bandwidth": "100",
  "feature_set": "default"
}
```

### Get Single Plan

`GET /api/plans/{plan_id}`

### Edit Plan

`PUT /api/plans/{plan_id}` (or `PATCH`)
**Request Body:** same fields as Create Plan — only send the ones you want to change, the rest fall back to defaults (`0` for limits, `1` for cpu/ram, `100` for bandwidth, `default` for feature_set).

### Delete Plan

`DELETE /api/plans/{plan_id}`

---

## Domains

### List All Domains

`GET /api/domains`

### Add Domain

`POST /api/domains/new`
**Request Body:**

```json
{
  "username": "current_user",
  "domain": "example.com",
  "docroot": "/var/www/html/example.com"
}
```

### Suspend Domain

`POST /api/domains/suspend/{domain}`

### Unsuspend Domain

`POST /api/domains/unsuspend/{domain}`

### Delete Domain

`POST /api/domains/delete/{domain}`

### View / Update Docroot

`GET /api/domains/docroot/{domain}`
`POST /api/domains/docroot/{domain}`
**Request Body:**

```json
{
  "docroot": "/var/www/html/newroot"
}
```

### View / Update DNS Zone File

`GET /api/domains/{domain}/dns`
`POST /api/domains/{domain}/dns`

Edits the raw bind zone file. On save the zone is validated with `named-checkzone` and reloaded — if validation fails, the previous version is restored automatically.

**Request Body:**

```json
{
  "content": "$TTL 86400\n@ IN SOA ..."
}
```

### View / Update Caddy Config for a Domain

`GET /api/domains/{domain}/caddy`
`POST /api/domains/{domain}/caddy`

Same validate-then-reload behavior as the DNS zone endpoint, but for the domain's Caddy config block.

**Request Body:**

```json
{
  "content": "example.com {\n  reverse_proxy ...\n}"
}
```

### View / Update VirtualHost Config

`GET /api/domains/{domain}/vhost/{username}`
`POST /api/domains/{domain}/vhost/{username}`

Updates the raw nginx/apache/openresty vhost file for that user's domain and restarts the webserver container.

**Request Body:**

```json
{
  "content": "..."
}
```

### SSL: View Status, Request AutoSSL, Upload Custom Cert, View Logs

`GET /api/domains/{domain}/ssl`
`POST /api/domains/{domain}/ssl`

**Request Body (AutoSSL):**

```json
{ "action": "autossl" }
```

**Request Body (custom certificate — paths must live under `/var/www/html/`):**

```json
{
  "action": "custom",
  "public_path": "/var/www/html/example.com/cert.pem",
  "private_path": "/var/www/html/example.com/key.pem"
}
```

**Request Body (issuance logs):**

```json
{ "action": "logs" }
```

### View Domain Access Log

`GET /api/domains/{domain}/log`

Supports `?page=N` and `?show_all=true` query params for pagination.

### View Domain GoAccess Stats

`GET /api/domains/{domain}/stats/{username}`

Returns the pre-rendered GoAccess HTML report (regenerated every 24h).

### DNS Cluster

`GET /api/dns/cluster` — view cluster config (allowed slave IPs, enabled state)
`POST /api/dns/cluster` — enable/disable the cluster, or add/remove a slave IP

**Request Body examples:**

```json
{ "action": "create", "ip": "1.2.3.4" }
```
```json
{ "action": "delete", "ip": "1.2.3.4" }
```
```json
{ "action": "enable" }
```

### DNS Cluster Node Status

`GET /api/dns/cluster/{ip}`

Checks reachability of a slave node via `rndc` first, falling back to SSH.

### DNS Zone Templates

`GET /api/dns/zone-templates`
`POST /api/dns/zone-templates`
**Request Body:**

```json
{
  "zone_template_ipv4": "...",
  "zone_template_ipv6": "..."
}
```

### Domain File Templates

`GET /api/domains/file-templates`
`POST /api/domains/file-templates`

Default templates used for new domains: `default_page`, `suspended_user`, `suspended_website`, `docker_nginx_domain`, `docker_openresty_domain`, `docker_apache_domain`, `docker_varnish`, `docker_caddy`. Send only the keys you want to update.

---

## Usage

### API Usage Info

`GET /api/usage`

### Current Usage Stats

`GET /api/usage/stats`
**Response:**

```json
{
  "usage_stats": "{\"timestamp\": \"2024-09-03\", \"users\": 1, \"domains\": 2, \"websites\": 0}"
}
```

### CPU Usage

`GET /api/usage/cpu`
**Response Example:**

```json
{
  "core_0": 0,
  "core_1": 0
}
```

### Memory Usage

`GET /api/usage/memory`

### Disk Usage Per User

`GET /api/usage/disk`

### Server Disk Usage

`GET /api/usage/server`
**Response Example:**

```json
[
  {
    "device": "/dev/vda1",
    "mountpoint": "/",
    "fstype": "ext4",
    "total": 123690532864,
    "used": 63366230016,
    "free": 60307525632,
    "percent": 51.2
  }
]
```

---

## System

### Get System Information

`GET /api/system`
**Response Example:**

```json
{
  "hostname": "stefi",
  "os": "Ubuntu 24.04 LTS",
  "time": "2024-09-04 15:09:16",
  "kernel": "6.8.0-36-generic",
  "cpu": "DO-Premium-Intel(x86_64)",
  "openpanel_version": "0.2.7",
  "running_processes": 178,
  "package_updates": 98,
  "uptime": "18905"
}
```

### Docker Info

`GET /api/docker/info`

Raw output of the Docker Engine `info` API call (containers, images, storage driver, etc).

---

## Services

### List Monitored Services

`GET /api/services`

### Edit Services List

`PUT /api/services`
**Request Body Example:**

```json
{
  "service1": { "name": "Service One", "status": "active" },
  "service2": { "name": "Service Two", "status": "inactive" }
}
```

### Check Status of Services

`GET /api/services/status`

### Start Service

`POST /api/service/start/{service_name}`

### Restart Service

`POST /api/service/restart/{service_name}`

### Stop Service

`POST /api/service/stop/{service_name}`

---

## Notifications

### List Notifications

`GET /api/notifications`

**Response Example:**

```json
[
  { "title": "Update available", "message": "New package updates available." }
]
```

### Mark Notification as Read

`POST /api/notifications/{line_number}/read`

Pass `?command=mark_all_as_read` to mark every notification as read in one call.

### Delete Notification

`DELETE /api/notifications/{line_number}`

Pass `?command=delete_all` to clear the whole log.

---

## Emails

### Mailserver Settings

`GET /api/emails/settings`
`POST /api/emails/settings`

**Request Body (set webmail domain):**

```json
{ "webmail_domain": "webmail.example.com" }
```

**Request Body (switch webmail client):**

```json
{ "webmail_software": "roundcube" }
```

**Request Body (toggle mailserver services — any of `ENABLE_POSTFWD`, `ENABLE_AMAVIS`, `ENABLE_DNSBL`, `ENABLE_RSPAMD`, `ENABLE_SPAMASSASSIN`, `ENABLE_MTA_STS`, `ENABLE_OPENDKIM`, `ENABLE_OPENDMARC`, `ENABLE_POP3`, `ENABLE_IMAP`, `ENABLE_CLAMAV`, `ENABLE_FAIL2BAN`, `SMTP_ONLY`, `ENABLE_SRS`):**

```json
{ "ENABLE_CLAMAV": "1", "ENABLE_FAIL2BAN": "1" }
```

**Request Body (change storage location — only allowed before any mailbox exists):**

```json
{ "storage_type": "custom", "email_storage_location": "/mnt/mail" }
```

### List / Manage Mailboxes

`GET /api/emails/accounts` — list all mailboxes and their quotas
`POST /api/emails/accounts` — update password, quota, or send/receive restrictions
`DELETE /api/emails/accounts` — delete one or more mailboxes

**Request Body (change password):**

```json
{ "action": "password", "email": "user@example.com", "password": "newpass" }
```

**Request Body (set quota, in MB):**

```json
{ "action": "quota-set", "email": "user@example.com", "quota": "1024" }
```

**Request Body (remove quota):**

```json
{ "action": "quota-del", "email": "user@example.com" }
```

**Request Body (restrict send/receive):**

```json
{ "action": "restrict", "email": "user@example.com", "type": "send", "restrict_action": "add" }
```

**Request Body (delete accounts):**

```json
{ "emails": ["user@example.com"] }
```

### Mail Queue

`GET /api/emails/queue` — list queued messages
`POST /api/emails/queue` — retry or delete queued messages

**Request Body:**

```json
{ "action": "retry", "scope": "all" }
```
```json
{ "action": "delete", "scope": "selected", "queue_ids": ["ABC123"] }
```

### Per-Domain Email Rate Limits

`GET /api/emails/domain-limits` — view postfwd rules (add `?hits=example.com` to view recent hits for a domain)
`POST /api/emails/domain-limits` — manage rules

**Request Body (set a limit):**

```json
{ "action": "update-domain", "domain": "example.com", "username": "user1", "limit": 100 }
```

**Request Body (reset counters):**

```json
{ "action": "reset-all" }
```

Other supported actions: `reset-domain`, `reset-user`, `delete-domain`, `delete-user`, `delete-all`. You can also send `{"raw_content": "..."}` to overwrite the postfwd rules file directly.

---

## Security

### HTTP Basic Auth

`GET /api/security/basic-auth`
`POST /api/security/basic-auth`
**Request Body:**

```json
{
  "basic_auth": "yes",
  "basic_auth_username": "admin",
  "basic_auth_password": "secret"
}
```

### Blacklisted User-Agents

`GET /api/security/blacklist-useragents`
`POST /api/security/blacklist-useragents`
**Request Body:**

```json
{
  "blacklist_useragents": "BadBot\nEvilCrawler",
  "enabled": "yes"
}
```

### Disable OpenAdmin

`POST /api/security/disable-admin`

:::danger
Irreversible from the web UI/API — the panel can only be re-enabled from the terminal (`opencli admin on`). Restricted to the Super Administrator.
:::

### ConfigServer Firewall (CSF)

`GET /api/security/firewall` — check whether CSF is installed
`POST /api/security/firewall` — forward a raw command to the CSF UI backend

**Request Body:**

```json
{ "action": "cf" }
```

### CorazaWAF Status

`GET /api/security/waf`
`POST /api/security/waf`
**Request Body:**

```json
{ "status": "yes" }
```

### CorazaWAF Rule Sets

`GET /api/security/waf/rules`
`POST /api/security/waf/rules`
**Request Body:**

```json
{ "rule_name": "REQUEST-901-INITIALIZATION", "action": "off" }
```

### Two-Factor Authentication

`GET /api/security/2fa` — returns `{"totp_enabled": true}`, or a fresh `secret` + `qr_data_uri` to scan if 2FA isn't enabled yet
`POST /api/security/2fa/enable` — confirm enrollment
`POST /api/security/2fa/disable` — turn 2FA off

**Request Body (enable):**

```json
{ "secret": "BASE32SECRET", "code": "123456" }
```

**Request Body (disable):**

```json
{ "password": "your-account-password" }
```

### Passkeys

`GET /api/security/passkeys` — list passkeys for the authenticated API user
`POST /api/security/passkeys` — rename a passkey
`DELETE /api/security/passkeys` — remove a passkey

**Request Body:**

```json
{ "id": 1, "name": "Yubikey" }
```

:::info
Passkey **registration** requires a live WebAuthn ceremony in a browser and isn't available over the API — use the OpenAdmin UI to enroll a new key, then manage it here.
:::

---

## Server

### Cron Jobs

`GET /api/server/crons`
`POST /api/server/crons`
**Request Body:**

```json
{
  "jobs": [
    { "line_number": 1, "schedule": "0 3 * * *", "logging": true }
  ]
}
```

### SSH

`GET /api/server/ssh` — status, running config, authorized keys, and current settings
`POST /api/server/ssh` — start/stop/restart the service, update settings, or manage authorized keys

**Request Body (service action):**

```json
{ "action": "restart" }
```

**Request Body (settings):**

```json
{ "port": "2222", "password_auth": "no", "pubkey_auth": "yes", "permit_root_login": "no" }
```

**Request Body (add/remove a key):**

```json
{ "new_key": "ssh-ed25519 AAAA... comment" }
```
```json
{ "key_to_remove": "ssh-ed25519 AAAA..." }
```

### SSH Full Config

`GET /api/server/ssh/config` — raw `sshd_config`
`POST /api/server/ssh/config` — overwrite it (restarts SSH)

### Timezone

`GET /api/server/timezone`
`POST /api/server/timezone`
**Request Body:**

```json
{ "timezone": "Europe/Belgrade" }
```

### Drop Memory Cache / Swap

`POST /api/server/memory/drop-cache`
`POST /api/server/memory/drop-swap`

### Process Manager

`GET /api/server/processes` — supports `?sort=cpu|memory|priority|name|owner|command|pid` (prefix with `-` to reverse)
`POST /api/server/processes/{pid}/kill`

### Clustering Node

`GET /api/server/node`
`POST /api/server/node`
**Request Body:**

```json
{ "default_node": "root@1.2.3.4", "default_ssh_key_path": "/root/.ssh/id_rsa" }
```

### Root Password

`POST /api/server/root-password`
**Request Body:**

```json
{ "password": "newpassword" }
```

:::danger
Restricted to the Super Administrator. Changes the actual root SSH password on the server.
:::

### Reboot

`POST /api/server/reboot`
**Request Body:**

```json
{ "reboot_type": "graceful" }
```

`reboot_type` is `graceful` (waits, then `reboot`) or `hard` (sysrq trigger). The request blocks until the reboot delay elapses, then the connection drops.

`GET /api/server/reboot/status` — used to poll whether the panel is back up.

### Server Migration

`GET /api/server/migrate` — check status/log of an in-progress or finished migration
`POST /api/server/migrate` — start importing data from another OpenPanel server over SSH

**Request Body:**

```json
{ "host": "1.2.3.4", "root": "root", "password": "secret" }
```

---

## Settings

### Administrators

`GET /api/settings/administrators`
`POST /api/settings/administrators`

**Request Body (create — Enterprise only):**

```json
{ "action": "create", "username": "newadmin", "password": "secret123" }
```

Other supported actions: `suspend`, `unsuspend`, `delete`, `rename_user` (with `new_username`), `reset_password` (with `new_password`), `disable_2fa`, `disable_passkeys`.

### Resellers

`GET /api/settings/resellers`
`POST /api/settings/resellers`

**Request Body (create):**

```json
{ "action": "create", "username": "reseller1", "password": "secret123" }
```

**Request Body (update limits):**

```json
{ "action": "update", "username": "reseller1", "allowed_plans": "1,2", "max_accounts": "10", "max_disk_blocks": "1000" }
```

Other supported actions: `suspend`, `unsuspend`, `delete`, `rename_user`, `reset_password`, `disable_2fa`, `disable_passkeys`. A reseller token can only manage its own account and is restricted to `reset_password`.

### General Settings

`GET /api/settings/general`
`POST /api/settings/general`
**Request Body:**

```json
{ "2083_port": "2083", "2087_port": "2087", "force_domain": "panel.example.com", "dev_mode": "off" }
```

### Defaults for New Accounts

`GET /api/settings/defaults`
`POST /api/settings/defaults`
**Request Body:**

```json
{
  "values": { "DEFAULT_PHP_VERSION": "8.2", "WEB_SERVER": "nginx" },
  "services": ["phpmyadmin"]
}
```

### Default Compose/Env Templates

`GET /api/settings/defaults/files`
`POST /api/settings/defaults/files`
`DELETE /api/settings/defaults/files` — resets both files from the upstream GitHub source

**Request Body:**

```json
{ "env": "...", "compose": "..." }
```

### Per-User Compose/Env Files

`GET /api/settings/defaults/files/{username}`
`POST /api/settings/defaults/files/{username}`

### Feature Sets

`GET /api/settings/features` — list feature sets
`GET /api/settings/features/{plan}` — view one set
`POST /api/settings/features` — create a new set
`POST /api/settings/features/{plan}` — update/enable-all/disable-all/delete a set

**Request Body (create):**

```json
{ "feature_name": "my_set" }
```

**Request Body (update):**

```json
{ "action": "update", "features": ["ftp", "cron"] }
```

Other actions: `enable_all`, `disable_all`, `delete`.

### Locales

`GET /api/settings/locales`
`POST /api/settings/locales`

**Request Body (install):**

```json
{ "locale": "en-us" }
```

**Request Body (set default):**

```json
{ "default": "en-us" }
```

### Modules / Plugins

`GET /api/settings/modules`
`POST /api/settings/modules`
**Request Body:**

```json
{ "enabled_modules": ["dns", "malware_scan", "phpmyadmin"] }
```

### Custom Code

`GET /api/settings/custom-code`
`POST /api/settings/custom-code`

Send any of: `custom_css`, `custom_js`, `in_header`, `in_footer`, `post_update`, `pre_startup`, `custom_section`, `forbidden_usernames`, `restricted_domains`, `howto_guides`, `wp_themes`, `wp_plugins`, `pagespeed_api_key`.

```json
{ "custom_css": "body { background: #111; }" }
```

### PHP

`GET /api/settings/php`
`POST /api/settings/php`
**Request Body:**

```json
{ "php82": "memory_limit=512M\nupload_max_filesize=64M" }
```

Send `"options"` to update `/etc/openpanel/php/options.txt`, or `phpXY` (`php56` … `php84`) to update a specific version's `php.ini`.

### Caddy Metrics

`GET /api/settings/caddy/metrics`

Returns raw Prometheus-format metrics from Caddy's `/metrics` endpoint.

### Update Preferences

`GET /api/settings/updates` — current preference, latest available version, and update log
`POST /api/settings/updates`
**Request Body:**

```json
{ "preference": "minor_only" }
```

`preference` is one of `minor_and_major`, `minor_only`, `major_only`, `none`.

### Run Update Now

`POST /api/settings/updates/now`

Starts `opencli update --force` in the background.

### Pin/Downgrade Version

`GET /api/settings/updates/tags` — list available `openpanel/openpanel-ui` docker tags
`POST /api/settings/updates/tags`
**Request Body:**

```json
{ "version": "1.7.60" }
```

### Notification Preferences

`GET /api/settings/notifications`
`POST /api/settings/notifications`
**Request Body:**

```json
{
  "cpu": "90",
  "webhook_url": "https://example.com/hook",
  "update": "on",
  "ssh_whitelist": "1.2.3.4/32\n5.6.7.0/24"
}
```

Supports numeric thresholds (`load`, `cpu`, `ram`, `du`, `swap`, `max_total_conn`, `max_conn_per_ip`), `webhook_url`, `email`, SMTP fields (`mail_server`, `mail_port`, `mail_use_tls`, `mail_use_ssl`, `mail_debug`, `mail_username`, `mail_password`, `mail_default_sender`), the `services` toggle, boolean event toggles (`update`, `attack`, `limit`, `login`, `ssh`, `reboot`, `dns`, plus the `ACTIONS` section like `admin_create`, `user_delete`, `domains_ssl`, etc — pass `"on"`/`"off"`), and `ssh_whitelist` (newline-separated IPs/CIDRs).

---

## License & Support

### License Key

`GET /api/license` — current key
`POST /api/license` — set/validate a key
`DELETE /api/license` — remove the license

**Request Body:**

```json
{ "key": "YOUR-LICENSE-KEY" }
```

### License Info

`GET /api/license/info`

### Verify License

`POST /api/license/verify`

### Generate Support Report

`GET /api/support/report`

Runs `opencli report --public --non-interactive` and returns the generated report.

---

## Import & Transfer

### Import from cPanel / CyberPanel Backup

`GET /api/import/{panel_type}` — list import job statuses (`{panel_type}` is `cpanel` or `cyberpanel`)
`POST /api/import/{panel_type}` — start an import in the background

**Request Body:**

```json
{ "path": "/root/backup-user.tar.gz", "plan_name": "default_plan_apache" }
```

### View Import Log

`GET /api/import/logs/account/{log_filename}`

### List Detected Backup Files

`GET /api/import/backup-files`

Scans `/`, `/home`, and `/root` for `backup-*.tar.gz` files.

### OpenPanel-to-OpenPanel Transfers

`GET /api/import/transfers` — list transfer log files
`POST /api/import/transfers` — start a transfer

**Request Body:**

```json
{
  "openpanel_username": "newuser",
  "server": "1.2.3.4",
  "username": "root",
  "password": "secret",
  "live_transfer": true
}
```

### Transfer Status for a User

`GET /api/import/transfers/{username}`

### View Transfer Log

`GET /api/import/logs/transfer/{log_filename}`
