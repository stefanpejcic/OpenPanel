---
sidebar_position: 1
---

# General Settings

From this page Administrators can configure the domain to access both OpenPanel and OpenAdmin interfaces, as well as ports for those services.

![openadmin general panel settings](/img/admin/adminpanel_general_settings.png)

## Domain

Both OpenAdmin and OpenPanel are accessed through the same single **Hostname** field: enter either a domain name or the server's IP address there, and both interfaces become reachable at `https://HOSTNAME:PORT/`.

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
   Set the domain name in the **Hostname** field on *OpenAdmin > Settings > General Settings*.

:::info
You can also [set a separate domain just for OpenPanel UI](/docs/articles/dev-experience/separate-domain-for-openpanel-access)
:::

## Set IP address for OpenPanel

To access OpenPanel and OpenAdmin via the server's public IP address, enter the IP address in the same **Hostname** field in *OpenAdmin > Settings > General Settings* (this field accepts either a domain name or an IP address).

## Ports

Port configurations for OpenAdmin and OpenPanel interfaces can be modified from their default settings (`2087` for OpenAdmin and `2083` for OpenPanel). 

- To change the port for the OpenPanel from the default `2083` to another value, set the desired port in the "OpenPanel port" field.
- To change the port for the OpenAdmin from the default `2087` to another value, set the desired port in the "OpenAdmin port" field.

## Redirect

By default, when users add a domain, the addition of "/openpanel" to the domain URL will redirect them to the OpenPanel interface. However, you have the flexibility to customize this, such as changing it to "/awesome," allowing users to access their OpenPanel via "their-domain.com/awesome".

To change the "/openpanel" to something else, simply set the value in the "/openpanel" field and click on save. Changes take effect instantly without service interruption.

## Debugging (Dev Mode)

Toggle **Dev mode** to enable verbose/detailed logging for OpenPanel and OpenAdmin, useful for error diagnostics. Logs can be viewed under *OpenAdmin > Services > View Log Files*.

:::info
After saving, an OpenAdmin restart is needed for the new value to take effect.
:::

