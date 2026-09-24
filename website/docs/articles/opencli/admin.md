
# Admin

enable/disable the admin panel, reset password, add admin users, etc.

```bash
Usage: opencli admin <command> [options]

Commands:
  on                                            Enable and start the OpenAdmin service.
  off                                           Stop and disable the OpenAdmin service.
  port [new_port]                               Display or update OpenAdmin port.
  log                                           Display the last 25 lines of the OpenAdmin error log.
  logs                                          Display live logs for all OpenAdmin services.
  list                                          List all current admin users.
  new <user> <pass> [--reseller|--super]        Add a new user, optionally as a reseller or super admin.
  password <user> <pass>                        Reset the password for the specified admin user.
  update <user> --allowed_plans=[] --max_accounts=<int> --max_disk_blocks=1000000 --logo_url=<url> Assign plans, limits, and branding for reseller.
  rename <old> <new>                            Change the admin username.
  suspend <user>                                Suspend admin user.
  unsuspend <user>                              Unsuspend admin user.
  delete <user>                                 Delete admin user.
  notifications <command> <param> [value]       Control notification preferences.

  Notifications Commands:
    check                                       Check and write notifications.
    get <param>                                 Get the value of the specified notification parameter.
    update <param> <value>                      Update the specified notification parameter with the new value.

Examples:
  opencli admin on
  opencli admin off
  opencli admin port
  opencli admin port 8443
  opencli admin log
  opencli admin logs
  opencli admin list
  opencli admin new stefan SuperStrong1
  opencli admin new stefan SuperStrong1 --reseller
  opencli admin new stefan SuperStrong1 --super
  opencli admin password stefan SuperStrong2
  opencli admin rename stefan pejcic
  opencli admin suspend pejcic
  opencli admin unsuspend pejcic
  opencli admin notifications check
  opencli admin notifications get reboot
  opencli admin notifications update reboot yes
  opencli admin notifications update load 10
  opencli admin notifications update webhook_url https://discord.com/api/webhooks/XXXX/YYYYYYY
```



### Check Status

Check if admin panel is enabled or disabled and display link on which the OpenAdmin is accessibe:

```bash
opencli admin
```

Example:
```bash
# opencli admin
● OpenAdmin is running and is available on: https://server.openpanel.co:2087/
```

### Enable / Disable OpenAdmin

OpenAdmin can run in headless mode, that is to say without a front-end UI. This can further save on memory requirements and keep your server secure.

To disable access to the OpenAdmin panel:

```bash
opencli admin off
```

To enable access to the OpenAdmin panel:

```bash
opencli admin on
```
### List Admin users

To view all admin accounts:

```bash
opencli admin list
```

### Create new Admin

To create new admin accounts:

```bash
opencli admin new <username> <password>
```


### Create new Reseller

To create new reseller accounts:

```bash
opencli admin new <username> <password> --reseller
```

### Create Super Admin

To create the super admin account (the `admin` role). Only one super admin can exist:

```bash
opencli admin new <username> <password> --super
```

### Update Reseller

To set allowed plans, limits and logo for a reseller account. Only the flags that are passed are changed:

```bash
opencli admin update <username> --allowed_plans=<id1,id2> --max_accounts=<NUMBER> --max_disk_blocks=<NUMBER> --logo_url=<URL>
```

Example:
```bash
opencli admin update reseller1 --allowed_plans=1,2 --max_accounts=10
```

### Reset Admin Password

To reset the password for an admin user:

```bash
opencli admin password <username> <new_password>
```

### Rename Admin User

To rename an existing admin user:

```bash
opencli admin rename <old_username> <new_username>
```

### Suspend Admin User

To suspend an existing admin user:

```bash
opencli admin suspend <username>
```

### Unsuspend Admin User

To unsuspend an existing admin user:

```bash
opencli admin unsuspend <username>
```

### Delete Admin User

To delete an existing admin user:

```bash
opencli admin delete <username>
```

:::info
Note: User with 'admin' role can not be deleted.
:::



## Notifications

`opencli admin notifications` allows you to change the notification preferences.

Settings are stored in `/etc/openpanel/openadmin/config/notifications.ini` file. However, it is recommended not to modify this file directly. Instead, it's best to utilize the opencli commands.

#### Get

The `get` parameter allows you to view current notification settings.

```bash
opencli admin notifications get <OPTION>
```

Example:

```bash
# opencli admin notifications get reboot
yes
```

#### Update

The `update` parameter allows you to change the notification settings.


```bash
opencli admin notifications update <OPTION> <NEW-VALUE>
```

Example:
```bash
opencli admin notifications update load 10
Updated load to 10
```

#### Check

Run all checks now and write notifications (same as running `opencli sentinel`):

```bash
opencli admin notifications check
```

#### Options

Any option from `notifications.ini` can be read or updated. The main options are:

| Option        | Description                                                        | Default |
|---------------|--------------------------------------------------------------------|---------|
| `reboot`      | Notify when the server is rebooted.                                | `yes`   |
| `attack`      | Notify on unusual traffic or SYN flood (DDoS) attacks.             | `yes`   |
| `limit`       | Notify on Out of Memory (OOM) errors in the last 24 hours.         | `yes`   |
| `update`      | Notify when a new OpenPanel update is available.                   | `yes`   |
| `dns`         | Notify when the panel domain or nameservers don't resolve to this server. | `no` |
| `login`       | Notify on OpenAdmin logins from a new IP address.                  | `yes`   |
| `ssh`         | Notify on root SSH logins from a new IP address.                   | `yes`   |
| `services`    | Comma-separated list of services to monitor.                       | `panel,admin,caddy,podman,mysql,csf,phpmyadmin` |
| `load`        | Notify when server load is above this value.                       | `20`    |
| `cpu`         | Notify when CPU usage (%) is above this value.                     | `90`    |
| `ram`         | Notify when RAM usage (%) is above this value.                     | `85`    |
| `du`          | Notify when disk usage (%) is above this value.                    | `85`    |
| `swap`        | Notify when SWAP usage (%) is above this value.                    | `85`    |
| `webhook_url` | Also send notifications to this webhook (e.g. Discord).            | empty   |

The `[ACTIONS]` section of the file has a `yes`/`no` option for each OpenCLI action that sends a notification, for example `user_create`, `domains_add` or `admin_api`.


### Port

View the current OpenAdmin port:

```bash
opencli admin port
```

Change the OpenAdmin port (must be greater than 443). The port is opened in the firewall and OpenAdmin is restarted:

```bash
opencli admin port <NUMBER>
```

Add `--no-restart` to skip opening the port in the firewall and restarting OpenAdmin:
```bash
opencli admin port <NUMBER> --no-restart
```

### View OpenAdmin logs

To multitail [all OpenAdmin logs](/docs/admin/services/logs/):

```bash
opencli admin logs
```


To tail OpenAdmin error log:

```bash
opencli admin log
```




