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
  <TabItem value="openpanel-install-on-ansible" label="Ansible">

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
- [Access the Admin panel](/docs/articles/dev-experience/how-to-access-openadmin)
- [Configure Domain and SSL for OpenPanel](/docs/admin/settings/general/#set-domain-for-openpanel)
- [Configure Custom Nameservers](/docs/admin/settings/openpanel/#set-nameservers)
- [Create Hosting Packages](/docs/admin/plans/hosting_plans#create-a-plan)
- [Create New User Accounts](/docs/admin/accounts/users/#create-users)
- [Set Email address to receive Alerts](/docs/admin/notifications/#email-alerts)
- [Change Update Preferences](/docs/admin/settings/updates)
- [Secure OpenPanel for Production Use](/docs/articles/security/securing-openpanel/)
