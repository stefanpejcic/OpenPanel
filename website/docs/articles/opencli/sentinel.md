# Sentinel

Sentinel is an AI-powered service that autonomously monitors your server, makes decisions, and sends alerts to the notifications center and emails.

Using the command you can check current system health and get suggestions on improving configuration:

```bash
opencli sentinel
```

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel
--------------------------------------------------------------------------------
  Sentinel - OpenPanel server health monitor
--------------------------------------------------------------------------------
Checking services:
[✔] openpanel is active and responding.
[✔] admin is active.
[✔] caddy is active and responding.
[✔] podman.socket is active.
[✔] MariaDB container active and responding.
[✔] csf is active.
[✔] phpmyadmin container is active.
[✔] No OOM errors detected.
--------------------------------------------------------------------------------
Checking traffic:
[✔] No unusual traffic detected on web ports (80|443).
--------------------------------------------------------------------------------
Checking logins, resources, and DNS...
[✔] No new logins to OpenAdmin.
[✔] No active SSH sessions.
[✔] Disk 41% < threshold 85%
[✔] Load 0.42 < threshold 20.
[✔] RAM 38% < threshold 85%
[✔] CPU 7% < threshold 90%
[✔] No SWAP configured.
[✔] All nameservers resolve to local IPs.
[✔] No dead user containers found (checked in 2s).
--------------------------------------------------------------------------------
All Tests Passed!
--------------------------------------------------------------------------------
18 PASS  0 WARN  0 FAIL
--------------------------------------------------------------------------------
```
</details>

![sentinel openpanel cli](/img/docs-content/kg56D2x2-sentinel-openpaenl.png)

Additional flags:

- `--startup` - run the actions performed after a server reboot: start stopped containers for root and all users and send the reboot notification.
- `--report` - send the *Daily Usage Report* email (if email alerts are enabled).
- `--action=<name> --title=<title> --message=<message>` - log a custom notification to OpenAdmin > Notifications (and email/webhook) if notifications are enabled for that action.

Example:
```bash
opencli sentinel --action=admin_api --title="OpenAdmin API is on" --message="API access is now on."
```

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel --action=admin_api --title="OpenAdmin API is on" --message="API access is now on."
[!] Notifications are disabled for action: admin_api
```
</details>

Nothing is printed when the notification is logged. The message above is shown only when notifications for that action are disabled in `notifications.ini`.


<details>
  <summary>Example email notifications</summary>

### Login to OpenAdmin from an unknown ip address
![sentinel_adminlogin.png](/img/docs-content/jjdk02DP-sentinel-adminlogin.png)

### Server reboot alert
![sentinel_reboot.png](/img/docs-content/1XpKN1gy-sentinel-reboot.png)

### Daily usage report
![sentinel_dailyusage.png](/img/docs-content/kgjvSc5R-sentinel-dailyusage.png)

### Disk usage alert
![sentinel_diskspace.png](/img/docs-content/zvSnGwMB-sentinel-diskspace.png)

### High LOAD alert
![sentinel_highload.png](/img/docs-content/PrTWFcp8-sentinel-highload.png)

### CPU usage alert
![sentinel_highcpu.png](/img/docs-content/s2kSP0Cx-sentinel-highcpu.png)

### MEM usage alert
![sentinel_highmem.png](/img/docs-content/PJ4wjXpv-sentinel-highmem.png)

### SWAP usage alert
![sentinel_highswap.png](/img/docs-content/JzgBmdnx-sentinel-highswap.png)

### MySQL service inactive alert
![sentinel_mysql.png](/img/docs-content/761C3HNt-sentinel-mysql.png)

### Nameservers resolution alert
![sentinel_nameservers.png](/img/docs-content/GtqHfXVk-sentinel-nameservers.png)

### OpenPanel container inactive alert
![sentinel_opanelcontainer.png](/img/docs-content/xTpXyFKQ-sentinel-opanelcontainer.png)

### Hostname resolution alert
![sentinel_resolve.png](/img/docs-content/158mPRqx-sentinel-resolve.png)

</details>
