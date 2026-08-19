# Configuration Files Reference

This page lists the configuration files used across the OpenPanel stack: the **OpenPanel** user panel, **OpenAdmin** admin panel, **opencli** command-line tool, and the default files shipped by [openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration), deployed to `/etc/openpanel/` on install.

Unless noted otherwise, all paths are on the **host filesystem**, not inside containers. Files under `/etc/openpanel/` are considered defaults/templates — most are copied into a user's home directory (`/home/<username>/...`) on account creation, where the user's own copy takes over.

:::danger
Manually editing these files can break the panel or be overwritten during updates. Prefer changing settings from **OpenAdmin > Settings** (or the relevant panel UI) whenever an option is exposed there.
:::

## Panel core settings

| Path | Used by | Format | Purpose |
|---|---|---|---|
| `/etc/openpanel/openpanel/conf/openpanel.config` | OpenPanel, OpenAdmin, opencli | `key=value` | Main panel settings: license key, enabled modules, session/2FA settings, nameservers, update settings |
| `/etc/openpanel/openpanel/secret.key` | OpenPanel, OpenAdmin | raw text (hex) | Signing key for session cookies/CSRF, also reused as the Dovecot master-password source |
| `/etc/openpanel/openpanel/default_locale` | OpenPanel | text | Default UI locale |
| `/etc/openpanel/openpanel/translations/` | OpenPanel, OpenAdmin | dir | Translation catalogs |
| `/etc/openpanel/openpanel/static/` | OpenPanel | dir | Override dir for `css/custom.css`, `js/custom.js`, `robots.txt`, `security.txt` |
| `/etc/openpanel/openpanel/features/default.txt`, `<plan>.txt` | OpenPanel, OpenAdmin | text list | Feature flags available per plan |
| `/etc/openpanel/openpanel/quota_report.json` | OpenPanel, OpenAdmin | JSON | Cached disk/quota usage report |
| `/etc/openpanel/openpanel/conf/custom_dashboard_section.json` | OpenPanel, OpenAdmin | JSON | Admin-injected custom dashboard content |
| `/etc/openpanel/openpanel/conf/knowledge_base_articles.json` | OpenPanel, OpenAdmin, opencli | JSON | Cached "how-to" / knowledge base links |
| `/etc/openpanel/openpanel/conf/domain_restriction.txt` | OpenAdmin, opencli | text list | Domains forbidden from being added by users |
| `/etc/openpanel/openpanel/conf/blacklist_useragents.txt` | OpenAdmin, opencli | text list | Blocked user-agents (bot/scraper blocking) |
| `/etc/openpanel/openpanel/conf/public_suffix_list.dat` | opencli | text | Public suffix (TLD) list, used for DNS zone-apex handling |
| `/etc/openpanel/openpanel/custom_code/custom.css`, `custom.js`, `in_header.html`, `in_footer.html` | OpenPanel, OpenAdmin | CSS/JS/HTML | User-editable custom code injected into the panel UI |
| `/etc/openpanel/openpanel/service/pagespeed.api` | OpenPanel, OpenAdmin | text | PageSpeed Insights API key |
| `/etc/openpanel/openpanel/service/service.config.py` | (system) | Python config | systemd/service config, fetched during updates |
| `/etc/openpanel/modules/` | OpenPanel, OpenAdmin | dir | Third-party/custom plugin modules (`<plugin>/readme.txt` metadata) |
| `/etc/openpanel/skeleton/` | opencli | dir | Template files copied into every new user's account directory |
| `/etc/openpanel/no_port` | opencli | flag file | Marker file |
| `/etc/openpanel/upgrade/skip_versions` | opencli | text list | Versions to skip during `opencli update` |

### Per-user account data — `/etc/openpanel/openpanel/core/users/<username>/`

| Path | Format | Purpose |
|---|---|---|
| `ip.json` | JSON | User's dedicated IP assignment |
| `.lastlogin` | text | Last login timestamp |
| `aliases.yml` | YAML | Cached email alias list |
| `emails.yml` | YAML | Cached email account list |
| `notifications.yaml` | YAML | Per-user notification preferences |
| `ip.json`, `activity.log`, `custom_message.html` | JSON/text/HTML | Allowed IPs, activity log, suspension message (OpenAdmin) |

### Per-website data

| Path | Purpose |
|---|---|
| `/etc/openpanel/openpanel/websites/` | Per-website metadata directory |
| `/var/www/html/<domain>` | Document root (base dir for File Manager) |

## OpenAdmin settings

| Path | Format | Purpose |
|---|---|---|
| `/etc/openpanel/openadmin/config/admin.ini` | INI | Main OpenAdmin settings (incl. `email_storage_location`) |
| `/etc/openpanel/openadmin/secret.key` | text | Session/auth secret key |
| `/etc/openpanel/openadmin/users.db` | SQLite | OpenAdmin user accounts database |
| `/etc/openpanel/openadmin/config/notifications.ini` | INI | Sentinel/notification thresholds |
| `/etc/openpanel/openadmin/config/services.json` | JSON | Service enable/status list |
| `/etc/openpanel/openadmin/config/features.json` | JSON | Feature toggle config |
| `/etc/openpanel/openadmin/config/log_paths.json` | JSON | Custom log file paths shown in the log viewer |
| `/etc/openpanel/openadmin/config/forbidden_usernames.txt` | text list | Blacklisted usernames for new accounts |
| `/etc/openpanel/openadmin/config/reseller_template.json` | JSON | Default template applied to new resellers |
| `/etc/openpanel/openadmin/config/shortcuts.json` | JSON | UI shortcut/menu definitions |
| `/etc/openpanel/openadmin/config/terms.md` | Markdown | Terms of Service text shown at signup |
| `/etc/openpanel/openadmin/config/quick_start.dismissed`, `tour.skip` | flag file | Onboarding/tour dismissal flags |
| `/etc/openpanel/openadmin/ssh_whitelist.conf` | text | SSH login IP whitelist |
| `/etc/openpanel/openadmin/resellers/<name>.json` | JSON | Per-reseller configuration (allowed plans, etc.) |
| `/etc/openpanel/openadmin/usage_stats.json` | JSON | Usage/telemetry state |
| `/etc/openpanel/openadmin/service/openadmin.service` | systemd unit | Installed as the `admin` systemd service |
| `/usr/local/admin/custom.css` | CSS | Custom OpenAdmin UI override |
| `/usr/local/admin/modules/api/available_endpoints.txt` | text | Enabled API endpoints |
| `/usr/local/admin/modules/security/csf.pl` | script | ConfigServer Firewall control script |
| `/usr/local/admin/core/search/filtered.json`, `filter.json` | JSON | Command-palette / search index |
| `/usr/local/admin/templates/emails/` | dir | Admin email report templates |

## Docker / Podman & containers

| Path | Format | Purpose |
|---|---|---|
| `/etc/openpanel/docker/compose/docker-compose.yml`, `.env` | YAML/env | Main podman/docker-compose stack (mysql, redis, openpanel, caddy, bind9, ftp, phpMyAdmin, clamav) |
| `/etc/openpanel/docker/compose/nodejs.yml`, `python.yml` | YAML | Compose overlays for Node.js/Python app hosting |
| `/etc/openpanel/docker/compose/1.0/docker-compose.yml`, `.env`, `autostart.services` | YAML/env/text | Per-account container stack template (legacy "1.0" webserver stack) |
| `/etc/openpanel/docker/templates/docker_*.conf` | conf | Per-website nginx-vhost snippets matching the container backend |
| `/etc/openpanel/docker/daemon/rootless.json` | JSON | Rootless Docker daemon config template |
| `/etc/openpanel/docker/overlay2/*.json` | JSON | `/etc/docker/daemon.json` templates per filesystem/distro |
| `/etc/openpanel/docker/modprobe.txt` | text | Kernel modules loaded at boot (symlinked to `/etc/modules-load.d/podman.conf`) |
| `/etc/openpanel/docker/selinux/pasta_local.pp` | SELinux policy | Rootless podman networking policy |
| `/root/.env` | env | Root-level docker-compose environment (e.g. `CADDY_IMAGE`) |
| `/root/docker-compose.yml` | YAML | Host-level compose stack for core services |
| `/home/<user>/.env` | env | Per-user container stack env vars (`WEB_SERVER`, `MYSQL_TYPE`, etc.) |
| `/home/<user>/docker-compose.yml` | YAML | Per-user container stack definition |
| `/home/<user>/.config/containers/{storage,containers}.conf` | conf | Rootless podman container config |

## Web servers & reverse proxy (Caddy, Nginx, Apache, OpenLiteSpeed, OpenResty, Varnish)

| Path | Format | Purpose |
|---|---|---|
| `/etc/openpanel/caddy/Caddyfile` (mounted as `/etc/caddy/Caddyfile` in container) | Caddyfile | Main reverse-proxy config for all websites |
| `/etc/openpanel/caddy/domains/<domain>.conf` | Caddyfile snippet | Per-domain vhost config, incl. WAF toggles |
| `/etc/openpanel/caddy/suspended_domains/<domain>.conf` | Caddyfile snippet | Vhost for suspended domains |
| `/etc/openpanel/caddy/templates/*.conf`, `*.html`, `wp.rules` | templates | Domain, suspended-user/domain, WordPress vhost templates |
| `/etc/openpanel/caddy/redirects.conf` | Caddyfile snippet | Global HTTP redirects (webmail, OpenAdmin port) |
| `/etc/openpanel/caddy/coraza_rules.conf`, `check.conf` | Coraza rules | WAF rule set |
| `/etc/openpanel/caddy/coreruleset/rules/*.conf[.disabled]` | OWASP CRS | WAF (Coraza) core rule set |
| `/etc/openpanel/caddy/ssl/custom/`, `ssl/certs/`, `ssl/acme-v02.api.letsencrypt.org-directory/` | dir | Custom and Let's Encrypt SSL certificate storage |
| `/etc/openpanel/nginx/nginx.conf` | conf | Default Nginx template |
| `/etc/openpanel/nginx/user-nginx.conf` | conf | Per-user Nginx include, copied to `/home/<user>/nginx.conf` |
| `/etc/openpanel/nginx/vhosts/openpanel_proxy.conf` | conf | Reverse-proxy vhost for the panel itself |
| `/etc/openpanel/nginx/vhosts/1.1/docker_*.conf`, `nginx_proxy_headers.txt` | conf | Per-website nginx-vhost templates (current stack) |
| `/etc/openpanel/nginx/vhosts/docker_*.conf` | conf | Per-website nginx-vhost templates (legacy "1.0" stack) |
| `/etc/openpanel/nginx/certs/cert.crt`, `cert.key` | PEM | Default/self-signed TLS cert used before a real cert is issued |
| `/etc/openpanel/nginx/default_page.html`, `suspended_user.html`, `suspended_website.html` | HTML | Placeholder/suspension pages |
| `/etc/openpanel/nginx/cloudflare.inc` | conf | Cloudflare real-IP restore config |
| `/etc/openpanel/nginx/options-ssl-nginx.conf` | conf | Recommended TLS options |
| `/etc/openpanel/nginx/modsecurity/*.conf` | ModSecurity rules | WAF rules for the nginx engine |
| `/etc/openpanel/apache/httpd.conf` | conf | Apache template, copied to `/home/<user>/httpd.conf` |
| `/etc/openpanel/openlitespeed/httpd_config.conf` | conf | OpenLiteSpeed template, copied to `/home/<user>/openlitespeed.conf` |
| `/etc/openpanel/openlitespeed/start.sh` | script | Container entrypoint |
| `/etc/openpanel/openresty/nginx.conf` | conf | OpenResty template, copied to `/home/<user>/openresty.conf` |
| `/etc/openpanel/varnish/default.vcl`, `default` | VCL | Varnish cache config, copied to `/home/<user>/default.vcl` |
| `/home/<user>/nginx.conf`, `openresty.conf`, `openlitespeed.conf`, `httpd.conf`, `default.vcl` | — | Per-user generated webserver configs |
| `/var/log/caddy/domlogs/` | dir | Per-domain access logs |
| `/var/log/caddy/stats/` | dir | GoAccess stats data |
| `/var/log/caddy/coraza_waf/<domain>.log` | JSON-lines | Coraza WAF access log |
| `/etc/openpanel/goaccess/goaccess.conf` | conf | GoAccess web-log analytics config |
| `/etc/openpanel/goaccess/GeoLite2-*.tar.gz` | MaxMind DB | GeoIP data for GoAccess |

## DNS (BIND9)

| Path | Format | Purpose |
|---|---|---|
| `/etc/bind/named.conf`, `named.conf.local`, `named.conf.options`, `named.conf.default-zones` | BIND conf | Core BIND9 configuration, DNS cluster settings |
| `/etc/bind/zones/<domain>.zone` | BIND zone | Per-domain DNS zone file |
| `/etc/bind/zones/backups/` | dir | Zone file backups |
| `/etc/bind/rndc.key` | key | BIND RNDC key |
| `/etc/openpanel/bind9/zone_template.txt`, `zone_template_ipv6.txt` | template | Templates used when generating new DNS zones |

## Mail (OpenMail / Dovecot / Postfix)

| Path | Format | Purpose |
|---|---|---|
| `/usr/local/mail/openmail/compose.yml` | YAML | Mail server docker-compose stack |
| `/usr/local/mail/openmail/mailserver.env`, `.env` | env | docker-mailserver environment config |
| `/usr/local/mail/openmail/docker-data/dms/config/postfix-accounts.cf` | conf | Postfix mail account list |
| `/usr/local/mail/openmail/docker-data/dms/config/postfix-regex.cf` | conf | Postfix regex/alias rules |
| `/usr/local/mail/openmail/docker-data/dms/config/opendkim/keys/` | dir | Per-domain DKIM signing keys |
| `/usr/local/mail/openmail/docker-data/dms/dovecot/userdb` | conf | Dovecot mailbox user database |
| `/usr/local/mail/openmail/docker-data/dms/mail-logs/mail.log` | log | Mail server log |
| `/usr/local/mail/openmail/postfwd/postfwd.cf` | conf | Postfwd policy config (per-domain sending limits) |
| `/etc/dms-settings` | conf | docker-mailserver settings snapshot |
| `/etc/openpanel/email/snappymail` | dir | SnappyMail webmail config |

## FTP

| Path | Format | Purpose |
|---|---|---|
| `/etc/openpanel/ftp/vsftpd.conf` (mounted as `/etc/vsftpd/vsftpd.conf`) | conf | vsftpd server config |
| `/etc/openpanel/ftp/all.users` | text | Master FTP user list |
| `/etc/openpanel/ftp/users/` | dir | Per-user FTP account config |
| `/etc/openpanel/ftp/server.conf` | conf | FTP server settings (e.g. hostname) |
| `/etc/openpanel/ftp/filezilla.conf`, `cyberduck.conf` | template | FTP client-config templates served to users |
| `/etc/openpanel/ftp/start_vsftpd.sh` | script | Container entrypoint |

## PHP

| Path | Format | Purpose |
|---|---|---|
| `/etc/openpanel/php/ini/<version>.ini` | php.ini | Default php.ini per PHP version (5.6–8.5), mounted into containers |
| `/etc/openpanel/php/options.txt` | text | Default PHP option definitions shown in the panel |
| `/etc/openpanel/php/ini.txt`, `extensions.txt` | text | Selectable php.ini directives / extensions |
| `/etc/openpanel/php/extensions_table.json` | JSON | Table of installable PHP extensions |
| `/etc/openpanel/openpanel/php/api_versions.json` | JSON | Cached list of available PHP versions |
| `/etc/openpanel/php/composer/*.phar` | binary | Composer versions installed into user containers |
| `/etc/openpanel/php/ioncube/` | dir | ionCube loaders per PHP version |
| `/home/<user>/php.ini/<version>.ini` | php.ini | Per-user, per-version PHP config |
| `/home/<user>/php.ini/options.txt` | text | Per-user PHP option overrides |

## Databases (MySQL/MariaDB, PostgreSQL)

| Path | Format | Purpose |
|---|---|---|
| `/etc/my.cnf` | my.cnf (INI) | Panel's own MySQL client credentials (`[client]` section) |
| `/etc/openpanel/mysql/host_my.cnf`, `container_my.cnf` | my.cnf | Host- and container-level MySQL client auth |
| `/etc/openpanel/mysql/mysqld.cnf`, `user.cnf` | conf | Server/user database config |
| `/etc/openpanel/mysql/keys.txt` | text | Allow-list of editable `my.cnf` keys for users |
| `/etc/openpanel/mysql/initialize/1.1/plans.sql` | SQL | Seeds the panel database with default hosting plans |
| `/etc/openpanel/mysql/scripts/dump.sh` | script | DB dump/backup helper mounted into containers |
| `/etc/openpanel/mysql/phpmyadmin/pma.php`, `config.inc.php` | PHP conf | phpMyAdmin container config |
| `/home/<user>/custom.cnf` | my.cnf | User's MySQL resource-limit config override |
| `/home/<user>/my.cnf` | my.cnf | User's MySQL root client credentials |
| `/etc/openpanel/postgres/postgresql.conf` | conf | Default PostgreSQL config template |
| `/etc/openpanel/postgres/keys.txt` | text | Allow-list of editable PostgreSQL config keys |
| `/home/<user>/postgre_custom.conf` | conf | Per-user PostgreSQL config override |

## Cron (Ofelia)

| Path | Format | Purpose |
|---|---|---|
| `/etc/cron.d/openpanel` | crontab | System-wide cron jobs for OpenPanel (do not edit manually) |
| `/etc/openpanel/ofelia/users.ini` | INI | Template cron config, copied to `/home/<user>/crons.ini` |
| `/home/<user>/crons.ini` | INI | Per-user scheduled job definitions |

## Backups

| Path | Format | Purpose |
|---|---|---|
| `/etc/openpanel/backups/backup.env` | env | Default backup destination/schedule template |
| `/etc/openpanel/backups/customized.template` | Go template | Backup success/failure notification message template |
| `/home/<user>/backup.env` | env | Per-user backup destination config (S3/WebDAV/SSH/Azure) |

## Firewall & security

| Path | Format | Purpose |
|---|---|---|
| `/etc/csf/csf.conf` | conf | ConfigServer Firewall main config |
| `/etc/csf/csf.deny` | text | CSF deny list |
| `/etc/openpanel/csf/csf.blocklists` | text | CSF blocklist sources |
| `/etc/sysconfig/imunify360/integration.conf` | conf | Imunify360 panel integration config |
| `/etc/sysconfig/imunify360/malware-filters-admin-conf/ignored.txt` | text | Files ignored by malware scanning |
| `/etc/pam.d/imunify360-deny` | PAM conf | Deny rule for blocked SSH logins |
| `/etc/openpanel/clamav/domains.list`, `extensions.txt` | text | ClamAV scan target lists |
| `/etc/openpanel/wordpress/vulnerability/` | dir | Cached WordPress vulnerability database |
| `/etc/ssh/sshd_config` | conf | SSH daemon config (port, root login) |
| `~/.ssh/authorized_keys` | OpenSSH | Per-user SSH keys |

## WordPress & app installers

| Path | Format | Purpose |
|---|---|---|
| `/etc/openpanel/wordpress/htaccess/apache.htaccess`, `litespeed.htaccess` | .htaccess | Default `.htaccess` per webserver engine |
| `/etc/openpanel/wordpress/mu-plugin.php` | PHP | Must-use plugin injected into new WordPress installs |
| `/etc/openpanel/wordpress/sets/plugins.txt`, `themes.txt` | text list | Default plugin/theme bundles offered in the installer |
| `/etc/openpanel/wordpress/wp-cli.phar` (mounted as `/usr/local/bin/wp`) | binary | WP-CLI |
| `/etc/openpanel/<app>/archives/` (wordpress, joomla, matomo, mediawiki, moodle, nextcloud, opencart, prestashop) | dir | Cached installer archives per application |
| `/etc/openpanel/matomo/credentials` | text | Stored Matomo admin credentials |

## System integration

| Path | Format | Purpose |
|---|---|---|
| `/etc/passwd`, `/etc/shadow`, `/etc/group` | system | User/group accounts, read/modified by `opencli` user management |
| `/etc/subuid`, `/etc/subgid` | system | Subordinate UID/GID ranges for rootless podman |
| `/etc/fstab` | system | Checked/modified for storage/quota mounts |
| `/etc/hosts` | system | Updated with domain/IP mappings |
| `/etc/logrotate.d/openpanel`, `caddy-logs`, `syslog` | logrotate | Log rotation rules |
| `/etc/openpanel/ssh/admin_welcome.sh` | script | SSH login MOTD/welcome script |
| `/root/openpanel_restart_needed`, `openadmin_restart_needed`, `openadmin_is_disabled`, `disable_openadmin_reboot_ui`, `disable_openadmin_terminal_ui`, `disable_2087_port` | flag files | Restart/feature-disable state flags |
| `/tmp/.openpanel_license_cache.json` | JSON | Cached license validation result |

## Related documentation

* [OpenAdmin system crons](../server/openadmin-system-crons.md)
* [How to free up disk space on Linux](../server/how-to-free-up-disk-space-on-linux.md)
