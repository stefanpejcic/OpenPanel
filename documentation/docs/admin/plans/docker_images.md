---
sidebar_position: 2
---

# Docker images

Docker images are the base for every hosting plan and they define which services are pre-installed for a user.

We offer and maintain two default Docker images, each preconfigured with a LAMP stack:

[openpanel/nginx](https://dev.openpanel.co/images/browse.html#openpanel-apache)
[openpanel/apache](https://dev.openpanel.co/images/browse.html#openpanel-nginx)

For adding custom services or files, such as preinstalling PostgreSQL, [create a custom Docker image](https://dev.openpanel.co/images/create.html) and assign it to the plan.

## openpanel/apache

Our official Docker image for Apache enables OpenPanel users to utilize the Apache web server for website management. Apache natively supports .htaccess files, although restarting Apache is required to activate any modifications made through .htaccess.

- Download: https://hub.docker.com/r/openpanel/apache

Technology Stack:
| Service/Tool |                 Purpose                |
|:------------:|:--------------------------------------:|
| Ubuntu 24    | Operating System                       |
| Apache2      | web server                             |
| MySQL        | database                               |
| PHP 8.2      | serving .php files                     |
| phpMyAdmin   | manage databases from GUI              |
| OpenSSH      | SSH access                             |
| NodeJS 18    | nodejs                                 |
| Python 3.9   | python                                 |
| curl / wget  | downloading files                      |
| pwgen        | password generate                      |
| zip / unzip  | create/extract archives                |
| ttyd         | Web Terminal                           |
| screen       | background processes                   |
| cron         | run scheduled tasks                    |


When this image is set for a OpenPanel plan, all new users created on that plan will use this image as a base.

## openpanel/nginx

The official Docker image for Nginx provides OpenPanel users with the capability to manage websites using the Nginx web server. Unlike Apache, Nginx does not natively support .htaccess files. For configuration changes, directly editing the Nginx configuration files is necessary, and requires reloading Nginx.

- Download: https://hub.docker.com/r/openpanel/nginx


Technology Stack: 
| Service/Tool |                 Purpose                |
|:------------:|:--------------------------------------:|
| Ubuntu 24    | Operating System                       |
| Nginx        | web server                             |
| MySQL        | database                               |
| PHP 8.2      | serving .php files                     |
| phpMyAdmin   | manage databases from GUI              |
| OpenSSH      | SSH access                             |
| NodeJS 18    | nodejs                                 |
| Python 3.9   | python                                 |
| curl / wget  | downloading files                      |
| pwgen        | password generate                      |
| zip / unzip  | create/extract archives                |
| ttyd         | Web Terminal                           |
| screen       | background processes                   |
| cron         | run scheduled tasks                    |

When this image is set for a OpenPanel plan, all new users created on that plan will use this image as a base.

## Custom Docker image

https://dev.openpanel.co/images/create.html



