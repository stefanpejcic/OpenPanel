---
sidebar_position: 1
---

# Get Started

The OpenAdmin offers an administrator-level interface where you can efficiently handle tasks such as creating and managing users, setting up hosting plans, enabling features, and editing OpenPanel settings.

## Requirements

Minimum Requirements:

- A blank full virtual machine or bare metal server
- Minimum of 1GB RAM and 5GB storage (4GB RAM and 50GB is recommended)
- AMD64(x86_64) or ARM(AArch64) architecture
- IPv4 address

Supported operating systems:
- **Ubuntu 24.04** (recommended)
- **Debian 10, 11, 12, 13**
- **AlmaLinux 9.5 and 10** (9.5 is recommended for ARM CPU, 10 has a known issue [#744](https://github.com/stefanpejcic/OpenPanel/issues/744))
- **RockyLinux 9.6, 10**
- **CentOS 9.5**

On AlmaLinux 10 and RockyLinux 10, you must switch from `nftables` to `iptables`. See [#1472](https://github.com/docker/for-linux/issues/1472) and [#745](https://github.com/stefanpejcic/OpenPanel/issues/745#issuecomment-3451272947).

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
  <TabItem value="openpanel-install-on-cloud" label="Cloud">

[Amazon Web Services (AWS)](/docs/articles/install-update/install-on-aws)

[DigitalOcean](/docs/articles/install-update/install-on-digitalocean)

[Google Cloud Platform (GCP)](/docs/articles/install-update/install-on-google-cloud)

[Microsoft Azure](/docs/articles/install-update/install-on-microsoft-azure)

[Vultr](/docs/articles/install-update/install-on-vultr)
    
  </TabItem>
  <TabItem value="openpanel-install-on-other" label="Other">

[CloudInit](/docs/articles/install-update/install-using-cloudinit)

[Ansible](/docs/articles/install-update/install-using-ansible)

[Virtualizor](/docs/articles/install-update/install-on-virtualizor)

  </TabItem>  
</Tabs>


## Post Install Steps

Recommended steps after installing OpenPanel:
- [Access the OpenAdmin panel](/docs/articles/dev-experience/how-to-access-openadmin)
- [Configure Domain and SSL for OpenPanel](/docs/admin/settings/general/#set-domain-for-openpanel)
- [Enable Modules (features) in OpenPanel UI](/docs/admin/settings/modules/)
- [Configure Custom Nameservers](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/)
- [Create Hosting Packages](/docs/admin/plans/hosting_plans#create-a-plan)
- [Create New User Accounts](/docs/admin/accounts/users/#create-users)
- [Set Email address to receive Alerts](/docs/admin/notifications/#email-alerts)
- [Change Update Preferences](/docs/admin/settings/updates)
- [Secure OpenPanel for Production Use](/docs/articles/security/securing-openpanel/)
