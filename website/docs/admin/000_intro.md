---
sidebar_position: 1
---

# Get Started with OpenAdmin

The OpenAdmin offers an administrator-level interface where you can efficiently handle tasks such as creating and managing users, setting up hosting plans, configuring backups, and editing OpenPanel settings.

![openpanel vs openadmin](/img/admin/openadmin_vs_openpanel_what_is_the_difference.png)



## Requirements

Minimum Requirements:

- A blank full virtual machine or bare metal server
- Minimum of 1GB RAM and 5GB storage (4GB RAM and 50GB is recommended)
- x86_64/amd64 architecture **[support for ARM (AArch64) is in progress](https://github.com/stefanpejcic/OpenPanel/issues/63)*
- IPv4 address

Supported OS:
- **Ubuntu 24.04**

:::info
If you are using external firewall, the following ports should be opened:  `25` `53` `80` `443` `465` `993` `2083` `2087` `32768:60999`
:::

## Installation

OpenPanel can be installed on both VPS and bare-metal servers. 

The installation process takes about 5 minutes. To install openpanel follow these steps: 

<Tabs>
  <TabItem value="openpanel-install-on-dedicated" label="Install script" default>

1. Log in to your new server;
- as root via SSH or
- as a user with sudo privileges and type "sudo -i"
2. Copy and paste openpanel installation command into the terminal
```shell
bash <(curl -sSL https://openpanel.org)
```

The installation script supports [optional flags](/install) that can be used to configure openpanel, skip certain installation steps or simply display debugging information.

If you encountered any errors while running the installation script, please copy & paste the installation log file to [the community forums](https://community.openpanel.org).

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

Recommended steps after installing OpenPanel:
- [access admin panel](/docs/admin/intro#access-openadmin)
- [set domain for accessing panels](/docs/admin/settings/general/#set-domain-for-openpanel)
- [set custom nameservers](/docs/admin/settings/openpanel/#set-nameservers)
- [create a hosting plan](/docs/admin/plans/hosting_plans#create-a-plan)
- [create a new user account](/docs/admin/users/openpanel#create-users)
- [set admin email for server alerts](/docs/admin/notifications/#email-alerts)

### Access OpenAdmin

Run `opencli admin` command to find the address on which admin panel is accessible. Example output:

```bash
root@server:/home# opencli admin
● OpenAdmin is running and is available on: https://server.openpanel.org:2087/
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
● OpenAdmin is running and is available on: https://server.openpanel.co:2087/

- username: stefan
- password: ba63vfav7fq36vas

===============================================================
```

### Enable Automatic Updates

- `autopatch` option allows Administrator to automatically update OpenPanel to minor versions. MINOR versions include only security updates and bug fixes.
- `autoupdate` option allows Administrator to enable or disable automatic updates to major versions. MAJOR versions add new functionality in a backward compatible manner.

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

















## License

OpenPanel is available in two editions:

- [**OpenPanel Community edition**](/product/openpanel-free-control-panel/) is a free version of the panel that is limited to 3 user accounts and 50 domains, which should be more than enough for personal use.
- [**OpenPanel Enterprise edition**](/product/openpanel-premium-control-panel/) unlocks premium features for user isolation and management, suitable for web hosting providers. It has API access and can easily be integrated with 3rd party billing tools like WHMCS and FOSSBilling.


[OpenPanel Community VS Enterprise](/beta/)



## Security

OpenPanel has been built from the ground up with security in mind. Internet history is littered with painful security incidents, so we traded old software compatibility and insecure authentication methods for modern day security measures.


### Firewall
OpenPanel supports both [ConfigServer & Firewall (CSF)](/docs/admin/security/firewall/#csf) and [UncomplicatedFirewall (UFW)](/docs/admin/security/firewall/#ufw).

### Rootless Docker
Each user runs docker in rootless mode, ensuring containers can not commmunicate using docker networks and providing more issolation.

### Isolated Services
Each user is provided with a containerized environment similar to a VPS, featuring their own web server (Nginx or Apache) and database (MySQL or MariaDB). This setup prevents resource hogging commonly associated with standard shared hosting.

### CorazaWAF

OpenPanel uses CorazaWAF compatible with OWASP CoreRuleset. Each user can enable/disable WAF fpr their domains.

### Two-Factor Authentication
Users have the option to [enable Two-Factor Authentication (2FA)](/docs/panel/account/2fa/) for added security on their accounts. Administrators can manage this feature at the server level or for individual users.

### Detailed Logging
All actions taken by OpenPanel users are recorded in per-user activity logs. This eliminates confusion over issues like file or webmail account deletions—every action is logged and can be reviewed by users.

### Isolated UIs
OpenPanel and OpenAdmin operate independently from one another. One runs as a systemd service while the other runs as a Docker container. OpenPanel utilizes SQLite for its database, whereas OpenAdmin relies on MySQL. Importantly, users can perform actions on their panel even if the admin panel is unreachable or disabled.


### Disabling the Admin Panel
For production environments, particularly with the Community edition—which does not offer API access and lacks third-party integrations—it is advisable to disable the admin panel after configuring your server. Alternatively, you can restrict access to the admin port `2087` by whitelisting your team's IP addresses.

To disable OpenAdmin, navigate to **OpenAdmin > Settings > OpenAdmin** and click on *"Disable Admin Panel"* or use the terminal command `opencli admin off`. This will deactivate the admin panel, and you can re-enable it when necessary with the command `opencli admin on`.

### Limit Access to Admin Panel
To restrict OpenAdmin access to your team, whitelist your server's IP addresses in CSF/UFW, and then disable port `2087`.

### HTTP Basic Authentication
As an additional security measure, HTTP Basic Authentication can be enabled for the admin panel.

### Brute-Force Protection

Both user and admin interfaces have a built-in rate limiting and IP address blocking to protect against brute-force attacks. You can configure the maximum number of failed login attempts allowed per IP (default is `5` per minute) and the total number of failed attempts (default is `20`), after which the offending IP will be temporarily blocked by the firewall for one hour.

For user panel imits are configurable in: /etc/openpanel/openpanel/conf/openpanel.config file:
```bash
[USERS]
login_ratelimit=5
login_blocklimit=20
```

![user ratelimit](/img/panel/v1/user_block.png)

For admin panel imits are configurable in: /etc/openpanel/openadmin/config/admin.ini file:
```bash
[PANEL]
login_ratelimit=5
login_blocklimit=20
```

![admin ratelimit](/img/admin/admin_block.png)

If a user successfully logs in, the counter for `login_blocklimit` will reset.
Failed login attempts and blocked IP addresses are logged in the `/var/log/openpanel/admin/failed_login.log` file for OpenAdmin and in the `/var/log/openpanel/user/failed_login.log` file for OpenPanel.

### IP blocking per domain

Users can block IP addresses per domain name.
