---
sidebar_position: 1
---

# General Settings

Change ports, domain, enable ssl, etc.

![openadmin general panel settings](/img/admin/adminpanel_general_settings.png)

The General Settings page enables you to edit settings that affect both admin and user panel:

- domain to access panels
- ports to be used
- update preferences

## Set domain for OpenPanel

To enable access to both OpenAdmin and OpenPanel through a domain name, such as srv.your-domain.com:2083, follow these three steps:

1. Set hostname:
   Set the desired subdomain as server hostname:
   ```
   hostnamectl set-hostname server.example.net
   ```
2. Configure DNS:
   Point the subdomain to the public IP of the server.
   Use a tool such as https://www.whatsmydns.net/ to check that domain is pointed to server ip.
   
3. Set in General Settings
   Switch from IP to the domain name in *OpenAdmin > General Settings*.

![openadmin set_domainname](/img/admin/adminpanel_domainname.png)

## Set IP address for OpenPanel

To access OpenPanel and OpenAdmin via server public IP address, choose the "Server IP address" option and click save. The modification is immediate, redirecting you to the designated IP:2087 for the admin panel upon saving.

![openadmin set ip](/img/admin/adminpanel_serverip.png)

## Change OpenPanel port

Port configurations for OpenAdmin and OpenPanel interfaces can be modified from their default settings (2087 for OpenAdmin and 2083 for OpenPanel). 

![openadmin set port for openpanel](/img/admin/openpanel_settings.png)

To change the port for the OpenPanel from the default `2083` to another value, you can easily set the desired port in the "OpenPanel Port" field.
It's important to note that the port must fall within the range of 1000-33000.

## Change OpenAdmin port

:::info
NOTE: Changing admin port from 2087 is currently not possible, as external tools such as WHMCS do not have option to set custom port.
:::

![openadmin set port for openadmin](/img/admin/openadmin_settings.png)

# Change /openpanel

By default, when users add a domain, the addition of "/openpanel" to the domain URL will redirect them to the OpenPanel interface. However, you have the flexibility to customize this, such as changing it to "/awesome," allowing users to access their OpenPanel via "their-domain.com/awesome".

To change the "/openpanel" to something else, simply set the value for the "OpenPanel is also available on:" field and click on save. Changes take effect instantly without service interruption.
![openadmin set custom path](/img/admin/openpanel_settings_available_on_openpanel.png)
