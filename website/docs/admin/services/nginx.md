---
sidebar_position: 2
---

# Nginx Configuration

The *OpenAdmin > Services > Nginx Configuration* feature enables Administrators to view current Nginx status and edit configuration.

![Nginx Configuration OpenAdmin](/img/admin/openadmin_services_nginx.png)

## How to View Nginx Configuration

Navigate to *Services > Nginx Configuration*

Displayed information:
- *Status* - Current status of the Nginx service
- *Version* - Current version
- *Active Connections* - The current number of active client connections including Waiting connections.
- *Virtual Hosts* - The current number of virtualhosts (domain) files.
- *Modules* - Lists installed Nginx modules and compiled flags
- *Current Reading Connections* - The current number of connections where nginx is reading the request header.
- *Current Writing Connections* - The current number of connections where nginx is writing the response back to the client.
- *Current Waiting Connections* - The current number of idle client connections waiting for a request.
- *Total Accepted Connections* - The total number of accepted client connections since the Nginx service is active.
- *Total Handled Connections* - The total number of handled connections. Generally, the parameter value is the same as accepts unless some resource limits have been reached (for example, the worker_connections limit).
- *Total Requests* - The total number of client requests since the Nginx service is active.

## How to stop/start Nginx

Navigate to *Services > Nginx Configuration*

Under the 'Actions' section click on the button:

- *Start* - `docker exec nginx nginx start`
- *Stop* - `docker exec nginx nginx stop`
- *Restart* - `docker exec nginx nginx restart`
- *Reload* - `docker exec nginx nginx reload`
- *Validate* - `nginx -t`


## How to edit Nginx Configuration

This section allows you to view and edit configuration for the Nginx webserver. 

Select the configuration file you would like to view or edit.

After making changes click on the Save button.

OpenPanel automatically backs up the configuration file. It runs nginx -t to validate the changes; if successful, the webserver reloads, applying the new configuration instantly.

In case of an error in the configuration, OpenPanel displays the exact issue and reverts to the previous backup file.


### nginx.conf
nginx.conf is the main configuration file located in `/etc/nginx/nginx.conf` inside nginx container.

```bash
/etc/openpanel/nginx/nginx.conf
```

![nginx main conf](/img/admin/nginx/nginx_mainconf.png)



### Default VHost Template
The default nginx configuration file (in nginx container: `/etc/nginx/sites-available/default`). This file is used for domains that are pointed to the server IP but not currently added by any OpenPanel user.

```bash
/etc/openpanel/nginx/vhosts/default.conf
```
![nginx default conf](/img/admin/nginx/nginx_defaultconf.png)

### Default Landing Page
Landing page is displayed on domains that have no website yet.
```bash
/etc/openpanel/nginx/default_page.html
```
![nginx landing](/img/admin/nginx/nginx_landing.png)


### Suspended User Template
Suspended User Template is displayed on every domain for a suspended user.
```bash
/etc/openpanel/nginx/suspended_user.html
```
![nginx suspended user](/img/admin/nginx/nginx_suspendeduser.png)


### Suspended Domain Template
Suspended Domain Template is displayed on suspended domains.
```bash
/etc/openpanel/nginx/suspended_website.html
```
![nginx suspended website](/img/admin/nginx/nginx_suspendedwebsite.png)

### Domain VHost Template
This file is copied for every new domain added by OpenPanel users.
```bash
/etc/openpanel/nginx/vhosts/domain.conf
```
![nginx domain](/img/admin/nginx/nginx_domain.png)

### Domain ModSecurity VHost Template
If ModSecurity is installed on the server, this file is copied for every new domain added by OpenPanel users.
```bash
/etc/openpanel/nginx/vhosts/domain.conf_with_modsec
```
![nginx domain modsec](/img/admin/nginx/nginx_domainmodsec.png)

### Domain ModSecurity VHost Template
If ModSecurity is installed on the server, this file is copied for every new domain added by OpenPanel users.
```bash
/etc/openpanel/nginx/vhosts/domain.conf_with_modsec
```
![nginx domain modsec](/img/admin/nginx/nginx_domainmodsec.png)

## Domain Nginx Docker VHost Template
For users with Nginx webserver, this file is copied inside their docker container for every new domain.
```bash
/etc/openpanel/nginx/vhosts/docker_nginx_domain.conf
```
![docker_nginx_domain](/img/admin/nginx/docker_nginx_domain.png)

## Domain Apache Docker VHost Template
For users with Apache webserver, this file is copied inside their docker container for every new domain.
```bash
/etc/openpanel/nginx/vhosts/docker_apache_domain.conf
```
![docker_apache_domain](/img/admin/nginx/docker_apache_domain.png)

## OpenPanel Proxy
This file affects all user domains and makes accessing services via user domains: `/openpanel` `/openadmin` `/webmail`
```bash
/etc/openpanel/nginx/vhosts/openpanel_proxy.conf
```
![openpanel_proxy](/img/admin/nginx/openpanel_proxy.png)


## OpenPanel Proxy
These error pages are shown to visitors when the domain configuration file is malformed or no response from docker container.
```bash
/etc/openpanel/nginx/vhosts/error_pages/
```
![error_pages](/img/admin/nginx/error_pages.png.png)




