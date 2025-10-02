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

Enter your email address in the **Email for notifications** field. Leave it empty to disable email alerts.

If email address is set, daily usage report will be sent to the address, schedule is configurable from 'Advanced > System Cron Jobs'.

![report example](/img/admin/daily_report.png)


---

## Services

Receive notifications when services are down or unresponsive. Services are checked every 5 minutes.

- **OpenPanel:** Notification if OpenPanel UI fails.
- **OpenAdmin:** Notification if OpenAdmin UI fails.
- **Caddy:** Notification if webserver is not responding.
- **MySQL:** Notification if database is unreachable.
- **Docker:** Notification if Docker service is down.
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

## Actions

Receive notifications when specific actions occur:

* Server rebooted
* Website under attack
* User reaches plan limit
* OpenAdmin accessed from a new IP address
* New OpenPanel update available

---

## SMTP Settings

By default, email alerts are sent from `noreply@openpanel.com`.

To use your own SMTP server for email delivery, configure the following:

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
