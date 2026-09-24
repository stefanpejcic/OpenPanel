# Emails 

The following commands are available **for OpenPanel Enterprise** users:

## MailServer

Using the `opencli email-server` command you can install and manage the mailserver:

```bash
opencli email-server
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server
Usage:

opencli email-server status                           Show status
opencli email-server config                           Show configuration
opencli email-server install                          Install the email server  
opencli email-server start                            Start the email server
opencli email-server stop                             Stop the email server
opencli email-server restart                          Restart the email server
opencli email-server queue                            Show mail queue
opencli email-server flush                            Flush mail queue
opencli email-server view   <queue id>                Show mail by queue id
opencli email-server unhold <queue id> [<queue id>]   Release mail that was put "on hold" (marked with '!')
opencli email-server unhold ALL                       Release all mails that were put "on hold" (marked with '!')
opencli email-server delete <queue id> [<queue id>]   Delete mail from queue
opencli email-server delete ALL                       Delete all mails from queue
opencli email-server fail2ban [<ban|unban> <IP>]      Interact with fail2ban
opencli email-server fail2ban log                     Show fail2ban log
opencli email-server ports                            Show published ports
opencli email-server postconf                         Show postfix configuration
opencli email-server logs [-f]                        Show logs. Use -f to 'follow' the logs
opencli email-server login                            Run container shell
opencli email-server supervisor                       Interact with supervisorctl
opencli email-server pflogsumm                        Generate summary reports
opencli email-server postfwd                          Enable/disable postfwd rate-limiting and edit limits
opencli email-server uninstall                        Uninstall the email server
opencli email-server update-check                     Check for container package updates
opencli email-server update-packages                  Update container packages
opencli email-server versions                         Show versions
```
</details>

### Install

To install mailserver:

```bash
opencli email-server install
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server install
Installing the mailserver...
Cloning into '/usr/local/mail/openmail'...
...
```
</details>


### Start

To start the mailserver:

```bash
opencli email-server start
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server start
Starting mailserver...
MailServer started successfully.
```
</details>


### Restart

To restart the mailserver:

```bash
opencli email-server restart
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server restart
Restarting the mailserver...
MailServer stopped successfully.
MailServer started successfully.
```
</details>


### Stop

To stop the mailserver:

```bash
opencli email-server stop
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server stop
Stopping the mailserver...
MailServer stopped successfully.
```
</details>


### Status

To view current mailserver status:

```bash
opencli email-server status
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server status
Container:    Up 21 minutes (healthy)

Version:      v14.0.0


Packages:     Updates available

Ports:        25/tcp -> 0.0.0.0:25
              25/tcp -> [::]:25
              143/tcp -> 0.0.0.0:143
              143/tcp -> [::]:143
              465/tcp -> 0.0.0.0:465
              465/tcp -> [::]:465
              587/tcp -> 0.0.0.0:587
              587/tcp -> [::]:587
              993/tcp -> 0.0.0.0:993
              993/tcp -> [::]:993

Postfix:      Mail queue is empty

Supervisor:   amavis                           RUNNING   pid 910, uptime 0:20:23
              changedetector                   RUNNING   pid 949, uptime 0:20:23
              cron                             RUNNING   pid 803, uptime 0:20:25
              dovecot                          RUNNING   pid 813, uptime 0:20:25
              mailserver                       RUNNING   pid 8, uptime 0:21:06
              opendkim                         RUNNING   pid 848, uptime 0:20:24
              opendmarc                        RUNNING   pid 864, uptime 0:20:24
              postfix                          RUNNING   pid 876, uptime 0:20:24
              rsyslog                          RUNNING   pid 807, uptime 0:20:25
              update-check                     RUNNING   pid 822, uptime 0:20:24
              clamav                           STOPPED   Not started
              fail2ban                         STOPPED   Not started
              fetchmail                        STOPPED   Not started
              mta-sts-daemon                   STOPPED   Not started
              postgrey                         STOPPED   Not started
              postsrsd                         STOPPED   Not started
              rspamd                           STOPPED   Not started
              rspamd-redis                     STOPPED   Not started
              saslauthd_ldap                   STOPPED   Not started
              saslauthd_mysql                  STOPPED   Not started
              saslauthd_pam                    STOPPED   Not started
              saslauthd_rimap                  STOPPED   Not started
              saslauthd_shadow                 STOPPED   Not started
```
</details>




### Config

To view active mailserver configuration:

```bash
opencli email-server config
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server config
ACCOUNT_PROVISIONER='FILE'
AMAVIS_LOGLEVEL='0'
CLAMAV_MESSAGE_SIZE_LIMIT='25M'
DEFAULT_RELAY_HOST=''
DMS_VMAIL_GID='5000'
DMS_VMAIL_UID='5000'
DOVECOT_INET_PROTOCOLS='ipv4'
DOVECOT_MAILBOX_FORMAT='maildir'
DOVECOT_TLS='no'
ENABLE_AMAVIS='1'
ENABLE_CLAMAV='0'
ENABLE_DNSBL='0'
ENABLE_FAIL2BAN='0'
ENABLE_FETCHMAIL='0'
ENABLE_GETMAIL='0'
ENABLE_IMAP='1'
ENABLE_MANAGESIEVE='0'
ENABLE_OAUTH2='0'
ENABLE_OPENDKIM='1'
ENABLE_OPENDMARC='1'
ENABLE_POLICYD_SPF='1'
ENABLE_POP3='0'
ENABLE_POSTGREY='0'
ENABLE_QUOTAS='1'
ENABLE_RSPAMD='0'
ENABLE_RSPAMD_REDIS='0'
ENABLE_SASLAUTHD='0'
ENABLE_SPAMASSASSIN='0'
```
</details>

### Queue

To view current mail queue:

```bash
opencli email-server queue
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server queue
-Queue ID-  --Size-- ----Arrival Time---- -Sender/Recipient-------
4XbLqK2mPz*    1842 Wed Sep 24 10:12:05  wordpress@example.com
                                         customer@gmail.com

-- 2 Kbytes in 1 Request.
```
</details>

### Flush Queue

To flush current mail queue:

```bash
opencli email-server flush
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server flush
Queue flushed.
```
</details>

### View

To view mail by queue id:

```bash
opencli email-server view <queue id>
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server view 4XbLqK2mPz
*** ENVELOPE RECORDS deferred/4/4XbLqK2mPz ***
message_size:            1842             421               1               0            1842               0
message_arrival_time: Wed Sep 24 10:12:05 2026
sender: wordpress@example.com
...
*** MESSAGE CONTENTS deferred/4/4XbLqK2mPz ***
...
```
</details>


### Unhold

To release a specific mail that was put "on hold" (marked with '!'):
```bash
opencli email-server unhold <queue id> [<queue id>]
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server unhold 4XbLqK2mPz
postsuper: 4XbLqK2mPz: released from hold
postsuper: Released 1 message
```
</details>

To release ALL mails that are put "on hold" (marked with '!'):
```bash
opencli email-server unhold ALL
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server unhold ALL
postsuper: Released 3 messages
```
</details>

### Delete

To delete specific mail from queue:
```bash
opencli email-server delete <queue id>
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server delete 4XbLqK2mPz
postsuper: 4XbLqK2mPz: removed
postsuper: Deleted: 1 message
```
</details>

To delete multiple mails from wueue using their queue IDs :
```bash
opencli email-server delete <queue id> [<queue id>] [<queue id>] [<queue id>]
```

To delete ALL mails from queue:
```bash
opencli email-server delete ALL
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server delete ALL
postsuper: Deleted: 12 messages
```
</details>

### Fail2Ban

To interact with fail2ban:
```bash
opencli email-server fail2ban
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server fail2ban
Banned in custom:        198.51.100.23
Banned in postfix-sasl:  203.0.113.77
```
</details>

To ban IP address using fail2ban:
```bash
opencli email-server fail2ban ban <IP>
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server fail2ban ban 198.51.100.23
Banned custom IP: 1
```
</details>

To unban IP address using fail2ban:
```bash
opencli email-server fail2ban unban <IP>
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server fail2ban unban 198.51.100.23
Unbanned IP from postfix-sasl: 1
```
</details>

To view fail2ban logs:
```bash
opencli email-server fail2ban log
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server fail2ban log
2026-09-24 09:41:02,114 fail2ban.actions [512]: NOTICE  [postfix-sasl] Ban 203.0.113.77
...
```
</details>


### Postfix

To view current postfix configuration:
```bash
opencli email-server postconf
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server postconf
2bounce_notice_recipient = postmaster
access_map_defer_code = 450
access_map_reject_code = 554
address_verify_cache_cleanup_interval = 12h
address_verify_default_transport = $default_transport
address_verify_local_transport = $local_transport
address_verify_map = btree:$data_directory/verify_cache
address_verify_negative_cache = yes
address_verify_negative_expire_time = 3d
address_verify_negative_refresh_time = 3h
address_verify_pending_request_limit = 5000
address_verify_poll_count = ${stress?{1}:{3}}
address_verify_poll_delay = 3s
address_verify_positive_expire_time = 31d
address_verify_positive_refresh_time = 7d
...
```
</details>

### Logs

Display mailserver logs:
```bash
opencli email-server logs
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server logs
[   INF   ]  Welcome to docker-mailserver v14.0.0
[   INF   ]  Checking configuration
[   INF   ]  Configuring mail server
[   INF   ]  Starting daemons
[   INF   ]  mail.example.com is up and running
...
```
</details>

Use -f to 'follow' the logs:
```bash
opencli email-server logs -f
```

### pflogsumm

Generate HTML reports from mail logs:
```bash
opencli email-server pflogsumm
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server pflogsumm
Generating reports...
Generating email statistics reports.. This can take a while.
Done, adding reports to OpenAdmin interface
Completed
```
</details>

### Rate-limiting (postfwd)

Show whether outgoing email rate-limiting (postfwd) is enabled:
```bash
opencli email-server postfwd
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server postfwd
Rate-limiting is currently ENABLED.
```
</details>

Enable or disable rate-limiting:
```bash
opencli email-server postfwd enable
opencli email-server postfwd disable
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server postfwd enable
Rate-limiting has been ENABLED.

# opencli email-server postfwd disable
Rate-limiting has been DISABLED.
```
</details>

View or edit the rate-limiting rules (the mailserver is reloaded after editing):
```bash
opencli email-server postfwd view
opencli email-server postfwd edit
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server postfwd view
id=limit_stefan_example_com ; sender=~.+@example.com ; protocol_state==RCPT
                action=rate(stefan_ratelimit/100/3600/450 4.7.1 sorry, OpenPanel account reached limit of 100 emails per hour)
...
```
</details>

Rules for users and domains are generated with [`opencli email-ratelimit`](#rate-limits).

### Supervisor

Interact with the supervisor:
```bash
opencli email-server supervisor
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server supervisor
amavis                           RUNNING   pid 910, uptime 0:38:23
changedetector                   RUNNING   pid 949, uptime 0:38:23
clamav                           STOPPED   Not started
cron                             RUNNING   pid 803, uptime 0:38:25
dovecot                          RUNNING   pid 813, uptime 0:38:25
fail2ban                         STOPPED   Not started
fetchmail                        STOPPED   Not started
mailserver                       RUNNING   pid 8, uptime 0:39:06
mta-sts-daemon                   STOPPED   Not started
opendkim                         RUNNING   pid 848, uptime 0:38:24
opendmarc                        RUNNING   pid 864, uptime 0:38:24
postfix                          RUNNING   pid 876, uptime 0:38:24
postgrey                         STOPPED   Not started
postsrsd                         STOPPED   Not started
rspamd                           STOPPED   Not started
rspamd-redis                     STOPPED   Not started
rsyslog                          RUNNING   pid 807, uptime 0:38:25
saslauthd_ldap                   STOPPED   Not started
saslauthd_mysql                  STOPPED   Not started
saslauthd_pam                    STOPPED   Not started
saslauthd_rimap                  STOPPED   Not started
saslauthd_shadow                 STOPPED   Not started
update-check                     RUNNING   pid 822, uptime 0:38:24
```
</details>

### Login

Login as root user to the mailserver container:
```bash
opencli email-server login
```
Opens an interactive `bash` shell inside the mailserver container. Type `exit` to leave it.

### Ports

Display ports currently in use by the mailserver:
```bash
opencli email-server ports
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server ports
Published ports:

25/tcp -> 0.0.0.0:25
25/tcp -> [::]:25
143/tcp -> 0.0.0.0:143
143/tcp -> [::]:143
465/tcp -> 0.0.0.0:465
465/tcp -> [::]:465
587/tcp -> 0.0.0.0:587
587/tcp -> [::]:587
993/tcp -> 0.0.0.0:993
993/tcp -> [::]:993
```
</details>


### Update Check

Check for container package updates:
```bash
opencli email-server update-check
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server update-check
...
Listing... Done
libssl3/oldstable-security 3.0.15-1~deb12u1 amd64 [upgradable from: 3.0.14-1~deb12u2]
```
</details>

### Update Packages

Update container packages:
```bash
opencli email-server update-packages
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server update-packages
...
The following packages will be upgraded:
  libssl3 openssl
2 upgraded, 0 newly installed, 0 to remove and 0 not upgraded.
Do you want to continue? [Y/n]
```
</details>



### Versions

Display versions:
```bash
opencli email-server versions
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server versions
Mailserver:    v14.0.0

amavisd-new:   1:2.13.0-3
clamav:        1.0.3+dfsg-1~deb12u1
dovecot-core:  1:2.3.19.1+dfsg1-2.1
fail2ban:      1.0.2-2
fetchmail:     6.4.37-1
getmail6:      6.18.11-2
rspamd:        3.8.4-1~93fa4f6dc~bookworm
opendkim:      2.11.0~beta2-8+deb12u1
opendmarc:     1.4.2-2+b1
postfix:       3.7.10-0+deb12u1
spamassassin:  4.0.0-6
supervisor:    4.2.5-1
```
</details>


### Uninstall

To uninstall the mailserver:

```bash
opencli email-server uninstall
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-server uninstall
Uninstalling the mailserver...
Are you sure you want to uninstall the MailServer and remove all its configuration? (y/n)
y
MailServer uninstalled successfully.
```
</details>



### Debug

Add `--debug` to any `email-server` command to display verbose information.

## Webmail

Display current webmail domain:

```bash
opencli email-webmail
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-webmail
webmail.example.com
```
</details>

Set 'webmail.example.com' as a webmail domain:

```bash
opencli email-webmail domain webmail.example.com
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-webmail domain webmail.example.com
Webmail domain updated to https://webmail.example.com
```
</details>

Add `--debug` to display verbose information.

## Rate limits

`opencli email-ratelimit` generates postfwd rate-limiting rules for users and domains, based on the `max_hourly_email` limit of their plan.

Show current rules:
```bash
opencli email-ratelimit
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-ratelimit
=== /usr/local/mail/openmail/postfwd/postfwd.cf ===
id=limit_stefan_example_com ; sender=~.+@example.com ; protocol_state==RCPT
                action=rate(stefan_ratelimit/100/3600/450 4.7.1 sorry, OpenPanel account reached limit of 100 emails per hour)

---
3 lines total
```
</details>

Regenerate rules for all users:
```bash
opencli email-ratelimit --all-users
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-ratelimit --all-users
OK: stefan limit=100/hr domain=example.com
OK: stefan limit=100/hr domain=example.net
OK: demo limit=50/hr domain=demo.net
---
Generated 9 lines written to /usr/local/mail/openmail/postfwd/postfwd.cf
```
</details>

Regenerate rules for a single user (removes their rules, re-fetches their domains and adds them again):
```bash
opencli email-ratelimit --username=<USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-ratelimit --username=stefan
Updating rules for user: stefan
OK: stefan limit=100/hr domain=example.com
OK: stefan limit=100/hr domain=example.net
---
Done. 9 lines in /usr/local/mail/openmail/postfwd/postfwd.cf
```
</details>

Add or update the rule for a single domain:
```bash
opencli email-ratelimit --domain=<DOMAIN>
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-ratelimit --domain=example.com
Updating rule for domain: example.com
Updating: limit_stefan_example_com
OK: stefan limit=100/hr domain=example.com
---
```
</details>

Remove rules for a user or a domain:
```bash
opencli email-ratelimit --delete-user=<USERNAME>
opencli email-ratelimit --delete-domain=<DOMAIN>
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-ratelimit --delete-user=demo
Deleting all rules for user: demo
Removed all rules for user 'demo'.
---
Done. 6 lines in /usr/local/mail/openmail/postfwd/postfwd.cf

# opencli email-ratelimit --delete-domain=example.net
Deleting rule for domain: example.net
Removed rule: limit_stefan_example_net
---
Done. 3 lines in /usr/local/mail/openmail/postfwd/postfwd.cf
```
</details>

Add `--skip-reload` to not reload postfix after the change.

## Fix mail permissions

Set the correct owner on mail folders for all domains:
```bash
opencli email-quotas
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-quotas
2026-09-24 10:30:00 : Total mail domains in /home/*/mail/: [2]
2026-09-24 10:30:00 : [*] DOMAIN: example.com [1/2]
2026-09-24 10:30:00 : Checking username and docker context for domain..
...
2026-09-24 10:30:01 : Finished processing a total of 2 domains.
```
</details>

## Manage

Run a command inside the mailserver container:
```bash
opencli email-manage <COMMAND> [<ARGS>...]
```

Example:
```bash
opencli email-manage postqueue -p
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-manage postqueue -p
Mail queue is empty
```
</details>

## Emails

`opencli email-setup` command is used to create and manage email accounts. It runs the [docker-mailserver `setup`](https://docker-mailserver.github.io/docker-mailserver/latest/config/setup.sh/) command inside the mailserver container.

Commands that add, change or delete email accounts, aliases, quotas, restrictions, relay settings or Dovecot master accounts print no output when they succeed. If the mailserver is not running, the change is written directly to the mailserver configuration files and this warning is shown:

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup email add info@example.com 'StrongPass1'
Warning: 'openadmin_mailserver' is not running - writing directly to mailserver config files instead.
```
</details>

To view a list of all available sub-commands:

```bash
opencli email-setup help
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup help
SETUP(1)

NAME
    setup - 'docker-mailserver' Administration & Configuration CLI

SYNOPSIS
    setup [ OPTIONS... ] COMMAND [ help | ARGUMENTS... ]

    COMMAND := { email | alias | quota | dovecot-master | config | relay | fail2ban | debug } SUBCOMMAND
...
```
</details>

### List emails

To view a list of all email addresses on server:

```bash
opencli email-setup email list
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup email list
* stefan@example.com ( 5.4M / 1G ) [0%]
* info@example.com ( 12K / ~ ) [0%]
```
</details>

### Create email

To create new email account:

```bash
opencli email-setup email add <EMAIL ADDRESS> [<PASSWORD>]
```

### Change password

To change password for email account:

```bash
opencli email-setup email update <EMAIL ADDRESS> [<PASSWORD>]
```

### Delete email

To delete an email account:

```bash
opencli email-setup email del [ OPTIONS... ] <EMAIL ADDRESS> [ <EMAIL ADDRESS>... ]
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup email del info@example.com
Do you want to delete the mailbox as well (removing all mails)? [Y/n] y
```
</details>


### Restrict email

To suspend sending or receving of emails for an email account:

```bash
opencli email-setup email restrict <add|del|list> <send|receive> [<EMAIL ADDRESS>]
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup email restrict list send
info@example.com		REJECT
```
</details>



## Aliases

The `opencli email-setup alias` command is used to manage email aliases:



### List Aliases

To list all email aliases:
```bash
opencli email-setup alias list
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup alias list
* sales@example.com stefan@example.com
* contact@example.com info@example.com
```
</details>

### Add Alias

To create a new email alias:
```bash
opencli email-setup alias add <EMAIL ADDRESS> <RECIPIENT>
```


### Delete Alias

To delete existing email alias:
```bash
opencli email-setup alias del <EMAIL ADDRESS> <RECIPIENT>
```

## Quotas

The `opencli email-setup quota` command is used to manage email quotas:

### Set Quota

To set quota for email account:

```bash
opencli email-setup quota set <EMAIL ADDRESS> [<QUOTA>]
```

### Remove Quota

To remove quota for email account:

```bash
opencli email-setup quota del <EMAIL ADDRESS>
```



## Dovecot Master

[Dovecot Master](https://doc.dovecot.org/2.3/configuration_manual/authentication/master_users/) accounts are used to auto-login from OpenPanel interface to any email address.

`opencli email-setup dovecot-master` 


### List Dovecot Master 

To list all dovecot-master accounts:

```bash
opencli email-setup dovecot-master list
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup dovecot-master list
* admin
```
</details>


### Add Dovecot Master 

To create a new dovecot-master account:

```bash
opencli email-setup dovecot-master add <USERNAME> [<PASSWORD>]
```

### Update Dovecot Master 

To chaneg password for existing dovecot-master account:

```bash
opencli email-setup dovecot-master update <USERNAME> [<PASSWORD>]
```

### Add Dovecot Master 

To delete an dovecot-master account:

```bash
opencli email-setup dovecot-master del [ OPTIONS... ] <USERNAME> [ <USERNAME>... ]
```





## DKIM

To manage DKIM use `opencli email-setup config` command:

### Setup DKIM

To setup DKIM:

```bash
opencli email-setup config dkim [ ARGUMENTS... ]
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup config dkim
Creating DKIM private key '/tmp/docker-mailserver/opendkim/keys/example.com/mail.private'
Creating DKIM KeyTable
Creating DKIM SigningTable
Creating DKIM TrustedHosts
```
</details>

## Email Relay

To setup [email relay](https://en.wikipedia.org/wiki/Open_mail_relay) use `opencli email-setup relay` command:


### Add Auth

To add auth to email relay:

```bash
opencli email-setup relay add-auth <DOMAIN> <USERNAME> [<PASSWORD>]
```

### Add Domain

To add auth domain to email relay:

```bash
opencli email-setup relay add-domain <DOMAIN> <HOST> [<PORT>]
```

### Exclude Auth

To exclude auth from email relay:

```bash
opencli email-setup relay exclude-domain <DOMAIN>
```




## Fail2Ban

[Fail2Ban](https://github.com/fail2ban/fail2ban) is used for restricting and blocking access to email accounts.

To display available options:
```bash
opencli email-setup fail2ban
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup fail2ban
Banned in custom:        198.51.100.23
Banned in postfix-sasl:  203.0.113.77
```
</details>

### Ban IP

To ban IP address from accessing mail server:
```bash
opencli email-setup fail2ban ban <IP>
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup fail2ban ban 198.51.100.23
Banned custom IP: 1
```
</details>

### UnBan IP

To unban IP address and allow it to access mail server:
```bash
opencli email-setup fail2ban unban <IP>
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup fail2ban unban 198.51.100.23
Unbanned IP from postfix-sasl: 1
```
</details>

### View Log

To view the fail2ban log:
```bash
opencli email-setup fail2ban log
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup fail2ban log
2026-09-24 09:41:02,114 fail2ban.actions [512]: NOTICE  [postfix-sasl] Ban 203.0.113.77
...
```
</details>


### View Status

To view the fail2ban status:
```bash
opencli email-setup fail2ban status
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup fail2ban status
Status
|- Number of jail:	3
`- Jail list:	custom, dovecot, postfix-sasl
```
</details>

## Debug

For debugging emails the following commands are available:

### Fetchmail

[Fetchmail](https://www.fetchmail.info/fetchmail-man.html) is a powerful tool that can be used to debug email:

```bash
opencli email-setup debug fetchmail
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup debug fetchmail
fetchmail: No mail for stefan at pop.example.com
```
</details>

### Login

To troubleshoot emaail address login:

```bash
opencli email-setup debug login <COMMANDS>
```
Runs the given commands inside the mailserver container, or opens an interactive shell when no command is given.

### Mail Logs

To view the mailserver logs:
```bash
opencli email-setup debug show-mail-logs
```

<details>
  <summary>Example output</summary>

```bash
# opencli email-setup debug show-mail-logs
Sep 24 10:12:05 mail postfix/smtpd[1834]: connect from unknown[203.0.113.5]
Sep 24 10:12:06 mail postfix/qmgr[876]: 4XbLqK2mPz: from=<wordpress@example.com>, size=1842, nrcpt=1 (queue active)
Sep 24 10:12:07 mail postfix/smtp[1840]: 4XbLqK2mPz: to=<customer@gmail.com>, relay=gmail-smtp-in.l.google.com[142.250.27.26]:25, delay=1.4, status=sent (250 2.0.0 OK)
...
```
</details>

