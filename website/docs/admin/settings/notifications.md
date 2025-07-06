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

To receive email alerts, enter your email address in the **Email for notifications** field. Leave it empty to disable email alerts.

Providing an email will also enable daily Usage Reports:

![image](/img/admin/daily_report.png)

---

## Services

Receive notifications when services are down or unresponsive. Services are checked every 5 minutes.

- **OpenPanel:** Notification if OpenPanel UI fails.
- **OpenAdmin:** Notification if OpenAdmin UI fails.
- **Caddy:** Notification if webserver is not responding.
- **MySQL:** Notification if database is unreachable.
- **Docker:** Notification if Docker service is down.
- **BIND9:** Notification if DNS service is down or unresponsive.
- **ConfigServer Firewall:** Notification if CSF is disabled.

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

  - **Server:** Your SMTP server domain or IP
  - **Port:** Outgoing SMTP port (default: 465)
  - **Use TLS:** Default is False (SSL is used instead)
  - **Use SSL:** Default is True
  - **Username:** Email address used for sending
  - **Password:** Password for the email account
  - **Default sender:** Email displayed as sender (defaults to username)
  - **Store Emails:** Enable to save sent emails in `/var/log/openpanel/admin/emails`

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
