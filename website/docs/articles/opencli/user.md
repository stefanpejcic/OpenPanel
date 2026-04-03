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
opencli user-add <USERNAME> <PASSWORD|generate> <EMAIL> "<PLAN_NAME>" [--send-email] [--debug] [--reseller=<RESELLER_USERNAME>] [--server=<IP_ADDRESS>]  [--key=<SSH_KEY_PATH>]
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
Provide `random` as password to generate a strong random password.
:::


To send email to the email address for the user with login credentials, pass the `--send-email` flag.


#### Create user on Slave server

To create a new user on another server:

1. Create ssh key pair and establish ssh connection from master to the slave server.
2. Run the following command:

```bash
opencli user-add <USERNAME> <PASSWORD> <EMAIL> "<PLAN_NAME>" --server=<IP_ADDRESS> --key=<SSH_KEY_PATH>
```

#### Create all new users on Slave server

To automatically create all future users on another server:

1. Create ssh key pair and establish ssh connection from master to the slave server.
2. Set the new server IP and path to the key file in `/etc/openpanel/openadmin/config/admin.ini` file:
   ```
   [CLUSTERING]
   default_node="11.22.33.44"
   default_ssh_key_path="/root/some-key.rsa"
   ```
3. That's it, all new user accounts are created on the remote server.

> NOTE: The remote server must have a fresh installation of Ubuntu 24.04 - other distributions and versions are not supported.

#### Create user for Reseller


```bash
opencli user-add <USERNAME> <PASSWORD> <EMAIL> "<PLAN_NAME>" --reseller=<RESELLER_USERNAME>
```

### Transfer User

To transfer user account to another server:

```bash
opencli user-transfer --account <OPENPANEL_USER> --host <DESTINATION_IP> --username <OPENPANEL_USERNAME> --password <DESTINATION_SSH_PASSWORD> --port 22
```

add `--live-transfer` flag to suspend account after the transfer, and forward DNS to the new server.

### Delete User

To delete a user and all his data run the following command:

```bash
opencli user-delete <USERNAME>
```

add `-y` flag to disable prompt.

:::warning
This action is irreversible and will permanently delete all user data.
:::

To delete all users and all their data use the `--all` flag:

```bash
opencli user-delete --all
```

### Suspend User

To suspend (temporary disable access) to user, run the follwowing command:

```bash
opencli user-suspend <USERNAME> [-y]
```

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

### Resources

Command: `opencli user-resources` lists a user's active services, allows editing of their CPU and RAM limits, and can start or stop the services.

```bash
Usage: opencli user-resources <context> [options]

Options:
  --json                         Output result in JSON format.
  --update_cpu=<value>           Update CPU allocation (global or per service).
  --update_ram=<value>           Update RAM allocation (global or per service).
  --service=<service>            Specify the service name to update.
  --activate=<service>           Start the specified service.
  --deactivate=<service>         Stop the specified service.
  --force                        Force image pull before activation.
  --dry-run                      Simulate actions without applying changes.
  --debug                        Display raw output of docker-compose commands.

Example:
  opencli user-resources stefan --json
  opencli user-resources stefan --service=apache --update_cpu=1.5
```


<details>
  <summary>Example output</summary>

```json
# opencli user-resources stefan --json
{
  "context": "stefan",
  "services": [
    {
      "name": "redis",
      "cpu": "0.1",
      "ram": "0.1"
    },
    {
      "name": "mariadb",
      "cpu": "2.0",
      "ram": "2"
    },
    {
      "name": "apache",
      "cpu": "0.5",
      "ram": "0.5"
    },
    {
      "name": "php-fpm-8.5",
      "cpu": "2",
      "ram": "2"
    }
  ],
  "limits": {
    "cpu": {
      "used": 4.6,
      "total": 4
    },
    "ram": {
      "used": 4.6,
      "total": 6
    }
  },
  "message": ""
}
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
opencli user-ip <USERNAME> <IP_ADDRESS> --y
```


To remove dedicated IP address from a user run:

```bash
opencli user-ip <USERNAME> delete
```


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




### View disk usage for user

To list real-time disk and inodes usage for a user:

```bash
opencli user-disk <USERNAME> <summary|detail|path> [--json]
```

Example usage:

- Disk usage summary for user:

  <details>
    <summary>Example output</summary>
  
    ```bash
    # opencli user-disk proba summary
    
    -------------- disk usage --------------
    - 564M	/home/proba
    - 864M	/var/lib/docker/devicemapper/mnt/ac28d2b066f5ffcacf4510b042623f6a3c196bd4f5fb9e842063c5325e4d0184
    ```
    
    ```bash
    # opencli user-disk proba summary --json
    
    {"home_directory_usage": "564564", "docker_container_usage": "883864", "home_path": "/home/proba", "docker_path": "/var/lib/docker/devicemapper/mnt/ac28d2b066f5ffcacf4510b042623f6a3c196bd4f5fb9e842063c5325e4d0184"}
    ```
  </details>



- Detailed disk usage report for user:

  <details>
    <summary>Example output</summary>

    ```bash
    # opencli user-disk proba detail
    ------------- home directory -------------
    - home directory:        /home/proba
    - mountpoint:            /home/proba
    - bytes used:            61440
    - bytes total:           10375548928
    - bytes limit:           true
    - inodes used:           20
    - inodes total:          1000960
    ---------------- container ---------------
    - container directory:   /var/lib/docker/devicemapper/mnt/ac28d2b066f5ffcacf4510b042623f6a3c196bd4f5fb9e842063c5325e4d0184
    - bytes used:            1025388544
    - bytes total:           10726932480
    - inodes used:           20905
    - inodes total:          5242880
    - storage driver:        devicemapper
    ```
  </details>

- Paths for user:
  ```bash
  # opencli user-disk proba path
  
  -------------- paths --------------
  - home_directory=/home/proba
  - docker_container_path=/var/lib/docker/devicemapper/mnt/ac28d2b066f5ffcacf4510b042623f6a3c196bd4f5fb9e842063c5325e4d0184
  ```
  
  ```bash
  # opencli user-disk proba path --json
  
  {"home_directory": "/home/proba","docker_container_path": "/var/lib/docker/devicemapper/mnt/ac28d2b066f5ffcacf4510b042623f6a3c196bd4f5fb9e842063c5325e4d0184"}
  ```



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
opencli user-loginlog <USERNAME> [--json]
```

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

