---
sidebar_position: 3
---

# Notifications

The **Account > Notifications** page lets users pick which events they get an email about. Each event has its own card with a switch, and some cards have extra options that show up once the switch is on.

![Email Notifications page with a card and switch for each event and the Save Preferences button](/img/openpanel-screenshots/account/notifications-form.png#gh-light-mode-only)
![Email Notifications page with a card and switch for each event and the Save Preferences button](/img/openpanel-screenshots/account/notifications-form_dark.png#gh-dark-mode-only)

:::info
If you do not see the Notifications page, ask your provider to enable the [notifications module](/docs/admin/settings/modules/#notifications).

Emails go to the contact email address set on the [Account](/docs/panel/account/) page. They're sent through OpenAdmin using the SMTP settings from **OpenAdmin > Settings > Notifications**, so if nothing arrives, ask your provider to check those.

Changes made through the [OpenPanel API](/docs/panel/api/) do send the same emails as changes made from the panel.
:::

## Security alerts

Emails about changes to the account (logins, password, email address, 2FA, passkeys and API tokens) show when the change was made, the IP address, the country with its flag, the browser and operating system, and what to do if it wasn't you.

### New login

Sends an email every time someone logs in to the account, from the panel or the API. Besides the details above, it shows how they logged in: password, password and 2FA code, passkey or the API. The country is looked up from the IP address and shows `UNKNOWN` if the lookup fails.

![New login card with the Email me switch on and its two extra options, also for known IP addresses and email me if login alerts get turned off](/img/openpanel-screenshots/account/notifications-login.png#gh-light-mode-only)
![New login card with the Email me switch on and its two extra options, also for known IP addresses and email me if login alerts get turned off](/img/openpanel-screenshots/account/notifications-login_dark.png#gh-dark-mode-only)

Here is an example of the email.

![New login email showing the login method, time, IP address, country with its flag and browser](/img/openpanel-screenshots/account/notifications-email-login.png)

By default, logins from an IP address that's already in your login history don't send an email, so you only hear about logins from new places. Turn on **Also for IP addresses I logged in from before** to get an email for every login.

If you didn't log in yourself, change your password right away and turn on [two-factor authentication](/docs/panel/security/2fa/).

### Password changed

Sends an email when the account password is changed, and says whether it was changed from the panel or the API. If you didn't change it, reset your password or contact your provider.

![Password changed card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-password.png#gh-light-mode-only)
![Password changed card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-password_dark.png#gh-dark-mode-only)

![Password changed email saying it was changed from the OpenPanel interface, with the time, IP address, country and browser](/img/openpanel-screenshots/account/notifications-email-password.png)

### Contact email changed

Sends an email when the contact email address for the account is changed. The email shows the old and the new address and goes to the **old** one, so the real owner knows about it even if someone else changed it.

![Contact email changed card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-contact-email.png#gh-light-mode-only)
![Contact email changed card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-contact-email_dark.png#gh-dark-mode-only)

![Email address changed email showing the old and new address, the time, IP address, country and browser](/img/openpanel-screenshots/account/notifications-email-contact-email.png)

### Two-factor authentication changed

Sends an email when two-factor authentication is turned on or off, or when 2FA setup is started.

![Two-factor authentication changed card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-twofa.png#gh-light-mode-only)
![Two-factor authentication changed card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-twofa_dark.png#gh-dark-mode-only)

When 2FA is turned on:

![Two-factor authentication enabled email saying a code from the authenticator app is now needed on every login, with the time, IP address, country and browser](/img/openpanel-screenshots/account/notifications-email-2fa-enabled.png)

When 2FA is turned off:

![Two-factor authentication disabled email saying only the password is needed to log in, with the time, IP address, country and browser](/img/openpanel-screenshots/account/notifications-email-2fa-disabled.png)


### Passkey added or removed

Sends an email when a [passkey](/docs/panel/security/passkeys/) is added to the account or removed from it, from the panel or the API. The email includes the passkey's name.

![Passkey added or removed card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-passkey.png#gh-light-mode-only)
![Passkey added or removed card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-passkey_dark.png#gh-dark-mode-only)

When a passkey is added:

![New passkey email with the passkey name MacBook Touch ID, the time, IP address, country and browser](/img/openpanel-screenshots/account/notifications-email-passkey.png)

When a passkey is removed:

![Passkey removed email with the passkey name MacBook Touch ID, the time, IP address, country and browser](/img/openpanel-screenshots/account/notifications-email-passkey-removed.png)

### API token created or revoked

Sends an email when a token for the [AI Assistant (MCP)](/docs/panel/account/mcp/) is created or revoked, from the panel or the API. The email includes the token's name, and for a new token whether it has full or read-only access and when it expires. Anyone with the token can manage the account, so treat it like a password.

![API token created or revoked card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-api-token.png#gh-light-mode-only)
![API token created or revoked card with the Email me switch and the option to get an email if this alert gets turned off](/img/openpanel-screenshots/account/notifications-api-token_dark.png#gh-dark-mode-only)

![New API token email with the token name Claude Desktop, read-only access that expires in 30 days, the time, IP address, country and browser](/img/openpanel-screenshots/account/notifications-email-api-token.png)

### SSL certificate problem

Checks the SSL certificate of every domain that points to this server once a day, and sends an email when:
- an AutoSSL certificate has 7 days or less left. AutoSSL certificates are renewed about 30 days before they expire, so this means renewal is failing. The email includes the last error from the web server, for example a DNS record pointing elsewhere.
- any certificate, AutoSSL or custom, has 1 day or less left.

Each alert is sent once per certificate, and all affected domains of the account are listed in one email. Domains that don't point to this server are skipped.

![SSL certificate problem card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-ssl.png#gh-light-mode-only)
![SSL certificate problem card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-ssl_dark.png#gh-dark-mode-only)

![SSL certificate problem email listing an AutoSSL certificate that fails to renew with the error, and a custom certificate that expires in less than a day](/img/openpanel-screenshots/account/notifications-email-ssl.png)

### Malware found

Sends an email when the scheduled malware scan finds infected files and moves them to quarantine. The email lists up to 20 files with the malware signature found in each, and points to the **Malware Scanner > Quarantine** page to delete or restore them.

Scans you start yourself from the Malware Scanner page don't send an email, since you see the results right away.

![Malware found card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-malware.png#gh-light-mode-only)
![Malware found card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-malware_dark.png#gh-dark-mode-only)

![Malware found email listing two quarantined files with their malware signatures, when the scan ran and what to do next](/img/openpanel-screenshots/account/notifications-email-malware.png)

### Email me if this alert gets turned off

Every security alert except **SSL certificate problem** and **Malware found** has this option, and it's on by default. When it's on and someone turns the alert itself off, an email is sent listing the alerts that were turned off. This way nobody can quietly turn off your login or password alerts before doing something with your account.

To really stop an alert, turn off this option first, save, and then turn off the alert.

![Notification preferences changed email listing New login and Password changed as turned off, with the time, IP address, country and browser](/img/openpanel-screenshots/account/notifications-email-alert-disabled.png)

## Usage alerts

### Disk space running out

Sends an email when the account uses 85% of its disk space or 95% of its inodes (number of files). The email shows disk and inode usage as bars against the plan limits, marks the one over the limit, and suggests the Disk Usage and Inodes Explorer pages to find what takes up space.

![Disk space running out card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-disk.png#gh-light-mode-only)
![Disk space running out card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-disk_dark.png#gh-dark-mode-only)

Usage is checked every 5 minutes. The email is sent once, and again only after usage drops below the limit and crosses it again, so you don't get the same email every few minutes. Your provider is also told about it.

Once the limit is reached, websites and email can stop working. Delete files you no longer need or ask your provider for a bigger plan.

![Disk space email with usage bars for disk space at 87 percent over the 85 percent limit and inodes at 42 percent, tips to free up space and an Upgrade to Business section](/img/openpanel-screenshots/account/notifications-email-disk.png)

### Mailbox almost full

Sends an email when one of your email accounts uses 90% of its quota, showing a usage bar for each account that is almost full.

![Mailbox almost full card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-mailbox.png#gh-light-mode-only)
![Mailbox almost full card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-mailbox_dark.png#gh-dark-mode-only) It's sent once per email account, and again only after usage drops below 90% and reaches it again.

When a mailbox is full, new emails to it are rejected. Delete old emails or raise the quota on the [Email Accounts](/docs/panel/emails/) page.

![Mailbox email with usage bars for info@example.com at 92 percent and a full sales@example.com, tips to free up space and an Upgrade to Business section](/img/openpanel-screenshots/account/notifications-email-mailbox.png)

### Hourly email limit reached

Sends an email when the account reaches the hourly email sending limit of its hosting plan and new emails are rejected. The limit counts emails from all domains on the account together. The email shows how many emails were rejected, when, and from which addresses.

A sudden spike often means a mailbox password was stolen or a contact form is abused by spammers, so the email also says what to check.

Rejections are checked every 30 minutes, and the email is sent at most once a day.

![Hourly email limit reached card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-ratelimit.png#gh-light-mode-only)
![Hourly email limit reached card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-ratelimit_dark.png#gh-dark-mode-only)

![Hourly email limit email with a full usage bar, rejected emails by sender, an Upgrade to Business section and tips on what to check](/img/openpanel-screenshots/account/notifications-email-ratelimit.png)

### Service stopped

Sends an email when a service on the account, like MySQL or PHP, had stopped and the server's hourly automatic check had to start it again. The email lists each service, whether it ran out of memory, and whether it could be started again.

The email is sent at most once a day.

![Service stopped card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-service.png#gh-light-mode-only)
![Service stopped card with the Email me switch and a Learn more link](/img/openpanel-screenshots/account/notifications-service_dark.png#gh-dark-mode-only)

![Services restarted email listing MySQL that ran out of memory and PHP, both restarted, with tips for services that keep stopping](/img/openpanel-screenshots/account/notifications-email-service.png)

### Upgrade offer

On Enterprise, if the account's hosting plan has an [upsell plan](/docs/admin/plans/hosting_plans/) with an upgrade URL, the disk space, mailbox and hourly email limit emails also get an **Upgrade to** section named after the upsell plan, with a button to **Dashboard > Upgrade**. It's only added when the upsell plan actually raises the limit that's running out: more disk space or inodes for the disk email, bigger maximum mailbox size for the mailbox email, and more emails per hour for the hourly limit email.
