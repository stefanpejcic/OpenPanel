# Users

Manage users: Add, Delete, Suspend, Unsuspend, etc.

### List Users

To list all users, use the following command:

```bash
opencli user-list
```


<details>
  <summary>Example output</summary>

```bash
# opencli user-list
+----+----------------+-------------------+----------------+----------------+-------+---------------------+
| id | username       | email             | plan_name      | server         | owner | registered_date     |
+----+----------------+-------------------+----------------+----------------+-------+---------------------+
|  1 | stefan         | stefan@pejcic.rs  | Developer Plus | stefan         | NULL  | 2025-12-25 14:49:38 |
|  2 | panel          | stefan@netops.com | Developer Plus | panel          | NULL  | 2025-12-25 14:50:18 |
|  6 | emailfilterapi | emailfilterapi    | Developer Plus | emailfilterapi | NULL  | 2026-01-28 12:25:41 |
+----+----------------+-------------------+----------------+----------------+-------+---------------------+
```
</details>


You can also format the data as JSON:

```bash
opencli user-list --json
```

<details>
  <summary>Example output</summary>

```json
{
  "data": [
    {
      "id": 1000,
      "username": "stefan",
      "context": "stefan",
      "owner": "root",
      "package": {
        "name": "Developer Plus",
        "owner": "root"
      },
      "email": "stefan@pejcic.rs",
      "locale_code": "EN_us"
    },
    {
      "id": 1001,
      "username": "panel",
      "context": "panel",
      "owner": "root",
      "package": {
        "name": "Developer Plus",
        "owner": "root"
      },
      "email": "stefan@netops.com",
      "locale_code": "EN_us"
    },
    {
      "id": 1002,
      "username": "emailfilterapi",
      "context": "emailfilterapi",
      "owner": "root",
      "package": {
        "name": "Developer Plus",
        "owner": "root"
      },
      "email": "emailfilterapi",
      "locale_code": "EN_us"
    }
  ],
  "metadata": {
    "result": "ok"
  }
}
```
</details>


To display only user count:

```bash
opencli user-list --total
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-list --total
Total number of users: 3
```
</details>


or:

```bash
opencli user-list --total --json
```

<details>
  <summary>Example output</summary>

```bash
#opencli user-list --total --json
3
```
</details>


#### List all users Quotas


```bash
opencli user-list --quota
```


<details>
  <summary>Example output</summary>

```json
{
  "timestamp": "2026-04-01T10:16:01Z",
  "users": [
    {"username":"stefan","uid":1000,"home_path":"/home/stefan/","disk_used":269744,"disk_soft":5120000,"disk_hard":5120000,"inodes_used":130,"inodes_soft":1000000,"inodes_hard":1000000},
    {"username":"pejcic","uid":1001,"home_path":"/home/pejcic/","disk_used":550216,"disk_soft":5120000,"disk_hard":5120000,"inodes_used":143,"inodes_soft":1000000,"inodes_hard":1000000},
    {"username":"demo","uid":1002,"home_path":"/home/demo/","disk_used":4138604,"disk_soft":10240000,"disk_hard":10240000,"inodes_used":74889,"inodes_soft":1000000,"inodes_hard":1000000}
  ]
}   
```
</details>

### Add User

```
opencli user-add <USERNAME> <PASSWORD|generate> <EMAIL> "<PLAN_NAME>" [--send-email] [--debug] [--reseller=<RESELLER_USERNAME>] [--private-note=<NOTE>] [--webserver=<TYPE>] [--sql=<mysql|mariadb>] [--no-sentinel]
```

To create a new user run the following command:

```bash
opencli user-add <USERNAME> <PASSWORD> <EMAIL> <PLAN_NAME>
```

Example:
```bash
opencli user-add stefan pejcic324 stefan@pejcic.rs 'Default Plan Nginx'
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-add stefan pejcic324 stefan@pejcic.rs 'Default Plan Nginx'
[✔] Successfully added user stefan with password: pejcic324
```
</details>


:::tip
Provide `generate` as password to generate a strong random password.
:::

Optional flags:

- `--send-email` - send an email with login credentials to the user's email address.
- `--private-note=<NOTE>` - save a private note for this user (visible only in OpenAdmin).
- `--reseller=<RESELLER_USERNAME>` - assign a reseller as the owner of this user.
- `--webserver=<TYPE>` - webserver to use: `nginx`, `apache`, `openresty`, `openlitespeed` or `litespeed`, optionally prefixed with `varnish+` (e.g. `varnish+nginx`). Defaults to the plan/server setting.
- `--sql=<mysql|mariadb>` - database type to use.
- `--no-sentinel` - don't send the *user_create* notification.
- `--skip-images` - don't pull container images for the new user.
- `--debug` - display verbose information.

Example:
```bash
opencli user-add stefan generate stefan@pejcic.rs 'Default Plan Nginx' --webserver=varnish+nginx --sql=mariadb --send-email
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-add stefan generate stefan@pejcic.rs 'Default Plan Nginx' --webserver=varnish+nginx --sql=mariadb --send-email
[✔] Successfully added user stefan with password: k3Vq9XbT2mWz4LpR
```
</details>



#### Create user for Reseller


```bash
opencli user-add <USERNAME> <PASSWORD> <EMAIL> "<PLAN_NAME>" --reseller=<RESELLER_USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-add client1 'StrongPass1' client1@example.com 'Default Plan Nginx' --reseller=reseller1
[✔] Successfully added user client1 with password: StrongPass1
```
</details>

### Transfer User

To transfer user account to another server:

```bash
opencli user-transfer --account <OPENPANEL_USER> --host <DESTINATION_IP> --username <DESTINATION_SSH_USERNAME> --password <DESTINATION_SSH_PASSWORD> [--port 22] [--force] [--live-transfer]
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-transfer --account stefan --host 203.0.113.20 --username root --password 'DestinationPass'
Import started, log file: /var/log/openpanel/admin/transfers/stefan_203.0.113.20_2026-09-24_10-30-00.log
[2026-09-24 10:30:00] Testing SSH connection to root@203.0.113.20...
[2026-09-24 10:30:01] SSH connection established, starting transfer process..
[2026-09-24 10:30:01] Log file: /var/log/openpanel/admin/transfers/stefan_203.0.113.20_2026-09-24_10-30-00.log
[2026-09-24 10:30:01] PID: 48213
[2026-09-24 10:30:02] Resolved context for stefan: stefan
[2026-09-24 10:30:03]   Importing new plan...
[2026-09-24 10:30:03]   Importing user into database...
[2026-09-24 10:30:03] Listing features on remote server ...
[2026-09-24 10:30:03] Copying feature set 'default.txt' to remote server
[2026-09-24 10:30:04] Creating system user (stefan) on remote server ...
[2026-09-24 10:30:05] Syncing files for user stefan (context: stefan) ...
...
[2026-09-24 10:31:10] Importing sites ...
[2026-09-24 10:31:10]   Site imported: example.com
[2026-09-24 10:31:11] Enabling rootless podman for stefan on destination ...
[2026-09-24 10:31:12] Restoring FTP accounts for stefan (context: stefan) ...
[2026-09-24 10:31:13]   [FTP] restored ftp1@example.com -> /var/www/html/example.com
[2026-09-24 10:31:13] FTP accounts restored for context stefan.
[2026-09-24 10:31:15] Syncing mail data ...
[2026-09-24 10:31:16] Syncing maildir /var/mail/example.com → /var/mail/example.com ...
[2026-09-24 10:31:17] Checking podman context ....
[2026-09-24 10:31:17] Starting containers inside podman context on remote server ...
[2026-09-24 10:32:01] Reloading services on 203.0.113.20 server ...
[2026-09-24 10:32:12] Reloading mailserver and webmail on 203.0.113.20 server ...
[2026-09-24 10:32:20] Recalculating disk and inodes usage for all users on 203.0.113.20 ...
[2026-09-24 10:32:20] Elapsed time: 0h 2m 20s
[2026-09-24 10:32:20] SUCCESS: Transfer process for user stefan completed.
```
</details>

- `--account` - OpenPanel username to transfer.
- `-h`, `--host` - destination server IP.
- `-u`, `--username` - SSH user on the destination server (default `root`).
- `--password` - SSH password for the destination server, leave it out to use SSH keys.
- `--port` - SSH port on the destination server (default `22`).
- `--force` - overwrite the account if the username already exists on the destination server.
- `--live-transfer` - suspend the account after the transfer and forward DNS to the new server.

The destination server needs OpenPanel installed. Files, databases, domains, sites, DNS zones, SSL, FTP and email accounts (Enterprise) are copied, and containers are recreated on the destination from the user's volumes. If a plan with the same name already exists on the destination it's reused, and the log shows a warning when its feature set differs from the source. Progress is written to `/var/log/openpanel/admin/transfers/`.

### Backup User

To create a full account `.tar.gz` backup of a single user (files, databases, mail, and configuration):

```bash
opencli user-backup --account <USER> [--output <DIR>] [--quiet]
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-backup --account stefan
[2026-09-24 10:30:00] Backup started  log: /var/log/openpanel/admin/backups/stefan_backup_20260924_103000.log  (PID: 48213)
[2026-09-24 10:30:00] Account: stefan  Context: stefan  UID:1001 GID:1001  Plan: Default Plan Nginx
[2026-09-24 10:30:00] Checking disk space ...
[2026-09-24 10:30:00]   Source size  (~812 MB via du)
[2026-09-24 10:30:00]   Free at dest (~48210 MB at /backup)
[2026-09-24 10:30:00]   Disk space OK.
[2026-09-24 10:30:00] Writing manifest ...
[2026-09-24 10:30:00] Exporting panel DB rows ...
...
[2026-09-24 10:30:02] Streaming /home/stefan  →  homedir/ ...
...

═══════════════════════════════════════════════════════════════
  BACKUP COMPLETE — stefan_20260924_103000.tar.gz
═══════════════════════════════════════════════════════════════
  Archive:                 /backup/stefan_20260924_103000.tar.gz
  Size:                    402M
  Compression:             pigz
  Time taken:              0h 1m 12s

  Contents:
    Plan:                  "Default Plan Nginx"
    Domains (2):           example.com example.net
    Sites:                 1
    Feature set:           default
    FTP accounts:          2
    Home dir:              homedir/  (source ~812 MB)
    Containers:            nginx php-fpm-8.3 mysql
```
</details>

- `--output <DIR>` - custom destination directory for the archive (defaults to the location configured for backups).
- `--quiet` - only log to file, don't print progress to stdout.

### Restore User

To restore a user account from a `.tar.gz` backup created with `user-backup`:

```bash
opencli user-restore --file <ARCHIVE> [--force] [--new-username=<NAME>] [--temp-dir=<PATH>] [--quiet]
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-restore --file /backup/stefan_20260924_103000.tar.gz
Import started, log file: /var/log/openpanel/admin/imports/openpanel-import_20260924_110000.log
[2026-09-24 11:00:00] Log file: /var/log/openpanel/admin/imports/openpanel-import_20260924_110000.log
[2026-09-24 11:00:00] PID: 51234
[2026-09-24 11:00:00] Archive: /backup/stefan_20260924_103000.tar.gz
[2026-09-24 11:00:00] Checking disk space for extraction ...
[2026-09-24 11:00:00] Extracting archive ...
[2026-09-24 11:00:09] Restoring 'stefan' (context: stefan)  format v2
[2026-09-24 11:00:09] Restoring system user (stefan) ...
[2026-09-24 11:00:10] Restoring /home/stefan ...
[2026-09-24 11:00:31] Restoring panel DB rows ...
[2026-09-24 11:00:31] Plan 'Default Plan Nginx' exists (ID: 1) — reusing.
[2026-09-24 11:00:31] User row ready (ID: 7).
[2026-09-24 11:00:31] Restoring domains ...
[2026-09-24 11:00:31] Domain restored: example.com
[2026-09-24 11:00:31] Domain restored: example.net
[2026-09-24 11:00:31] Restoring FTP accounts ...
[2026-09-24 11:00:32] FTP user restored: backup@example.com
...
[2026-09-24 11:00:40] Reloading services ...
[2026-09-24 11:00:41] Recalculating quotas ...

═══════════════════════════════════════════════════════════════
  RESTORE COMPLETE — stefan
═══════════════════════════════════════════════════════════════
  Archive:                 stefan_20260924_103000.tar.gz
  Time taken:              0h 0m 41s

  Restored:
    System user:           stefan  [created]
    Home directory:        /home/stefan/  (812M)
...
```
</details>

- `--force` - overwrite the account if a user with the same username already exists.
- `--new-username=<NAME>` - restore under a different username than the one in the backup.
- `--temp-dir=<PATH>` - directory to extract the archive into during restore (must be empty). Defaults to `/tmp/`.
- `--quiet` - only log to file, don't print progress to stdout.

### Delete User

To delete a user and all his data run the following command:

```bash
opencli user-delete <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-delete stefan
This will permanently delete user 'stefan' and all associated data. Confirm? [Y/n]: y
User stefan deleted successfully.
```
</details>

add `-y` flag to disable prompt.

:::warning
This action is irreversible and will permanently delete all user data.
:::

### Suspend User

To suspend (temporary disable access) to user, run the follwowing command:

```bash
opencli user-suspend <USERNAME> [-y] [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-suspend stefan
Are you sure you want to suspend OpenPanel user 'stefan'? (y/N) y
User 'stefan' suspended successfully.
```
</details>

add `-y` flag to disable prompt.

### Unsuspend User

To unsuspend (enable access) to user, run the follwowing command:

```bash
opencli user-unsuspend <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-unsuspend stefan
User 'stefan' unsuspended successfully.
```
</details>

### Rename User

To change a username run:
```bash
opencli user-rename <USERNAME> <NEW_USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-rename stefan pejcic
User 'stefan' successfully renamed to 'pejcic'.
```
</details>

### Change Email

To change a email run:
```bash
opencli user-email <USERNAME> <NEW_EMAIL>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-email stefan stefan@example.com
Success: Email for user 'stefan' updated to 'stefan@example.com'
```
</details>

### Change Password

To reset the password for a OpenPanel user, you can use the `user-password` command:

```bash
opencli user-password <USERNAME> <NEW_PASSWORD>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-password stefan 'NewStrongPass1'
Successfully changed password for user stefan
```
</details>

Provide `random` as password to generate a strong random password:
```bash
opencli user-password <USERNAME> random
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-password stefan random
Successfully changed password for user stefan, new random generated password is: 7kQ!vR2m#Lx9TzpA
```
</details>

### Login as User

This command allows you to generate an auto-login link for any OpenPanel user.

```bash
opencli user-login <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-login demouser --open
https://demo.openpanel.org:2083/login_autologin?admin_token=RMWvZK1cdeRkZQJGVQv682qby9XIPr&username=demouser
```
</details>

To invalidate an existing token for a user:
```bash
opencli user-login <USERNAME> --delete
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-login demouser --delete
Auto-login token 'b7f3c9e2a1d84f06' for user demouser is now invalidated.
```
</details>

To open the link in a browser:
```bash
opencli user-login <USERNAME> --open
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-login demouser --open
https://srv.example.com:2083/login_autologin?admin_token=b7f3c9e2a1d84f06&username=demouser
```
</details>

### Change Plan

Command: `opencli user-change_plan` allows you to change plan for a user.

```bash
opencli user-change_plan <username> "<new_plan_name>"
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-change_plan stefan "Developer Plus"
[✔] Total CPU limit (4) set successfully.
[✔] Tasks ceiling set to 300 (derived from RAM; /home/stefan/TasksMax overrides).
[✔] Total RAM limit (6G) set successfully.
[✔] Disk limit (10) and inodes (1000000) applied successfully.
Changing port speed to 100 is not possible at the moment.
Plan changed successfully for user stefan from Standard plan to Developer Plus
```
</details>

### Quota

Command: `opencli user-quota` enforces and recalculates disk and inodes for specific or all users.

```bash
opencli user-quota <username|--all>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-quota stefan
[2026-01-28 17:51:40] Processing user: stefan
[2026-01-28 17:51:40] Quota set for user stefan: 20480000 blocks (20 GB) and 2500000 inodes
[2026-01-28 17:51:40] Updating repquota file...
[2026-01-28 17:51:40] Repquota file updated successfully: /etc/openpanel/openpanel/core/users/repquota
```
</details>

### Check / Disable 2FA

To disable **Two-Factor Authentication** for a user, run the following command:

```bash
opencli user-2fa <USERNAME> [disable]
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-2fa stefan
Two-factor authentication for stefan is ENABLED.

# opencli user-2fa stefan disable
Two-factor authentication for stefan is now DISABLED.
```
</details>


### Assign / Remove IP to User

To assign free IP address to a user run the following command:

```bash
opencli user-ip <USERNAME> <IP_ADDRESS>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-ip stefan 203.0.113.11
IP successfully changed for user stefan to dedicated IP address: 203.0.113.11
```
</details>

To assign IP address **that is currently used by another user** to this user, run the following command:


```bash
opencli user-ip <USERNAME> <IP_ADDRESS> -y
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-ip stefan 203.0.113.11 -y
IP successfully changed for user stefan to dedicated IP address: 203.0.113.11
```
</details>


To remove dedicated IP address from a user run:

```bash
opencli user-ip <USERNAME> delete [-y]
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-ip stefan delete
IP configuration deleted for user stefan.
IP successfully changed for user stefan to shared IP address: 203.0.113.10
```
</details>

Add `--debug` to any `user-ip` command to display verbose information.


### Check

Check files and security for a user:

```bash
opencli user-check <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-check mozda
===== Checking user: mozda =====
---- Docker Daemon Security ----
[INFO] Running for user: mozda
[WARN] Inter-container communication on default bridge is allowed
[WARN] Docker logging level is not set to 'info' ('default')
[PASS] Docker is not allowed to modify iptables
[PASS] Not using deprecated aufs storage driver
[WARN] TLS authentication for Docker daemon is NOT configured
[PASS] Docker daemon is NOT listening on TCP socket
[PASS] Experimental features are NOT enabled
[PASS] Rootless context is configured
[PASS] Containers can not get new privileges
[PASS] Google DNS resolvers are configured

---- System and User Files ----
[PASS] /home/mozda/docker-data is configured for docker data.
[PASS] home.mozda.bin.rootlesskit file exists.
[PASS] .env file exists.
[PASS] docker-compose.yml file exists.
[PASS] backup.env file exists.
[PASS] crons.ini file exists.
[PASS] custom.cnf file exists.
[PASS] default.vcl file exists.
[PASS] httpd.conf file exists.
[PASS] nginx.conf file exists.
[PASS] openresty.conf file exists.
[PASS] pma.php file exists.
[PASS] Disk usage is below quota (3.80 GB used of 4.88 GB).
[PASS] Inode usage is below quota (113350 used of 1000000).
[INFO] Checking files ownership for user: mozda (UID: 1004, GID: 1004)
[WARN] File '/home/mozda/docker-data/overlay2/25af867389efa6af7eb6120dceef0bb601796ab469af0c9c4c798671594be432/diff/var/www/html' is owned by UID:100032 instead of UID:1004
[FAIL] Some docker files in /home/mozda/docker-data/ are not owned by UID:1004

---- Container Security Checks ----
[INFO] Found 2 running container(s)
---- Container: openresty ----
[PASS] openresty: Container running from OpenPanel
[PASS] openresty: Container name matches compose service name: openresty
[PASS] openresty: Container is running
[PASS] openresty: Container is running as root
[PASS] openresty: Using specific tag for image: openresty/openresty:bullseye-fat
[WARN] openresty: No HEALTHCHECK configured
[WARN] openresty: SELinux options not set
[WARN] openresty: --no-new-privileges restriction not set
[PASS] openresty: No extra Linux capabilities added
[PASS] openresty: Privileged mode is disabled
[WARN] openresty: Bind mounts under /home/ or /etc/openpanel/
[PASS] openresty: Docker socket is NOT mounted
[PASS] openresty: sshd is NOT running inside container
[PASS] openresty: Using OpenPanel network (mozda_www)
[PASS] openresty: All ports are bound to 127.0.0.1
[PASS] openresty: Memory usage is limited to .50 GB
[PASS] openresty: CPU usage is limited to .50 CPUs
[PASS] openresty: PID cgroup limit is set (100)
[WARN] openresty: Root filesystem is NOT read-only
[PASS] openresty: Host devices are NOT exposed
[WARN] openresty: Default ulimit is used
---- Container: php-fpm-8.2 ----
[PASS] php-fpm-8.2: Container running from OpenPanel
[PASS] php-fpm-8.2: Container name matches compose service name: php-fpm-8.2
[PASS] php-fpm-8.2: Container is running
[PASS] php-fpm-8.2: Container is running as root
[WARN] php-fpm-8.2: No HEALTHCHECK configured
[WARN] php-fpm-8.2: SELinux options not set
[WARN] php-fpm-8.2: --no-new-privileges restriction not set
[PASS] php-fpm-8.2: No extra Linux capabilities added
[PASS] php-fpm-8.2: Privileged mode is disabled
[WARN] php-fpm-8.2: Bind mounts under /home/ or /etc/openpanel/
[PASS] php-fpm-8.2: Docker socket is NOT mounted
[PASS] php-fpm-8.2: sshd is NOT running inside container
[PASS] php-fpm-8.2: Using OpenPanel network (mozda_db)
[PASS] php-fpm-8.2: No exposed ports
[PASS] php-fpm-8.2: Memory usage is limited to .25 GB
[PASS] php-fpm-8.2: CPU usage is limited to .12 CPUs
[PASS] php-fpm-8.2: PID cgroup limit is set (100)
[WARN] php-fpm-8.2: Root filesystem is NOT read-only
[PASS] php-fpm-8.2: Host devices are NOT exposed
[WARN] php-fpm-8.2: Default ulimit is used

===== Audit Summary =====
Total checks performed: 70
✅ Passed: 50
❌ Failed: 1
⚠️  Warnings: 16
ℹ️  Info: 3

⚠️  Critical issues found! Please review and address failed checks.
```
</details>




### Report

Health report for a user: what is working and what needs attention, from the same checks the OpenPanel interface shows as warnings and the email notifications use.

```bash
opencli user-report <USERNAME|--all> [--section <name,...>] [--json]
```

The report starts with a summary of all problems found, errors first, then one part per section. Lines marked `[ERROR]` need fixing, `[WARN]` might cause problems, `[INFO]` is good to know. The command exits with `1` when any error is found, or when the user doesn't exist.

Sections, in report order:

| Section | What it checks |
|---|---|
| `info` | Account, server (load, RAM, disk), plan limits vs. usage, enabled modules and features, OpenPanel version, system containers |
| `account` | Locale, admin notes, custom message, email notification preferences and whether they can be delivered, favorites |
| `containers` | Status, CPU/RAM/PIDs vs. limits, restarts, OOM kills, containers not from the default template, hourly usage history |
| `processes` | Processes per container, top processes by CPU and RAM, zombie processes |
| `files` | Disk and inodes quota, biggest volumes and folders, permission problems, trash and auto-purge, FTP accounts and connections, malware scanner and quarantine |
| `backups` | Backup destination, schedule, retention, encryption, backup service status, last run, full account archives |
| `domains` | Domains with size, status (active, suspended), docroot or redirect |
| `dns` | Nameservers, zone validation, where each domain points, SPF/DKIM/DMARC, dynamic DNS |
| `ssl` | Certificate type, issuer and expiry per domain, AutoSSL errors |
| `waf` | WAF status, level and profiles per domain, blocked requests, top rules and IPs |
| `webserver` | Web server, Varnish, config test, vhosts and the folders they serve, error log |
| `php` | Versions per domain and their support status, php.ini changes vs. the defaults, extensions, ionCube, PHP-FPM errors |
| `websites` | Installed websites with version, available updates and whether they answer |
| `databases` | MySQL/MariaDB, PostgreSQL and MongoDB: databases, sizes, users, remote access, running queries, log errors |
| `cache` | Redis, Valkey, Memcached, Elasticsearch and OpenSearch status and usage |
| `emails` | Mail server, webmail, mailboxes and quotas, aliases, hourly sending limit |
| `crons` | Cron jobs with schedule, last and next run, results, timezone |
| `stats` | Visitor statistics, log rotation, log sizes per domain |
| `security` | 2FA, passkeys, logins, sessions, blocked IPs, API and MCP tokens, shell access, suspicious activity |
| `activity` | Last entries of the activity log |

When run in a terminal, the progress of each section is shown while the report is being made. It is left out when the output goes to a file or a pipe.

<details>
  <summary>Example output</summary>

```bash
# opencli user-report stefan --section info,ssl
################################################################################
 Report for stefan, generated 2026-10-06 09:25:45 GMT
################################################################################

== SUMMARY ==
  2 error(s), 0 warning(s), 0 info

  [ERROR] INFORMATION > SERVER: Server load (10.23) is over twice the CPU cores (4).
  [ERROR] SSL: example.com has no SSL certificate yet, AutoSSL hasn't issued one.

== INFORMATION ==
  Username:                stefan
  Status:                  Active
  Owner:                   root
  Email:                   stefan@example.com
  Created:                 2026-10-05 17:02:56
  Context:                 stefan
  IP address:              192.0.2.10 (shared)

  -- SERVER --
  Hostname:                server.example.com
  OS:                      Ubuntu 24.04 LTS, kernel 6.8.0-31-generic
  CPU:                     4 cores, x86_64, Intel Xeon Processor (Cascadelake)
  Load:                    8.19, 10.23, 10.04 (1, 5, 15 min)
  [ERROR] Server load (10.23) is over twice the CPU cores (4).
  RAM:                     2.98 GB used of 8.73 GB (34%)
  Disk:                    31.21 GB used of 98.31 GB, 62.09 GB free (34%)

  -- PLAN --
  Plan:                    Standard plan
  CPU / RAM:               2 cores / 2g
...
== SSL ==
  Domain                               Type     Issuer                 Expires      Days left
  example.com                          auto     -                      -            -
  [ERROR] example.com has no SSL certificate yet, AutoSSL hasn't issued one.
  site.example.com                     auto     Let's Encrypt          2026-12-30   85
...
  Valid:                   15 of 16
```
</details>

Report for every user:

```bash
opencli user-report --all
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-report --all
################################################################################
 Report for stefan, generated 2026-10-06 09:25:45 GMT
################################################################################

== SUMMARY ==
...
################################################################################
 Report for pejcic, generated 2026-10-06 09:25:58 GMT
################################################################################
...
```
</details>

#### JSON

With `--json` the report is saved as JSON instead of printed, so it can be read by other tools. For a single user it is saved to `/home/<context>/user_report_status.json`:

```bash
opencli user-report stefan --json
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-report stefan --json
Report for stefan saved to /home/stefan/user_report_status.json
```
</details>

With `--all` every user is saved as one line (NDJSON) to `/etc/openpanel/openpanel/user_report_status.json`:

```bash
opencli user-report --all --json
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-report --all --json
Reports for 2 user(s) saved to /etc/openpanel/openpanel/user_report_status.json
```
</details>

Both files are replaced in one step when the report is done, they never contain a half written report. `--section` works with `--json` too. In the JSON, `issues` has every problem found (errors first) and `sections` has each section's rows as `group`/`label`/`value`, its own issues and its plain text output:

<details>
  <summary>Example JSON</summary>

```json
{
  "username": "stefan",
  "context": "stefan",
  "owner": "root",
  "plan": "Standard plan",
  "suspended": false,
  "generated_at": "2026-10-06T09:26:02Z",
  "duration_ms": 6520,
  "summary": { "errors": 1, "warnings": 0, "info": 0 },
  "issues": [
    { "severity": "error", "section": "SSL", "message": "example.com has no SSL certificate yet, AutoSSL hasn't issued one." }
  ],
  "sections": [
    {
      "key": "ssl",
      "title": "SSL",
      "duration_ms": 6510,
      "rows": [
        { "group": "", "label": "Valid", "value": "15 of 16" }
      ],
      "issues": [
        { "severity": "error", "section": "SSL", "message": "example.com has no SSL certificate yet, AutoSSL hasn't issued one." }
      ],
      "text": "\n== SSL ==\n  Domain  ..."
    }
  ]
}
```
</details>

### Block IP

Prevent specific IP addresses or CIDR ranges from accessing any user websites:

```bash
opencli user-block_ip <username> [--list='ip_here another_ip' | --delete-all]
```

List blocked IPs for a user:
```bash
opencli user-block_ip <username>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-block_ip stefan
11.22.33.44
124.64.23.0/24
```
</details>

Block IP addresses from accessing user websites:
```bash
opencli user-block_ip <username> --list='11.22.33.44 124.64.23.0/24'
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-block_ip stefan --list='11.22.33.44 124.64.23.0/24'
Blocking IP 11.22.33.44 for user stefan...
Blocking IP 124.64.23.0/24 for user stefan...
Blocklist applied for domain 'example.com'
Blocklist applied for domain 'example.net'
```
</details>

Remove all blocked IP addresses for a user:
```bash
opencli user-block_ip <username> --delete-all
```
The blocklist is emptied and Caddy is reloaded. The command prints no output.

### View login log
View up to last 20 successfull logins for the user.

```bash
opencli user-loginlog <USERNAME> [--table|--text|--json]
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-loginlog stefan
IP            Country  Time
203.0.113.5   RS       2026-09-24 09:12:44
198.51.100.7  DE       2026-09-23 18:02:10

# opencli user-loginlog stefan --json
[
  {
    "ip": "203.0.113.5",
    "country": "RS",
    "time": "2026-09-24 09:12:44"
  }
]
```
</details>

Output is shown as a table by default. Use `--text` for plain text or `--json` for JSON.

### Varnish
Check Varnish Caching status for user and enable/disable Varnish service.

```bash
opencli user-varnish <USERNAME> [enable|disable|status]
```


Enable Varnish:
```bash
opencli user-varnish <USERNAME> enable
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-varnish stefan enable
...
Varnish Cache is now enabled.
```
</details>

Disable Varnish:
```bash
opencli user-varnish <USERNAME> disable
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-varnish stefan disable
...
Varnish Cache is now disabled.
```
</details>

Check status:
```bash
opencli user-varnish <USERNAME> status
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-varnish stefan status
Varnish Cache is enabled.
```
</details>

Check short status (returns *Current status: on/off*):
```bash
opencli user-varnish <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli user-varnish stefan
Current status: On
```
</details>

