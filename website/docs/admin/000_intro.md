---
sidebar_position: 1
---

# Get Started

The OpenAdmin offers an administrator-level interface where you can efficiently handle tasks such as creating and managing users, setting up hosting plans, enabling featrues, and editing OpenPanel settings.

## Requirements

Minimum Requirements:

- A blank full virtual machine or bare metal server
- Minimum of 1GB RAM and 5GB storage (4GB RAM and 50GB is recommended)
- AMD64(x86_64) or ARM(AArch64) architecture
- IPv4 address

Supported OS:
- **Ubuntu 24.04** (recommended)
- **Debian [10](https://voidnull.es/instalacion-de-openpanel-en-debian-10/), 11 and 12**
- **AlmaLinux 9.5** (recommended for ARM cpu)
- **RockyLinux 9**
- **CentOS 9.5**

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
  <TabItem value="openpanel-install-on-absible" label="Ansible">

```bash
---
- name: Install OpenPanel on target machine
  hosts: all
  become: true
  vars:
    openpanel_install_flags: "--debug --username=admin --password=super123"  # Customize your flags here, full list: https://openpanel.com/install

  tasks:
    - name: Download and run OpenPanel installer 
      shell: |
        curl -sSL https://openpanel.org | bash -s -- {{ openpanel_install_flags }}
      args:
        executable: /bin/bash
```

  </TabItem>
</Tabs>


## Post Install Steps

Recommended steps after installing OpenPanel:
- [access admin panel](/docs/admin/intro#access-openadmin)
- [set domain for accessing panels](/docs/admin/settings/general/#set-domain-for-openpanel)
- [set custom nameservers](/docs/admin/settings/openpanel/#set-nameservers)
- [create a hosting plan](/docs/admin/plans/hosting_plans#create-a-plan)
- [create a new user account](/docs/admin/accounts/users/#create-users)
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

To enable automatic updates, navigate to **OpenAdmin > Settings > Update Preferences** and change 'Update automatically' option to 'Both':

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

To disable automatic updates, navigate to **OpenAdmin > Settings > Update Preferences** and change 'Update automatically' option to 'Never':

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

To manually update OpenPanel, navigate to **OpenAdmin > Settings > Update Preferences** and click on the "Update Now" button. NOTE: update is visible only if newer version is available.

  </TabItem>
  <TabItem value="CLI-update-now" label="With OpenCLI">

To update OpenPanel manually from the terminal, run the following command:

```bash
opencli update --force
```
  </TabItem>
  
</Tabs>
