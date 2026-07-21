# OpenPanel Folder Structure

OpenPanel uses the following files and directories:

Configuration:
```
/etc/openpanel/
├── bind9/                    - Configuration for DNS service.
├── caddy/                    - Caddy webserver configuration and data.
├── clamav/                   - ClamAV service configuration.
├── docker/                   - Docker configuration.
├── ftp/                      - Settings and user data for FTP service.
├── goaccess/                 - Configuration for GoAccess reports.
├── mysql/                    - Database schema.
├── nginx/                    - Nginx webserver configuraton [DEPRECATED].
├── openadmin/                - Settings for the admin panel.
├── openpanel/                - Settings for the user panel.
├── php/                      - PHP blueprints.
├── services/                 - Services configuration files.
├── skeleton/                 - Files used for new user accounts.
├── ssh/                      - SSH configuration and motd.
└── varnish/                  - Varnish configuration - [DEPRECATED].
```

Default configuration can be viewed [here](https://github.com/stefanpejcic/OpenPanel-configuration).


## User Files

All files for a user are stored in their /home directory: `/home/$docker_context/`:

```
/home/stefan/
├── docker-compose.yml        - Contains all services that user can run.
├── .env                      - Contains ports, resource limits and passwrods for services.
├── httpd.conf                - main configuration that is used IF user runs Apache webserver.
├── nginx.conf                - main configuration that is used IF user runs Nginx webserver.
├── openresty.conf            - main configuration that is used IF user runs OpenResty webserver.
├── openlitespeed.conf        - main configuration that is used IF user runs OpenLitespeed webserver.
├── default.vcl               - Varnish Configuration that is used if Varnish cache is enabled for user.
├── php.ini/                  - PHP.INI files hat user can edit via 'OpenPanel > PHP > PHP.INI Editor' page.
├── crons.ini                 - Cron jobs that user can edit via 'OpenPanel > Cron Jobs' page.
├── backup.env                - Backup configuration that user can edit via 'OpenPanel > Files > Backups' page.
├── my.cnf                    - MySQL root user logins that are used for ALL actions from the OpenPanel UI.
├── custom.cnf                - MySQL Configuration that user can edit via 'OpenPanel > MySQL > Configuration' page.
├── postgre_custom.conf       - PostgreSQL Configuration that user can edit via 'OpenPanel > PostgreSQL > Configuration' page.
├── pma.php                   - used to allow auto-login to phpMyAdmin from OpenPanel UI.
├── docker-data/              - All docker files: containers, images, volumes..
│   └── volumes/ 
│       ├── $context_html_data/             - Website files (/var/www/html/ in 'OpenPanel > File Manager' page).
│       ├── $context_mysql_data/            - MySQL storage (all databases).
│       ├── $context_mysql_dumps/           - MySQL /tmp folder used for dumps.
│       ├── $context_pg_data/               - PostgreSQL storage (all databases).
│       └── $context_webserver_data/        - VirtualHosts files for domains.
├── .config/                  - systemd service files.
├── .docker/                  - Docker service files.
├── bin/                      - service files.
├── sockets/                  - sockets for running services.
└── .bashrc                   - Sets path and docker host.
```

## User Services files

Certain files are shared across user services to provide Administrators with a single point of management. This enables faster updates and easier customization for all users. For example, it allows updating WP-CLI simultaneously for everyone or configuring OpenLiteSpeed consistently across all accounts.

- `/etc/openpanel/wordpress/wp-cli.phar`           - WP-CLI binnary used in OpenLitespeed and all PHP containers.
- `/etc/openpanel/openlitespeed/start.sh`          - Sets permissions and maps all individual VHosts files on OpenLitespeed.
- `/etc/openpanel/nginx/certs/`                    - Self-signed SSL files for domains used inside user webservers.
- `/etc/openpanel/nginx/default_page.html`         - Landing page displayed on user domains if no files exist.
- `/etc/openpanel/mysql/scripts/dump.sh`           - Script used to generate mysql dumps when performing backups.
- `/etc/openpanel/mysql/phpmyadmin/config.inc.php` - phpMyAdmin configuration file that allows auto-login from UI.
- `/etc/openpanel/php/ioncube/`                    - Ioncube Loader files for PHP.
- `/etc/openpanel/php/ioncube_extension.ini`       - Ioncube Loader extension for PHP.

## Shared Services files

Files that are used by shared services like FTP, Caddy webserver and Mailserver.

- `/etc/openpanel/wordpress/wp-cli.phar`           - WP-CLI binnary used in OpenLitespeed and all PHP containers.
- `/etc/openpanel/wordpress/wp-cli.phar`           - WP-CLI binnary used in OpenLitespeed and all PHP containers.

## Templates files

Files that are used for new services/domains that users add.

- `/etc/openpanel/docker/compose/nodejs.yml`       - Used for new docker services via 'OpenPanel > AutoInstaller > NodeJS' page.
- `/etc/openpanel/docker/compose/python.yml`       - Used for new docker services via 'OpenPanel > AutoInstaller > Python' page.
- `/etc/openpanel/docker/daemon/rootless.json`     - Used for new users to configure docker rootless service.
- `/etc/openpanel/apache/httpd.conf`     - Main configuration file used for new users with Apache webserver.
- `/etc/openpanel/backups/backup.env`     - Template used for new user backups.
- `/etc/openpanel/backups/customized.template`     - Notifications template used for user backups.
- `/etc/openpanel/bind9/zone_template.txt` - DNS zone template used for new domains with IPv4 set for the user.
- `/etc/openpanel/bind9/zone_template_ipv6.txt` - DNS zone template used for new domains with IPv6 set for the user.
- `/etc/openpanel/bind9/zone_template_ipv6.txt` - DNS zone template used for new domains with IPv6 set for the user.
- `/etc/openpanel/caddy/templates/domain.conf` - Caddyfile used for new domains.
- `/etc/openpanel/caddy/templates/suspended_domain.html` - Landing page displayed on suspended domains for a user.
- `/etc/openpanel/caddy/templates/suspended_user.html` - Landing page displayed on domains for a suspended user.
- `/etc/openpanel/csf/csf.blocklists` - CSF blocklists used when configuring Sentinel Firewall.
- `/etc/openpanel/mysql/user.cnf` - Configuration file used for MySQL services.


## Configuration files

- `/etc/openpanel/caddy/templates/wp.rules` - Rules used for hardening options in WP Manager.
- `/etc/openpanel/clamav/extensions.txt` - List of extensions to scan using ClamAV (DEPRECATED).
- `/etc/openpanel/csf/csf.blocklists` - CSF blocklists used when configuring Sentinel Firewall.
- `/etc/openpanel/ftp/start_vsftpd.sh` - Entrypoint for FTP service (creates accounts on startup).
- `/etc/openpanel/ftp/vsftpd.conf` - Vsftpd main configuration file.
- `/etc/openpanel/goaccess/goaccess.conf` - GoAccess main configuration file.
- `/etc/openpanel/modules/` - Custom plugins for OpenPanel UI.
- `/etc/openpanel/mysql/initialize/1.1/plans.sql` - Initializes database on OpenPanel installation.
- `/etc/openpanel/mysql/host_mysql.cnf` - Stores mysql logins for the OpenPanel installation.
- `/etc/openpanel/mysql/keys.txt` - List of configuration keys editable from 'OpenPanel > MySQL > Configuration' page.
- `/etc/openpanel/mysql/mysqld.cnf` - Configuration file for MySQL service.
- `/etc/openpanel/docker/templates/docker_apache_domain.conf` - VirtualHosts template for new domains on Apache webserver.
- `/etc/openpanel/docker/templates/docker_litespeed_domain.conf` - VirtualHosts template for new domains on OpenLitespeed webserver.
- `/etc/openpanel/docker/templates/docker_nginx_domain.conf` - VirtualHosts template for new domains on Nginx webserver.
- NOTE: list is incomplete, please refer to https://github.com/stefanpejcic/openpanel-configuration/tree/main for a full list.
