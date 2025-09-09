---
sidebar_position: 2
---

# Add new domain

To add a new domain, simply enter the domain name and click **Add Domain**:

![add domain](/img/panel/v2/openpanel_add_domain.gif)

Optionally, you can specify a custom **Document Root** - folder where website files will be stored.

Once submitted, OpenPanel will display a detailed log of all actions performed. Example:

```bash
Using document root: /var/www/html/another-website
Checking if domain already exists on the server
Checking domain against system domains list
Adding another.com to the domains database
Purging cached list of domains for the account
Creating document root directory /var/www/html/another-website
Checking webserver configuration
Checking IPv4 address for the account
Starting apache container..
Creating another.com.conf
Creating vhosts proxy file for Caddy with ModSecurity OWASP Coreruleset
Caddy is running, validating new domain configuration
Domain successfully added and Caddy reloaded.
Creating DNS zone file with A records: /etc/bind/zones/another.com.zone
DNS service is running, adding the zone
Adding the newly created zone file to the DNS server
Starting container for the default PHP version 8.4
Domain another.com added successfully
```

Because OpenPanel starts services on demand to conserve resources, adding your first domain may take slightly longer - it will start the necessary webserver and PHP containers.
Subsequent domains will be added much faster.
