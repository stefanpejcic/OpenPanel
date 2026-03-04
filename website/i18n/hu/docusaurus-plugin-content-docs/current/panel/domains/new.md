---
sidebar_position: 2
---

# Új domain hozzáadása

Új domain hozzáadásához egyszerűen írja be a domain nevet, és kattintson a **Domain hozzáadása** lehetőségre:

![domain hozzáadása](/img/panel/v2/openpanel_add_domain.gif)

Opcionálisan megadhat egy egyéni **Dokumentumgyökér** - mappát, ahol a webhely fájljait tárolja.

A beküldés után az OpenPanel részletes naplót jelenít meg az összes végrehajtott műveletről. Példa:

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

Mivel az OpenPanel igény szerint indítja el a szolgáltatásokat az erőforrások kímélése érdekében, az első domain hozzáadása némileg tovább tarthat – elindítja a szükséges webszervert és PHP-tárolókat.
A későbbi domainek sokkal gyorsabban kerülnek hozzáadásra.
