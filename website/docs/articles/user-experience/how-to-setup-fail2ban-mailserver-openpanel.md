# Setup Fail2ban for Mailserver


Ensure that you're running the [**Enterprise edition**](https://openpanel.com/enterprise/) of OpenPanel. Email support is only available in this version.

Follow [this guide](/docs/articles/user-experience/how-to-setup-email-in-openpanel/) to enable Emails in OpenPanel.

---

Fail2ban is an intrusion prevention software framework. Written in the Python programming language, it is designed to prevent against brute-force attacks. It is able to run on POSIX systems that have an interface to a packet-control system or firewall installed locally, such as [NFTables] or TCP Wrapper. [More Information](https://en.wikipedia.org/wiki/Fail2ban)

Enabling Fail2Ban support requires granting at least the `NET_ADMIN` capability to the mailserver in order to interact with the kernel and ban IP addresses.


## 1. Enable Fail2ban

To enable Fail2ban, navigate to **OpenAdmin > Emails > Settings** and under 'Enable Services' select **fail2ban** then click on 'Save Stack'.

[![2025-07-17-14-33.png](https://i.postimg.cc/G3q4ByVh/2025-07-17-14-33.png)](https://postimg.cc/qg6JSzp9)

---

## 2. Recreate mailserver

Fail2ban requires that the mailserver container is stopped, and start again in order to activate. Navigate to **OpenAdmin > Services > Status** and click 'Stop' for mailserver, then again 'Start'.

---

## Default Configuration

Mailserver will automatically ban IP addresses of hosts that have generated **6** failed attempts over the course of the last week. The bans themselves last for one week. The Postfix jail is configured to use `mode = extra`.

This following configuration files inside the docker-data/dms/config/ volume will be copied inside the container during startup

- `docker-data/dms/config/fail2ban-jail.cf` - you can adjust the configuration of individual jails and their defaults, [example file](https://github.com/docker-mailserver/docker-mailserver/blob/master/config-examples/fail2ban-jail.cf)
- `docker-data/dms/config/fail2ban-fail2ban.cf` - you can adjust F2B behavior in general, [example file](https://github.com/docker-mailserver/docker-mailserver/blob/master/config-examples/fail2ban-fail2ban.cf)

## 
