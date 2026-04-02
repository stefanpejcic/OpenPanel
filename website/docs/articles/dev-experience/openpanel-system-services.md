# Services

OpenPanel uses the following services:

## Caddy

[Caddy](https://caddyserver.com/) is utilized as a reverse proxy to manage traffic to and from user websites hosted in their Docker containers. It also handles SSL certificate generation and renewals for both user domains and the main control panel domain.

Configuration Files:
- `/etc/openpanel/caddy/Caddyfile`: This is the main configuration file that includes all others.
- `/etc/openpanel/caddy/check.conf`: This file manages domain verification for SSL generation on the server.
- `/etc/openpanel/caddy/redirects.conf`: Contains redirections for `/openpanel`, `/openadmin`, and `/webmail` for all domains.
- `/etc/openpanel/caddy/domains/`: This directory holds a `.conf` file for each domain added to the server.
- `/etc/openpanel/caddy/templates/`: Contains configuration templates for new and suspended domains.
- `/etc/openpanel/caddy/suspended_domains/`: Stores user domains once they have been suspended.

You can view the default Caddy configuration [here](https://github.com/stefanpejcic/openpanel-configuration/tree/main/caddy).

Admins can set memory (RAM) and CPU limits for Caddy by modifying the `/root/.env` file:

```
CADDY_RAM="1g"
CADDY_CPUS="1.0"
```

OpenPanel uses a custom Caddy image, which includes [OWASP Coraza WAF](https://coraza.io/), a high-performance open-source Web Application Firewall. The source for this custom image can be found [here](https://github.com/stefanpejcic/OpenPanel/blob/main/docker/caddy-coraza/Dockerfile).

If you prefer not to use Coraza WAF, you can switch to the official Caddy image by updating the `/root/.env` file:

To use the custom Caddy image with Coraza:

```
CADDY_IMAGE="openpanel/caddy-coraza"
```

To use the official Caddy image without Coraza:

```
CADDY_IMAGE="caddy/latest"
```

## Docker


[Docker](https://www.docker.com/) rootful is used to run all the shared services. Default services can be viewed [here](https://github.com/stefanpejcic/openpanel-configuration/blob/main/docker/compose/docker-compose.yml).

[Docker Rootless mode](https://docs.docker.com/engine/security/rootless/) is used for each user account to isolate their containers.

## MySQL

[MySQL](https://www.mysql.com/) is used as a database to store all user related information:

- Users
- Plans
- Domains
- Websites

MySQL login information is stored in files: `/etc/openpanel/mysql/host_db.conf` and `/etc/openpanel/mysql/hcontainer_db.conf`. 

- `/etc/openpanel/mysql/host_db.conf` is symlinked to `/etc/my.cnf` and is used for OpenAdmin service and OpenCLI commands.
- `/etc/openpanel/mysql/hcontainer_db.conf` is mounted to `/etc/my.cnf` in OpenPanel container and is used by it.

MySQL data is stored inside the `root_mysql` docker volume.

Admins can set memory (RAM) and CPU limits for MySQL service by modifying the `/root/.env` file:

```
MYSQL_RAM="0.5g"
MYSQL_CPUS="1.0"
```

`MYSQL_ROOT_PASSWORD` stores the random mysql root password:
```
MYSQL_ROOT_PASSWORD=
```

## SQLite

SQLite database is used by the OpenAdmin panel in order to completely separate the Admin and end-user interface.
Database file: `/etc/openpanel/openadmin/users.db` contians all administrator accounts.

## Named

[Named (BIND9)](https://www.isc.org/bind/) container is used for DNS.

Administrators can set custom nameservers on *OpenAdmin > OpenPanel Settings* to be used for OpenPanel websites.

[ default named.conf.options` configuration file](https://github.com/stefanpejcic/openpanel-configuration/blob/main/bind9/named.conf.options)

Admins can set memory (RAM) and CPU limits for BIND9 service by modifying the `/root/.env` file:

```
BIND9_RAM="0.1g"
BIND9_CPUS="0.1"
```

BIND9 version can be set by changing the `latest` tag to any version from [hub.docker.com](https://hub.docker.com/r/ubuntu/bind9/tags):
```
BIND9_VERSION="latest"
```

Timezone for the BIND9 service can be changed by editing:

```
BIND_TIMEZONE="America/New_York"
```

## GoAccess

OpenPanel uses [GoAccess](https://goaccess.io/) to generate beautiful HTML reports from Nginx access logs for each domain.
OpenPanel also downloads [GeoIPLite2 City and Country](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) databases in order to display location info for each IP address in reports.
To customize these reports edit the [/etc/openpanel/goaccess/goaccess.conf](https://github.com/stefanpejcic/openpanel-configuration/blob/main/goaccess/goaccess.conf) file.

## CSF


## OpenPanel

[openpanel/openpanel](https://hub.docker.com/r/openpanel/openpanel) docker image is used to provide access to the OpenPanel interface that allows users to manage their accounts.

Admins can set the OpenPanel version by changing the tag in the `/root/.env` file with any tag from [hub.docker.com](https://hub.docker.com/r/openpanel/openpanel/tags):
```
VERSION="1.0.0"
```

By default OpenPanel is accesible on port `2083` but this can also be configured by Admin by editing the `PORT` value:

```
PORT="2083"
```


OpenPanel is a [Flask](https://flask.palletsprojects.com/en/3.0.x/)-based application that uses [MySQL](#MySQL) to store user data and operates on the [Gunicorn](https://gunicorn.org/) web server. This configuration ensures that OpenPanel remains functional even if the web server is down, thereby providing complete isolation between user websites and the admin panel.

OpenPanel operates in production mode by default, logging only [errors and access logs](/logs.html). For developers needing more detailed logs for troubleshooting or development purposes, it is possible to switch to a more verbose logging mode by enabling dev_mode.




## OpenAdmin

`admin` service is used by the OpenAdmin interface.

By default, OpenAdmin is accessible on port `2087`, but this can be changed by the Admin by modifying the `port` value in the admin service configuration. However, due to limitations in billing systems like WHMCS, which do not support custom port settings, changing the port through the admin interface is not allowed. If the port is changed via the terminal, it may disrupt any third-party integrations that rely on the default port `2087`.

OpenAdmin is a [Flask](https://flask.palletsprojects.com/en/3.0.x/) application that utilizes [SQLite](https://www.sqlite.org/) for its database and runs on a separate Gunicorn instance. It operates independently from the OpenPanel and webserver.

## ClamAV

[ClamAV](https://www.clamav.net/) is an open-source antivirus engine for detecting trojans, viruses, malware & other malicious threats. OpenPanel uses ClamAV for the `'malware_scanner' feature.


Admins can set memory (RAM) and CPU limits for ClamAV service by modifying the `/root/.env` file:

```
CLAMAV_RAM="1g"
CLAMAV_CPUS="1.0"
```
*Note:ClamAV requires at least 1GB of RAM nad 1 CPU: https://docs.clamav.net/#recommended-system-requirements*

ClamAV version can be set by changing the `latest` tag to any version from [hub.docker.com](https://hub.docker.com/r/clamav/clamav/tags):
```
CLAMAV_VERSION="latest"
```

ClamAV will daily update the singatures database and keep it in RAM. To enable this feature set:

```
CLAMAV_DB_AUTOUPDATE="host"
```

Admins can specify the delay in seconds to wait for the signature database to load before starting the deamon:

```
CLAMD_STARTUP_DELAY=30
```


## Redis

Redis service is used by the OpenPanel user interface to keep user session data and temporary files in the meomory.

Admins can set memory (RAM) and CPU limits for Redis service by modifying the `/root/.env` file:

```
CLAMAV_RAM="1g"
CLAMAV_CPUS="1.0"
```

By default `/tmp/redis/` is used for the redis.sock file for conencting the OpenPanel and Redis services. This path can be changed by editing:
```
REDIS_SOCKET="/tmp/redis/"
```

## FTP

[vsftpd](https://security.appspot.com/vsftpd.html) is used as FTP server for the Enterprise edition. OpenPanel provides terminal commands and UI for managing ftp sub-accounts.

Admins can set memory (RAM) and CPU limits for FTP service by modifying the `/root/.env` file:

```
FTP_RAM="0.5g"
FTP_CPUS="0.5"
```

Port range for passive ftp can also be configured:
```
FTP_PORT_RANGE="21000-21010"
```

## MailServer

OpenPanel Enterprise supports mailserver. [Docker Mailserver](https://docker-mailserver.github.io/docker-mailserver/latest/) is used as a fullstack mail server (SMTP, IMAP, LDAP, Anti-spam, Anti-virus, etc.)

OpenPanel provides [terminal commands](/cli/email.html) and [UI](https://openpanel.com/docs/panel/intro/) for users to manage emails.

## Webmail

[Roundcube](https://roundcube.net/) is provided as a Webmail client
