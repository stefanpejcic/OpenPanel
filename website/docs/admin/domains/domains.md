---
sidebar_position: 1
---

# Manage Domains

Domains page displays all domains currently hosted on server.


## List domains

<Tabs>
  <TabItem value="openadmin-domains-list" label="With OpenAdmin" default>


To list all current domains navigate to Domains page. 

![openadmin domains page](/img/admin/2.0/openadmin_domains_table.png)

The table shows these columns by default:

| Field           | Description                                                       |
|-----------------|-------------------------------------------------------------------|
| **Domain**      | The domain name.                                                  |
| **Status**      | Indicates whether the domain is active or suspended.              |
| **PHP Version** | The PHP version configured for the domain.                        |
| **SSL**         | Shows whether SSL is Automatic, Custom, or None for the domain.   |
| **WAF**         | Toggle to enable/disable Coraza WAF for the domain.                |
| **Owner**       | The user who added or owns the domain.                            |

Click **Show Columns** to also display **ID**, **Docroot**, and **HTTP Strict Transport Security (HSTS)** — HSTS can be toggled on/off directly from that column, the same way WAF can. Column visibility is remembered for your browser.

Use the search box (**Search by user/domain...**) to filter the list by domain name or owning username.

  </TabItem>
  <TabItem value="CLI-domains-list" label="With OpenCLI">

To list all current domains  run:

```bash
opencli domains-all
```

Example output:
```bash
root@server:~# opencli domains-all
openpanel.com
openpanel.org
community.openpanel.org
api.openpanel.com
support.openpanel.org
ip.openpanel.com
my.openpanel.com
```

  </TabItem>
</Tabs>

## Add domain


<Tabs>
  <TabItem value="openadmin-domain-new" label="With OpenAdmin" default>

  Click on 'Add Domain' button, insert the domain and select the user to add it, then click on 'Add Domain'.

![openadmin domains add](/img/admin/2.0/openadmin_domains_add.png)


  </TabItem>
  <TabItem value="CLI-domain-new" label="With OpenCLI">
    
To create a new plan run the following command:

```bash
opencli domains-add <DOMAIN_NAME> <USERNAME> [--debug]
```

Example:
```bash
root@stefan:/usr/local/admin# opencli domains-add pejcci.rs wzs11p2i --debug
Checking if domain already exists on the server
Adding pejcci.rs to the domains database
Purging cached list of domains for the account
Creating document root directory /home/wzs11p2i/pejcci.rs
Checking webserver configuration
Checking if default vhosts file exists for Nginx
Checking IPv4 address for the account
Creating /etc/nginx/sites-available/pejcci.rs.conf
Restarting nginx to apply changes
Creating vhosts proxy file for Nginx
Webserver is running, reloading configuration
Creating DNS zone file: /etc/bind/zones/pejcci.rs.zone
DNS service is running, adding the zone
Adding the newly created zone file to the DNS server
Checking and setting nginx service to automatically start on reboot
Starting service for the default PHP version 8.2
Checking and setting PHP service to automatically start on reboot
Checking and starting the ssl generation service
Starting Let'sEncrypt SSL generation in background
Domain pejcci.rs added successfully
```
  </TabItem>
</Tabs>

## Domain actions

Each row in the Domains table has an actions menu (the `⋮` button) with the following options:

![openadmin domains actions](/img/admin/2.0/openadmin_domains_actions.png)

- **Edit DNS Zone** — opens the [DNS Zone Editor](/docs/admin/domains/dns) for this domain. Only shown if the **dns** module is enabled.
- **Suspend domain** / **Unsuspend domain** — toggles the domain's status.
- **Manage SSL** — opens the SSL page for the domain. Not shown for suspended domains.
- **Edit VHosts** — opens the VirtualHost config editor for the domain.
- **Edit Caddyfile** — opens the Caddy config editor for the domain.
- **Delete domain** — permanently deletes the domain, see [Delete domain](#delete-domain) below.

## Move domain

This is currently not possible.

## Delete domain

Domains can be deleted directly from OpenAdmin:

1. Open the domain's actions menu (`⋮`) in the Domains table and click **Delete domain**.
2. A confirmation dialog lists what will be removed: VHost configuration files, Caddyfile entries, SSL certificates, and email accounts & redirects. Website files and the docroot directory are **not** deleted.
3. Type the domain name to confirm, then click **Delete Permanently**.

> Domains with attached websites cannot be deleted — remove all websites from Site Manager first. If the domain is added again later, default configurations are recreated.
