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

### Resolved issues

When a check passes again, Sentinel marks the matching unread notifications for that issue as read and resolved in OpenAdmin > Notifications, where they show a *Resolved after …* badge, since they no longer need admin attention. For example, if the OpenPanel container was reported as not running and is running on the next check, the *OpenPanel container not running!* notifications are marked as read. This applies to service and container checks (including recovery after a restart), disk, load, RAM, CPU, SWAP, web traffic, domain and nameserver checks. Event notifications like new logins, SSH logins, OOM kills and reboots are left unread.

It also sends a *Resolved: &lt;title&gt;* message by email/webhook saying when the issue was first reported, so admins who got the alert also learn that it's over.

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel
...
Checking services:
[✔] openpanel is active and responding.
[✔] Issue resolved, marked notification as read: OpenPanel container not running!
...
```
</details>

### What gets emailed

- **Alerts** need admin attention. They are logged as unread with a *critical* or *warning* severity and sent by email/webhook. Service alerts say what failed, what Sentinel tried, which command to check it with, and include the last log lines. If the same alert is detected again while it's still unread, its count and last seen time are updated instead of adding a new notification or sending another email. Disk, load, RAM and SSH login checks are skipped entirely while their alert is unread, so no new crashlog is generated either. This only lasts 24 hours from when the alert was first reported: after that, an unread alert no longer blocks new ones, so a new notification (and crashlog for load) is logged and emailed again.
- **Info** entries are for things Sentinel already fixed on its own, like restarting a stopped container. They are logged as already read for history and are not emailed. If the same thing happens again within 24 hours of the first entry, that entry's count, last seen time and message are updated instead of adding a new row, so a container that keeps crashing shows up once.
- All alerts from one run are sent as **one email and one webhook**. With a single alert the subject is its title, with more it's *N notifications from Sentinel on &lt;hostname&gt;* and the body lists them all. In the email each notification is shown with its severity.
- The webhook is sent as JSON with a `text` field for Slack and a `content` field for Discord, which is cut to 2000 characters since Discord rejects longer messages.

Load, CPU and RAM only alert after they stay over the threshold for **2 checks in a row** (about 10 minutes with the default cron), so short spikes don't trigger alerts. CPU usage is measured over 1 second.

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel
...
[!] CPU 96% > threshold 90% (1/2 checks), alerting if it stays high.
...
```
</details>

OOM kills are read from the kernel log with `journalctl -k`, at most once an hour, covering everything since the previous check (the last hour on the first run after a reboot). Both kills inside a container's memory limit and system-wide out of memory kills are counted, and the alert lists them per system service and per user.

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel
...
[✘] 2 user process(es) killed by OOM since 2026-10-02 14:05
...
```
</details>

### User containers

Once an hour Sentinel starts any stopped containers for root and every non-suspended user. Some user containers only run when they're needed, and Sentinel starts or stops them to match:

- `cron` and `docker-proxy` run while `crons.ini` has at least one enabled job (fully commented-out, disabled jobs don't count).
- `backup` and `docker-proxy` run while `backup.env` has a remote backup destination set. `docker-proxy` is left running while a backup from `opencli docker-backup` is in progress.
- `mysql` or `mariadb` (the one set in `MYSQL_TYPE` in the user's `.env`) runs while the user has at least one database. Databases are counted from folders in the user's MySQL data volume, so Sentinel never connects to the user's MySQL to check. If there are no databases and the account is older than 2 hours, it's stopped. It's started again when the user opens the Databases page. The type the account doesn't use is never started, since both share the same data volume.

Containers Sentinel stopped this way aren't counted or emailed to the user as restarted.

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel
...
stefan: stopping cron (not needed)
stefan: stopping docker-proxy (not needed)
stefan: stopping mariadb (not needed)
john: starting required container mysql
...
[!] Started/restarted 0 container(s) for root and 1 container(s) across 1 user(s) in 3s. Per user: john: 1.
...
```
</details>

### Timeouts

So a hung `podman` command can't stop monitoring, every podman call Sentinel makes has a timeout, each check in the parallel part (logins, disk, load, RAM, CPU, SWAP, DNS, user containers) is stopped after 10 minutes, and the whole run is stopped after 15 minutes. A stopped check sends a *Sentinel checks did not finish* alert listing which checks hung, and a stopped run sends a critical *Sentinel checks timed out!* alert. Both are resolved on the next run that finishes in time. If a previous run is still going, the new one exits with *Error: Another instance is already running.*

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel
...
[✘] check_user_containers did not finish in 600s and was stopped.
...
```
</details>

### Notifications log format

Notifications are stored in `/var/log/openpanel/admin/notifications.log`, one JSON object per line:

```json
{"id":"1790279898a1b2c3","time":"2026-09-24 19:58:17","last_seen":"2026-09-24 20:08:17","count":3,"status":"unread","severity":"critical","category":"service","source":"sentinel","title":"OpenPanel container not running!","message":"Container openpanel was not running. Sentinel tried to start it, but it is still not running. ...","resolved_at":"2026-09-24 20:13:17"}
```

- `status`: `unread` or `read`
- `severity`: `critical`, `warning` or `info`
- `category`: `service`, `resources`, `security`, `dns`, `traffic`, `system`, `update` or `action`
- `source`: `sentinel`, `update`, or the action name for `--action` notifications
- `count` and `last_seen`: how many times the alert was detected while unread (or, for info entries, happened again) within 24 hours of `time`, and when it was last detected
- `resolved_at`: when Sentinel detected that the issue was gone
- `details`: optional structured data used by the Notifications page, e.g. RAM/CPU/disk usage and top processes, OOM kills, or a link to the crashlog or update log

Lines in the old text format (from before 2.0.12) are still shown on the Notifications page as plain entries.

While an OpenPanel update is running, Sentinel skips all checks, so it doesn't alert about or recreate containers the update is restarting. An update lock older than 30 minutes is ignored.

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel
--------------------------------------------------------------------------------
  Sentinel - OpenPanel server health monitor
--------------------------------------------------------------------------------
[!] OpenPanel update in progress, skipping checks until it finishes.
--------------------------------------------------------------------------------
```
</details>

![sentinel openpanel cli](/img/docs-content/kg56D2x2-sentinel-openpaenl.png)

Additional flags:

- `--startup` - run the actions performed after a server reboot: start stopped containers for root and all users and send the reboot notification.

  User containers are started in stages so 50+ accounts don't all hit the disk at once after a reboot: up to half the CPU cores worth of users at a time (min 2, max 8), 2 seconds apart, and the next user waits (up to 60s) while IO pressure is above 40% or load is above 2x the cores. Regular 5-minute runs skip the user container check until the staged start is done.

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel --startup
root: starting openpanel_mysql (was: exited)
root: starting openpanel (was: exited)
root: starting caddy (was: exited)
...
50 users to start, staged (max 4 at a time, 2s apart, waiting on high IO/load)
[1/50] stefan: starting containers
stefan: starting nginx (was: exited)
stefan: starting mysql (was: exited)
[2/50] john: starting containers
john: starting apache (was: exited)
...
```
</details>
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

When SWAP usage is over the threshold and there is enough free RAM, Sentinel clears it with `swapoff -a && swapon -a` and logs a *SWAP cleared* info notification. It clears SWAP at most once every 24 hours. If SWAP fills up again within that time, Sentinel doesn't clear it again. Instead it sends one *High SWAP usage!* alert, because the server probably needs more RAM.

<details>
  <summary>Example output</summary>

```bash
# opencli sentinel
...
[!] SWAP 99% > threshold 40%, already cleared 5m ago. Skipping.
...
```
</details>

### MySQL service inactive alert
![sentinel_mysql.png](/img/docs-content/761C3HNt-sentinel-mysql.png)

### Nameservers resolution alert
![sentinel_nameservers.png](/img/docs-content/GtqHfXVk-sentinel-nameservers.png)

### OpenPanel container inactive alert
![sentinel_opanelcontainer.png](/img/docs-content/xTpXyFKQ-sentinel-opanelcontainer.png)

### Hostname resolution alert
![sentinel_resolve.png](/img/docs-content/158mPRqx-sentinel-resolve.png)

</details>
