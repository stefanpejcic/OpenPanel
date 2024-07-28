---
sidebar_position: 1
---

# Get Started with OpenAdmin

The OpenAdmin offers an administrator-level interface where you can efficiently handle tasks such as creating and managing users, setting up hosting plans, configuring backups, and editing OpenPanel settings.

![openpanel vs openadmin](/img/admin/openadmin_vs_openpanel_what_is_the_difference.png)



## Requirements

Mimumum Requirements:

- A blank full virtual machine or bare metal server
- Debian 11 / Debian 12 / Ubuntu 22 / Ubuntu 24*(without quotas)
- Minimum of 1GB RAM and 15GB storage
- x86_64/amd64 architecture

:::info
If you are using external firewall, the following ports should be opened:  `53` `80` `443` `2083` `2087` `32768:60999`
:::

## Installation

The installation process takes about 5 minutes. To install openpanel follow these steps: 



<Tabs>
  <TabItem value="openpanel-install-on-dedicated" label="Manual Installation" default>

1. Log in to your new server;
- as root via SSH or
- as a user with sudo privileges and type "sudo -i"
2. Copy and paste openpanel installation command into the terminal
```shell
bash <(curl -sSL https://get.openpanel.co/)
```

The installation script supports [optional flags](/install) that can be used to configure openpanel, skip certain installation steps or simply display debugging information.

If you encountered any errors while running the installation script, please copy & paste the installation log file to [the community forums](https://community.openpanel.co).

  </TabItem>
  <TabItem value="openpanel-install-on-digitalocean" label="DigitalOcean Droplet">

OpenPanel is available as a 1-Click app (droplet) on DigitalOcean. Click on the button to spin a droplet with OpenPanel already installed:

[![droplet](/img/do-btn-blue.svg)](https://marketplace.digitalocean.com/apps/openpanel?refcode=6498bfc47cd6&action=deploy)

or with DigitalOcean api:

```bash
curl -X POST -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer '$TOKEN'' -d \
    '{"name":"choose_a_name","region":"nyc3","size":"s-2vcpu-4gb","image":"openpanel"}' \
    "https://api.digitalocean.com/v2/droplets"
```

  </TabItem>
</Tabs>


## Post Install Steps

- [access admin panel](/docs/admin/intro#access-openadmin)
- [set custom nameservers](/docs/admin/intro#set-nameservers)
- [create a package](/docs/admin/plans/hosting_plans#create-a-plan)
- [create a new user account](/docs/admin/users/openpanel#create-users)

### Access OpenAdmin

Run `opencli admin` command to find the address on which admin panel is accessible. Example output:

```bash
root@server:/home# opencli admin
● AdminPanel is running and is available on: https://server.openpanel.co:2087/
```

To login to admin panel you need a username and password.

![openadmin login page](/img/admin/openadmin_login_page.png)

Both username and password are random generated on installation.

To view admin accounts:

```bash
opencli admin list
```

To set a new password for the admin account run command: `opencli admin password USER_HERE NEW_PASSWORD_HERE`

Example:
```bash
root@server:/home# opencli admin password stefan ba63vfav7fq36vas
Password for user 'stefan' changed.

===============================================================
● AdminPanel is running and is available on: https://server.openpanel.co:2087/

- username: stefan
- password: ba63vfav7fq36vas

===============================================================
```

### Enable Automatic Updates

- `autopatch` option allows Administrator to automatically update OpenPanel to minor versions. MINOR versions include only security updates and bug fixes.
- `autoupdate` option allows Aministrator to enable or disable automatic updates to major versions. MAJOR versions add new functionality in a backward compatible manner.

<Tabs>
  <TabItem value="openadmin-admin-updates" label="With OpenAdmin" default>

To enable automatic updates, navigate to **OpenAdmin > General Settings** and check both the 'Auto Updates' and 'Auto Patches' options:

![openadmin update preferences](/img/admin/openadmin_set_update_preferences.png)

  </TabItem>
  <TabItem value="CLI" label="With OpenCLI">

To enable automatic updates from the terminal use commands:

```bash
opencli config update autoupdate yes
```

```bash
opencli config update autopatch yes
```
  </TabItem>
</Tabs>


### Disable Automatic Updates

<Tabs>
  <TabItem value="openadmin-admin-updates" label="With OpenAdmin" default>

To disable automatic updates, navigate to **OpenAdmin > General Settings** and uncheck the 'Auto Updates' and 'Auto Patches' options:

![openadmin update preferences](/img/admin/openadmin_set_update_preferences.png)

  </TabItem>
  <TabItem value="CLI" label="With OpenCLI">

To disable automatic updates from the terminal use commands:

```bash
opencli config update autoupdate no
```

```bash
opencli config update autopatch no
```
  </TabItem>
</Tabs>




### Manual Updates

When a new update is available, you will receive a notification in the admin panel.

<Tabs>
  <TabItem value="openadmin-admin-update-now" label="With OpenAdmin" default>

To manually update OpenPanel, navigate to **OpenAdmin > General Settings** and click on the "Update Now" button. NOTE: update is visible only if newer version is available.

Step 1.             |  Step 2.
:-------------------------:|:-------------------------:
![openadmin update manually ](/img/admin/admin_jupdate_available.png) |  ![openadmin update manually 2 ](/img/admin/admin_jupdate_available2.png)


  </TabItem>
  <TabItem value="CLI-update-now" label="With OpenCLI">

To update OpenPanel manually from the terminal, run the following command:

```bash
opencli update --force
```
  </TabItem>
</Tabs>























## Security

OpenPanel has been built from the ground up with security in mind. Internet history is littered with painful security incidents, so we traded old software compatibility and insecure authentication methods for modern day security measures.

**OpenPanel Security features:**
- Each user container is isolated by Docker.
- Two-Factor Authentication (2FA) can be activated by users.
- phpMyAdmin and WebTerminal offer auto-login using one-time tokens.
- Users' public services (SSH, MySQL) are accessible via non-standard ports.
- All user actions on the panel are stored in activity log.
- Bruteforce protection and rate limiting are implemented for all panel pages.
- The user panel is segregated from the admin panel and websites.
- All user panel requests are processed in the backend.

**OpenAdmin Security features:**
- The admin panel can be entirely disabled while retaining all functionality.
- HTTP Basic Authentication can be enabled for the admin panel.
- Admins can change the default port (2083) for the user panel.
- Email alerts and notifications for admin logins from new IP address.
- Bruteforce protection is enforced for the admin panel.
- Passwords are stored as salted SHA512 hashes by default (5000 rounds).
- The admin panel is isolated from the user panel and websites.
- Separated database software for admin and user accounts.

**Websites:**
- ModSecurity Web Application Firewall (WAF) can be activated for domains, with the OWASP core ruleset.
- IP blocking per domain name.
- Hotlink protection per domain using vhost files.
- No outgoing emails, only SMTP!
- TLS (Transport Layer Security) is utilized.

