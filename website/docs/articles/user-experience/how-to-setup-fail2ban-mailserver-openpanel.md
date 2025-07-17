# Setup Fail2ban for Mailserver

Ensure that you're running the [**Enterprise edition**](https://openpanel.com/enterprise/) of OpenPanel. Email support is only available in this version.

Follow [this guide](/docs/articles/user-experience/how-to-setup-email-in-openpanel/) to enable Emails in OpenPanel.

---

Fail2ban is an intrusion prevention software framework. Written in the Python programming language, it is designed to prevent against brute-force attacks. It is able to run on POSIX systems that have an interface to a packet-control system or firewall installed locally, such as NFTables or TCP Wrapper. [More Information](https://en.wikipedia.org/wiki/Fail2ban)

Enabling Fail2Ban support requires granting at least the `NET_ADMIN` capability to the mailserver in order to interact with the kernel and ban IP addresses.


## 1. Enable Fail2ban

To enable Fail2ban, navigate to **OpenAdmin > Emails > Settings** and under 'Enable Services' select **fail2ban** then click on 'Save Stack'.

[![2025-07-17-14-33.png](https://i.postimg.cc/G3q4ByVh/2025-07-17-14-33.png)](https://postimg.cc/qg6JSzp9)

---

## 2. Recreate mailserver

Fail2ban requires that the mailserver container is stopped, and start again in order to activate. Navigate to **OpenAdmin > Services > Status** and click 'Stop' for mailserver, then again 'Start'.

[![2025-07-17-15-37.png](https://i.postimg.cc/d3hgY01F/2025-07-17-15-37.png)](https://postimg.cc/ctNF70wk)

---

## Default Configuration

Mailserver will automatically ban IP addresses of hosts that have generated **6** failed attempts over the course of the last week. The bans themselves last for one week. The Postfix jail is configured to use `mode = extra`.

If created, the following configuration files will be copied inside the container during startup. To customize the configuration, simply create a new file and customize the examples:

- `/usr/local/mail/openmail/docker-data/dms/config/fail2ban-jail.cf` - adjust the configuration of individual jails and their defaults, [example file](https://github.com/docker-mailserver/docker-mailserver/blob/master/config-examples/fail2ban-jail.cf)
- `/usr/local/mail/openmail/docker-data/dms/config/fail2ban-fail2ban.cf` - adjust F2B behavior in general, [example file](https://github.com/docker-mailserver/docker-mailserver/blob/master/config-examples/fail2ban-fail2ban.cf)

**Example:**

To modify the number of failed logins before IP is blocked (default: 6), or change the bantime (default: 1w):

- create file `/usr/local/mail/openmail/docker-data/dms/config/fail2ban-jail.cf`
- in it, paste [this example](https://github.com/docker-mailserver/docker-mailserver/blob/master/config-examples/fail2ban-jail.cf)
- change the `maxretry` and `bantime` values:

```yaml
[DEFAULT]

# "bantime" is the number of seconds that a host is banned.
bantime = 1w

# A host is banned if it has generated "maxretry" during the last "findtime"
# seconds.
findtime = 1w

# "maxretry" is the number of failures before a host get banned.
maxretry = 6

# "ignoreip" can be a list of IP addresses, CIDR masks or DNS hosts. Fail2ban
# will not ban a host which matches an address in this list. Several addresses
# can be defined using space (and/or comma) separator.
ignoreip = 127.0.0.1/8

# default ban action
# nftables-multiport: block IP only on affected port
# nftables-allports:  block IP on all ports
banaction = nftables-allports

[dovecot]
enabled = true

[postfix]
enabled = true
# For a reference on why this mode was chose, see
# https://github.com/docker-mailserver/docker-mailserver/issues/3256#issuecomment-1511188760
mode = extra

[postfix-sasl]
enabled = true

# This jail is used for manual bans.
# To ban an IP address use: setup.sh fail2ban ban <IP>
[custom]
enabled = true
bantime = 180d
port = smtp,pop3,pop3s,imap,imaps,submission,submissions,sieve
```

---

## View and Manage Bans

- `opencli email-setup fail2ban` - shows all banned IP addresses
- `opencli email-setup fail2ban status` - shows more detailed status
- `opencli email-setup fail2ban ban <IP>` - ban an IP address
- `opencli email-setup fail2ban unban <IP>` - unban an IP address
- `opencli email-setup fail2ban log` - view fail2ban log file
