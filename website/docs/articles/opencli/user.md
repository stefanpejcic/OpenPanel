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



#### Create user for Reseller


```bash
opencli user-add <USERNAME> <PASSWORD> <EMAIL> "<PLAN_NAME>" --reseller=<RESELLER_USERNAME>
```

### Transfer User

To transfer user account to another server:

```bash
opencli user-transfer --account <OPENPANEL_USER> --host <DESTINATION_IP> --username <DESTINATION_SSH_USERNAME> --password <DESTINATION_SSH_PASSWORD> [--port 22] [--force] [--live-transfer]
```

- `-h`, `--host` - destination server IP.
- `-u`, `--username` - SSH user on the destination server.
- `--port` - SSH port on the destination server (default `22`).
- `--force` - overwrite the account if the username already exists on the destination server.
- `--live-transfer` - suspend the account after the transfer and forward DNS to the new server.

### Backup User

To create a full account `.tar.gz` backup of a single user (files, databases, mail, and configuration):

```bash
opencli user-backup --account <USER> [--output <DIR>] [--quiet]
```

- `--output <DIR>` - custom destination directory for the archive (defaults to the location configured for backups).
- `--quiet` - only log to file, don't print progress to stdout.

### Restore User

To restore a user account from a `.tar.gz` backup created with `user-backup`:

```bash
opencli user-restore --file <ARCHIVE> [--force] [--new-username=<NAME>] [--temp-dir=<PATH>] [--quiet]
```

- `--force` - overwrite the account if a user with the same username already exists.
- `--new-username=<NAME>` - restore under a different username than the one in the backup.
- `--temp-dir=<PATH>` - directory to extract the archive into during restore (must be empty). Defaults to `/tmp/`.
- `--quiet` - only log to file, don't print progress to stdout.

### Delete User

To delete a user and all his data run the following command:

```bash
opencli user-delete <USERNAME>
```

add `-y` flag to disable prompt.

:::warning
This action is irreversible and will permanently delete all user data.
:::

### Suspend User

To suspend (temporary disable access) to user, run the follwowing command:

```bash
opencli user-suspend <USERNAME> [-y] [--debug]
```

add `-y` flag to disable prompt.

### Unsuspend User

To unsuspend (enable access) to user, run the follwowing command:

```bash
opencli user-unsuspend <USERNAME>
```

### Rename User

To change a username run:
```bash
opencli user-rename <USERNAME> <NEW_USERNAME>
```

### Change Email

To change a email run:
```bash
opencli user-email <USERNAME> <NEW_EMAIL>
```

### Change Password

To reset the password for a OpenPanel user, you can use the `user-password` command:

```bash
opencli user-password <USERNAME> <NEW_PASSWORD>
```

Provide `random` as password to generate a strong random password:
```bash
opencli user-password <USERNAME> random
```

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

To open the link in a browser:
```bash
opencli user-login <USERNAME> --open
```

### Change Plan

Command: `opencli user-change_plan` allows you to change plan for a user.

```bash
opencli user-change_plan <username> "<new_plan_name>"
```

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


### Assign / Remove IP to User

To assign free IP address to a user run the following command:

```bash
opencli user-ip <USERNAME> <IP_ADDRESS>
```

To assign IP address **that is currently used by another user** to this user, run the following command:


```bash
opencli user-ip <USERNAME> <IP_ADDRESS> -y
```


To remove dedicated IP address from a user run:

```bash
opencli user-ip <USERNAME> delete [-y]
```

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




### Block IP

Prevent specific IP addresses or CIDR ranges from accessing any user websites:

```bash
opencli user-block_ip <username> [--list='ip_here another_ip' | --delete-all]
```

List blocked IPs for a user:
```bash
opencli user-block_ip <username>
```

Block IP addresses from accessing user websites:
```bash
opencli user-block_ip <username> --list='11.22.33.44 124.64.23.0/24'
```

Remove all blocked IP addresses for a user:
```bash
opencli user-block_ip <username> --delete-all
```

### View login log
View up to last 20 successfull logins for the user.

```bash
opencli user-loginlog <USERNAME> [--table|--text|--json]
```

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

Disable Varnish:
```bash
opencli user-varnish <USERNAME> disable 
```

Check status:
```bash
opencli user-varnish <USERNAME> status
```

Check short status (returns *Current status: on/off*):
```bash
opencli user-varnish <USERNAME>
```

