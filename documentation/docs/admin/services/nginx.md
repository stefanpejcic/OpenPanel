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

- *Start* - `service nginx start`
- *Stop* - `service nginx stop`
- *Restart* - `service nginx restart`
- *Reload* - `service nginx reload`
- *Validate* - `nginx -t`


## How to edit Nginx Configuration

This section allows you to view and edit configuration for the Nginx webserver. 

Select the configuration file you would like to view or edit.

After making changes click on the Save button.

OpenPanel automatically backs up the configuration file. It runs nginx -t to validate the changes; if successful, the webserver reloads, applying the new configuration instantly.

In case of an error in the configuration, OpenPanel displays the exact issue and reverts to the previous backup file.

