# Emails 

Starting from the [0.2.5](https://openpanel.com/docs/changelog/0.2.5/) version, [OpenPanel Enterprise edition](https://openpanel.com/beta/) includes a mailserver.

The following commands are available **for OpenPanel Enterprise** users:

## MailServer

Using the `opencli email-server` command you can install and manage the mailserver:

```bash
root@pejcic:~# opencli email-server
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
opencli email-servers install                         Generate summary reports
opencli email-server update-check                     Check for container package updates
opencli email-server update-packages                  Update container packages
opencli email-server versions                         Show versions
```

### Install

To install mailserver:

```bash
opencli email-server install
```


### Start

To start the mailserver:

```bash
opencli email-server start
```


### Restart

To restart the mailserver:

```bash
opencli email-server restart
```


### Stop

To stop the mailserver:

```bash
opencli email-server stop
```


### Status

To view current mailserver status:

```bash
opencli email-server status
```

Example:
```bash
root@stefi:/usr/# opencli email-server status
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




### Config

To view active mailserver configuration:

```bash
opencli email-server config
```

Example:
```bash
root@stefi:~# opencli email-server config                           
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

### Queue

To view current mail queue:

```bash
opencli email-server queue
```

### Flush Queue

To flush current mail queue:

```bash
opencli email-server flush
```

### View

To view mail by queue id:

```bash
opencli email-server view <queue id>
```


### Unhold

To release a specific mail that was put "on hold" (marked with '!'):
```bash
opencli email-server unhold <queue id> [<queue id>]
```

To release ALL mails that are put "on hold" (marked with '!'):
```bash
opencli email-server unhold ALL
```

### Delete

To delete specific mail from queue:
```bash
opencli email-server delete <queue id>
```

To delete multiple mails from wueue using their queue IDs :
```bash
opencli email-server delete <queue id> [<queue id>] [<queue id>] [<queue id>]
```

To delete ALL mails from queue:
```bash
opencli email-server delete ALL
```

### Fail2Ban

To interact with fail2ban:
```bash
opencli email-server fail2ban
```

To ban IP address using fail2ban:
```bash
opencli email-server fail2ban ban <IP>
```

To unban IP address using fail2ban:
```bash
opencli email-server fail2ban unban <IP>
```

To view fail2ban logs:
```bash
opencli email-server fail2ban log
```


### Postfix

To view current postfix configuration:
```bash
opencli email-server postconf
```

Example:
```bash
root@stefi:~# opencli email-server postconf
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

### Logs

Display mailserver logs:
```bash
opencli email-server logs
```

Use -f to 'follow' the logs:
```bash
opencli email-server logs -f
```

### pflogsumm

Generate HTML reports from mail logs:
```bash
opencli email-server pflogsumm
```

### Supervisor

Interact with the supervisor:
```bash
opencli email-server supervisor
```

Example:
```bash
root@stefi:~# opencli email-server supervisor
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

### Login

Login as root user to the mailserver container:
```bash
opencli email-server login
```

### Ports

Display ports currently in use by the mailserver:
```bash
opencli email-server ports
```

Example:
```bash
root@stefi:~# opencli email-server ports
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


### Update Check

Check for container package updates:
```bash
opencli email-server update-check
```

### Update Packages

Update container packages:
```bash
opencli email-server update-packages
```



### Versions

Display versions:
```bash
opencli email-server versions
```

Example:
```bash
root@stefi:~# opencli email-server versions
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


### Uninstall

To uninstall the mailserver:

```bash
opencli email-server uninstall
```



## Webmail

Display current webmail domain:

```bash
opencli email-webmail
```

Set 'webmail.example.com' as a webmail domain:

```bash
opencli email-webmail domain webmail.example.com
```

## Emails

`opencli email-setup` command is used to create and manage email accounts.

To view a list of all available sub-commands:

```bash
opencli email-setup help
```

### List emails

To view a list of all email addresses on server:

```bash
opencli email-setup email list
```

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


### Restrict email

To suspend sending or receving of emails for an email account:

```bash
opencli email-setup email restrict <add|del|list> <send|receive> [<EMAIL ADDRESS>]
```



## Aliases

The `opencli email-setup alias` command is used to manage email aliases:



### List Aliases

To list all email aliases:
```bash
opencli email-setup alias list
```

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

### Ban IP

To ban IP address from accessing mail server:
```bash
opencli email-setup fail2ban ban <IP>
```

### UnBan IP

To unban IP address and allow it to access mail server:
```bash
opencli email-setup fail2ban unban <IP>
```

### View Log

To view the fail2ban log:
```bash
opencli email-setup fail2ban log
```


### View Status

To view the fail2ban status:
```bash
opencli email-setup fail2ban status
```

## Debug

For debugging emails the following commands are available:

### Fetchmail

[Fetchmail](https://www.fetchmail.info/fetchmail-man.html) is a powerful tool that can be used to debug email:

```bash
opencli email-setup debug fetchmail
```

### Login

To troubleshoot emaail address login:

```bash
opencli email-setup debug login <COMMANDS>
```

### Mail Logs

To view the mailserver logs:
```bash
opencli email-setup debug show-mail-logs
```
