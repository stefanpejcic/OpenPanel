---
sidebar_position: 7
---

# Notifications

Configure OpenAdmin notifications and Email alerts settings.

<Tabs>
  <TabItem value="openadmin-notifications-view" label="With OpenAdmin" default>

  To view or edit current notification settings, go to **OpenAdmin > Settings > Notifications** or click the 'Edit Settings' button on the Notification page.
  
  ![openadmin notifications settings](/img/admin/openadmin_notifications_settings.png)

  </TabItem>
  <TabItem value="CLI-notifications-view" label="With OpenCLI">

To view notification settings, run:

```bash
opencli admin notifications get <OPTION>
```

Example:

```bash
# opencli admin notifications get reboot
yes
```

To update notification settings, run:

```bash
opencli admin notifications update <OPTION> <NEW-VALUE>
```

Example:

```bash
opencli admin notifications update load 10
Updated load to 10
```

  </TabItem>
</Tabs>

---

## Email

Configure email address to be used for receiving system notifications and alerts.

Enter your email address in the **Email address** field. Leave it empty to disable email alerts.

If email address is set, daily usage report will be sent to the address, schedule is configurable from 'Advanced > System Cron Jobs'.

![report example](/img/admin/daily_report.png)


---

## Webhook

Send notifications to a webhook URL (discord or any other).

---

## Services

Receive notifications when services are down or unresponsive. Services are checked every 5 minutes.

- **OpenPanel:** Notification if OpenPanel UI fails.
- **OpenAdmin:** Notification if OpenAdmin UI fails.
- **Caddy:** Notification if webserver is not responding.
- **MySQL:** Notification if database is unreachable.
- **Podman:** Notification if Podman service is down.
- **BIND9:** Notification if DNS service is down or unresponsive.
- **Sentinel Firewall:** Notification if Sentinel (CSF) is disabled.

---

## Resource Usage

Get alerts when resource usage exceeds thresholds (checked every 5 minutes):

* Load Average
* CPU %
* RAM %
* Disk Usage %
* SWAP %

---

## Server actions

Receive notifications when specific server-level actions are detected:

* **Server reboot:** Triggered when the server is restarted.
* **Unusual traffic or SYN flood:** Fires when suspicious traffic or DDoS attacks are detected. When enabled, an additional **Website traffic** section appears where you can set the **Max total connections** and **Max connections per IP** thresholds (on ports 80/443) that trigger the notification.
* **Out of Memory (OOM) errors:** Checks journal logs for system services and user processes killed by OOM in the last 24 hours.
* **DNS issue detected:** Triggered when the panel domain or nameservers are misconfigured or not resolving to this server. Disable if using external nameservers or a Cloudflare proxy.
* **OpenAdmin login from new IP:** Triggered when the OpenAdmin panel is accessed from an unrecognized IP address.
* **SSH login from new IP:** Triggered when root SSH access is detected from an unknown IP address. The IP can be whitelisted in the SSH Allowlist section below.
* **New update available:** Triggered when a new version of OpenPanel is available for update.

---

## User Actions

Get notified whenever an action occurs in the admin or user panels.

- OpenAdmin enabled/disabled
- API access enabled/disabled
- Admin account created
- Reseller account created
- Admin password changed
- Admin/Reseller renamed
- Admin/Reseller suspended
- Admin/Reseller unsuspended
- WAF enabled/disabled for a domain
- WAF enabled/disabled on the server
- User added
- User deleted
- User suspended/unsuspended
- User email changed
- User IP changed
- User password changed
- User renamed
- FTP account created
- FTP account deleted
- FTP account password change
- Domain added
- Domain deleted
- Domain suspended/unsuspended
- SSL type changed
- HSTS enabled/disabled

---

## SSH Allowlist

Specify IP addresses (or CIDRs) that will be exempt from SSH login checks.

---

## SMTP Settings

No SMTP server is configured by default — email notifications will **not** be sent until you set one up here.

To configure an SMTP server for email delivery, configure the following:

<Tabs>
  <TabItem value="openadmin-notifications-smtp" label="With OpenAdmin" default>
    Set server port, TLS or SSL, Username and Password to use for authentication.
  </TabItem>
  <TabItem value="CLI-notifications-smtp" label="With OpenCLI">

Configure each value via `opencli config update` commands, for example:

```bash
opencli config update mail_server example.net
```

```bash
opencli config update mail_port 465
```

```bash
opencli config update mail_use_tls False
```

```bash
opencli config update mail_use_ssl True
```

```bash
opencli config update mail_username user@example.net
```

```bash
opencli config update mail_password strongpassword123
```

```bash
opencli config update mail_default_sender user@example.net
```

  </TabItem>
</Tabs>
