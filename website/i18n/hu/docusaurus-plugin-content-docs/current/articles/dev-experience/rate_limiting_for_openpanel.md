# Limit OpenPanel

Both OpenPanel and OpenAdmin interfaces have a built-in rate limiting and IP address blocking to protect against brute-force attacks.

This feature is enabled by default on all installations, and can be additionally customized via *OpenAdmin > Settings > OpenPanel* or from the terminal.

You can configure the maximum number of failed login attempts allowed per IP (default is `5` per minute) and the total number of failed attempts (default is `20`), after which the offending IP will be temporarily blocked by the firewall for one hour.

For OpenPanel limits are configurable in: /etc/openpanel/openpanel/conf/openpanel.config file:
```bash
[USERS]
login_ratelimit=5
login_blocklimit=20
```

![user ratelimit](/img/panel/v1/user_block.png)

For OpenAdmin limits are configurable in: /etc/openpanel/openadmin/config/admin.ini file:
```bash
[PANEL]
login_ratelimit=5
login_blocklimit=20
```

![admin ratelimit](/img/admin/admin_block.png)

If a user successfully logs in, the counter for `login_blocklimit` will reset.

Failed login attempts and blocked IP addresses are logged in the `/var/log/openpanel/admin/failed_login.log` file for OpenAdmin and in the `/var/log/openpanel/user/failed_login.log` file for OpenPanel.
