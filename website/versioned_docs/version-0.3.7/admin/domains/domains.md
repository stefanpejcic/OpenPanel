---
sidebar_position: 1
---

# Domains

Domains page displays all domains currently hosted on server.


## List domains

<Tabs>
  <TabItem value="openadmin-domains-list" label="With OpenAdmin" default>


To list all current domains navigate to Domains page:


| Field              | Description                                                               |
| ------------------ | ------------------------------------------------------------------------- |
| **ID**             | ID of the domain in database.                                                    |
| **Domain Name**      | The domain name.            |
| **Owner**    | User that added the domain.                                           |
| **DNS Zone**          | View and edit DNS zone for domain.        |
| **Virtual Hosts**     | View and edit Nginx configuration for a domain.          |
| **Access Logs**     | View live access log for the domain.           |
| **Visitors Report**  | Total number of domain names allowed per user on the plan.                  |


  </TabItem>
  <TabItem value="CLI-domains-list" label="With OpenCLI">

To list all current domains  run:

```bash
opencli domains-all
```

Example output:
```bash
opencli domains-all
stefan.openpanel.org
pejcic.rs
nesto.com
```

  </TabItem>
</Tabs>

## Add domain


<Tabs>
  <TabItem value="openadmin-domain-new" label="With OpenAdmin" default>

Domains can only be added [from the user interface](/docs/panel/domains/#adding-a-domain).

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


## Move domain

This is currently not possible.

## Delete domain

Domains can currently be deleted only [from the user interface](/docs/panel/domains/#delete-a-domain).
