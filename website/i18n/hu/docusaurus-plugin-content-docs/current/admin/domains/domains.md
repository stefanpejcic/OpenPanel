---
sidebar_position: 1
---

# Domainek kezelése

A Domains oldalon a szerveren jelenleg tárolt összes domain látható.


## Domainek listázása

<Tabs>
<TabItem value="openadmin-domains-list" label="OpenAdminnal" alapértelmezett>


Az összes jelenlegi domain listázásához lépjen a Domains oldalra:


| Mező | Leírás |
|-----------------|-------------------------------------------------------------------|
| **Domain** | A domain név.                                                  |
| **Állapot** | Azt jelzi, hogy a tartomány aktív vagy felfüggesztett.              |
| **PHP verzió** | A tartományhoz konfigurált PHP-verzió.                        |
| **SSL** | Megmutatja, hogy az SSL engedélyezve van-e a tartományban.                      |
| **WAF** | Azt jelzi, hogy a Coraza WAF engedélyezve van-e vagy letiltva a tartományban.    |
| **Tulajdonos** | A domaint hozzáadó vagy tulajdonos felhasználó.                            |
| **Analytics** | Tekintse meg a domain elemzési adatait és jelentéseit.                   |


</TabItem>
<TabItem value="CLI-domains-list" label="OpenCLI-vel">

Az összes jelenleg futó domain listája:

```bash
opencli domains-all
```

Példa kimenet:
```bash
opencli domains-all
stefan.openpanel.org
pejcic.rs
nesto.com
```

</TabItem>
</Tabs>

## Domain hozzáadása


<Tabs>
<TabItem value="openadmin-domain-new" label="OpenAdminnal" alapértelmezett>

Kattintson a "Domain hozzáadása" gombra, illessze be a domaint, válassza ki a hozzáadni kívánt felhasználót, majd kattintson a "Domain hozzáadása" gombra.

</TabItem>
<TabItem value="CLI-domain-new" label="OpenCLI-vel">
    
Új terv létrehozásához futtassa a következő parancsot:

```bash
opencli domains-add <DOMAIN_NAME> <USERNAME> [--debug]
```

Példa:
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


## Domain áthelyezése

Ez jelenleg nem lehetséges.

## Domain törlése

A domainek jelenleg csak [a felhasználói felületről] (/docs/panel/domains/#delete-a-domain) törölhetők.
