# OpenCLI
OpenPanel uses OpenCLI, a command line interface designed for administrators to manage server from the terminal.

OpenCLI keeps history of used commands and allows you to view last and top 5 used commands:
```bash
# opencli

Usage: opencli <COMMAND> [additional_arguments]

Suggested commands:
  opencli faq                       Display frequently asked questions and answers.
  opencli --help                    List all available OpenCLI commands and their usage.

Recently used commands:
  opencli docker-collect_stats
  opencli faq
  opencli update
  opencli update_check
  opencli user-add

Most commonly used commands:
  opencli admin
  opencli docker-collect_stats
  opencli faq
  opencli update
  opencli update_check
```

OpenCLI ships with real bash tab-completion (`lib/completion.bash`, installed to `/etc/bash_completion.d/opencli`). Typing `opencli` and pressing `TAB` suggests all available commands, and pressing `TAB` again after a command suggests its actual arguments — usernames, domain names, plan names, PHP versions, admin logins, FTP accounts, config/notification setting names, and more, all pulled live from the panel database and filesystem so suggestions always match what's actually on the server. It's installed and kept up to date automatically by `opencli update`, which also installs the `bash-completion` package if it's missing. Open a new shell session after updating for it to take effect.

## Help

`opencli help` command generates a list of available OpenCLI commands on the current server:

```
opencli --help
```

As new commands are added gradually, they might not yet be available in your OpenPanel version. To check if a command is available on a server, run `opencli <command> --help`.

<details>
  <summary>Example output</summary>

```bash
opencli server-ips
Description: Generates a file with a list of users with their dedicated IPs
Usage: opencli server-ips
------------------------
opencli server-logrotate
Description: Configures logrotate for caddy, openpanel, syslog.
Usage: opencli server-logrotate
------------------------
opencli server-migrate
Description: Migrates all data from this server to another.
Usage: opencli server-migrate -h <DESTINATION_IP> --user root --password <DESTINATION_PASSWORD> [--force] [--exclude-* options]
------------------------
opencli ftp-add
Description: Create FTP sub-user for openpanel user.
Usage: opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME>
------------------------
opencli ftp-delete
Description: Delete FTP sub-user for openpanel user.
Usage: opencli ftp-delete <username> <openpanel_username> [--debug]
------------------------
opencli ftp-path
Description: Change FTP path for a user.
Usage: opencli ftp-path <username> <path> <openpanel_username> [--debug]
------------------------
opencli ftp-password
Description: Change password for FTP sub-user of openpanel user.
Usage: opencli ftp-password <username> <new_password> <openpanel_username> [--debug]
------------------------
opencli ftp-connections
Description: Display all active FTP connection or for particular OpenPanel user.
Usage: opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME>
------------------------
opencli ftp-logs
Description: View the FTP server log
Usage: opencli ftp-log
------------------------
opencli ftp-list
Description: List FTP sub-users for openpanel user.
Usage: opencli ftp-list [OPENPANEL_USERNAME] [--json]
------------------------
opencli patch
Description: Download and install a patch
Usage: opencli patch <NAME>
------------------------
opencli domain
Description: View and set domain/ip for accessing panels.
Usage: opencli domain [set <domain_name> | ip] [--debug]
------------------------
opencli phpmyadmin
Description: View and set domain/ip for phpMyAdmin access.
Usage: opencli phpmyadmin [set <domain_name> | ip]
------------------------
opencli api-list
Description: List available API endpoints with usage examples.
Usage: opencli api-list [--save]
------------------------
opencli api
Description: Check status, enable or disable OpenAdmin API access, and list available API endpoints.
Usage: opencli api <status|on|off|list>
------------------------
opencli license
Description: Manage OpenPanel Enterprise license.
Usage: opencli license verify 
------------------------
opencli faq
Description: Display answers to most frequently asked questions.
Usage: opencli faq
------------------------
opencli email-manage
Description: Manage mailserver configuration and overview.
Usage: opencli email-manage <COMMAND> <ATTRIBUTES>
------------------------
opencli email-server
Description: Manage mailserver
Usage: opencli email-server <status|config|install|start|stop|restart|queue|flush|view|unhold|delete|fail2ban|ports|logs|login|supervisor|postfwd|pflogsumm|update-check|update-packages|versions> [args] [--debug]
------------------------
opencli email-quotas
Description: Fixes email permission issues for all domains.
Usage: opencli email-quotas
------------------------
opencli email-ratelimit
Description: Configure rate-limiting using postfwd for domains and users.
Usage: opencli email-ratelimit
------------------------
opencli email-webmail
Description: Display or update the domain used for accessing webmail.
Usage: opencli email-webmail [--debug] [domain <domain>]
------------------------
opencli email-setup
Description: Setup email addresses, forwarders, filters..
Usage: opencli email-setup <COMMAND> <ATTRIBUTES>
------------------------
opencli update
Description: Check if update is available, install updates.
Usage: opencli update [--check | --force | --admin | --panel | --cli]
------------------------
opencli webserver-get_webserver_for_user
Description: View cached or check the installed webserver inside user container.
Usage: opencli webserver-get_webserver_for_user <USERNAME>
------------------------
opencli version
Description: Displays the current (installed) version of OpenPanel container image.
Usage: opencli version 
------------------------
opencli dev
Description: Overwrite OpenPanel container files and restart service to apply.
Usage: opencli dev [path]
------------------------
opencli sentinel
Description: Check system services, traffic and resource usage, and log/send custom notifications on request.
Usage: opencli sentinel [--startup] [--report] [--action=<name> --title=<title> --message=<msg>]
------------------------
opencli docker-collect_stats
Description: Collect resource usage information for user(s).
Usage: opencli docker-collect_stats <username|--all>
------------------------
opencli docker-usage_stats_cleanup
Description: Rotates resource usage logs for all users according to the resource_usage_retention setting.
------------------------
opencli docker-backup
Description: Generates a backup for all users.
Usage: opencli docker-backup
------------------------
opencli docker-autostart
Description: Prefetch (pull) container images for services listed as autostart, skipping if free disk space is low.
Usage: opencli docker-autostart [-f|--force]
------------------------
opencli docker-images
Description: Check images for updates.
Usage: opencli docker-images [--all|<USERNAME>]
------------------------
opencli docker-logs
Description: Display log sizes for user and system containers
Usage: opencli docker-logs [--all|--system|--users|<USERNAME>]
------------------------
opencli docker-limits
Description: Set global container limits for all containers combined.
Usage: opencli docker-limits [--apply | --apply SIZE | --read]
------------------------
opencli user-2fa
Description: Check or disable 2FA for a user.
Usage: opencli user-2fa <username> [disable]
------------------------
opencli user-redis
Description: Check and enable/disable REDIS for user.
Usage: opencli user-redis [check|enable|disable] <USERNAME>
------------------------
opencli user-loginlog
Description: View user's .lastlogin file with last 20 successfull logins.
Usage: opencli user-loginlog <USERNAME> [--table|--text|--json]
------------------------
opencli user-ssh
Description: Manage SSH settings for user.
Usage: opencli user-ssh <check|enable|disable> <username>
------------------------
opencli user-email
Description: Change email for user
Usage: opencli user-email <USERNAME> <NEW_EMAIL>
------------------------
opencli user-varnish
Description: Enable/disable Varnish Caching for user and display current status.
Usage: opencli user-varnish <USERNAME> [enable|disable|status]
------------------------
opencli user-add
Description: Create a new user with the provided plan_name.
Usage: opencli user-add <USERNAME> <PASSWORD|generate> <EMAIL> "<PLAN_NAME>" [--send-email] [--debug] [--webserver="<nginx|apache|openresty|openlitespeed|litespeed|varnish+nginx|varnish+apache|varnish+openresty|varnish+openlitespeed>"] [--sql=<mysql|mariadb>] [--reseller=<RESELLER_USERNAME>] [--private-note="this user.."] [--no-sentinel]
------------------------
opencli user-rename
Description: Rename username.
Usage: opencli user-rename <old_username> <new_username>
------------------------
opencli user-delete
Description: Delete user account and permanently remove all their data.
Usage: opencli user-delete <username> [-y] [--all]
------------------------
opencli user-resources
Description: View services limits for user.
Usage: opencli user-resources <CONTEXT> [--activate=<SERVICE_NAME>] [--deactivate=<SERVICE_NAME>] [--update_cpu=<FLOAT>] [--update_ram=<FLOAT>] [--service=<NAME>] [--json]
------------------------
opencli user-change_plan
Description: Change plan for a user and apply new plan limits.
Usage: opencli user-change_plan <USERNAME> <NEW_PLAN_NAME>
------------------------
opencli user-password
Description: Reset password for a user.
Usage: opencli user-password <USERNAME> <NEW_PASSWORD | RANDOM> [--ssh]
------------------------
opencli user-memcached
Description: Check and enable/disable Memcached for user.
Usage: opencli user-memcached [check|enable|disable] <USERNAME>
------------------------
opencli user-login
Description: Login as a user container.
Usage: opencli user-login <USERNAME>
------------------------
opencli user-unsuspend
Description: Unsuspend user: start all containers and unsuspend domains
Usage: opencli user-unsuspend <USERNAME>
------------------------
opencli user-ip
Description: Assing or remove dedicated IP to a user.
Usage: opencli user-ip <USERNAME> <IP | DELETE> [-y] [--debug]
------------------------
opencli user-block_ip
Description: View, add, or remove blocked IP addresses for a user's domains.
Usage: opencli user-block_ip <username> [--list='ip_here another_ip' | --delete-all]
------------------------
opencli user-suspend
Description: Suspend user: stop all containers and suspend domains.
Usage: opencli user-suspend <USERNAME>
------------------------
opencli user-list
Description: Display all users: id, username, email, plan, registered date.
Usage: opencli user-list [--json]
------------------------
opencli user-transfer
Description: Transfers a single user account from this server to another.
Usage: opencli user-transfer --account <OPENPANEL_USER> --host <DESTINATION_IP> --username <DESTINATION_SSH_USERNAME> --password <DESTINATION_SSH_PASSWORD> [--live-transfer]
------------------------
opencli user-backup
Description: Creates a full account .tar.gz backup of a single OpenPanel user account.
Usage: opencli user-backup --account <USER> [--output <DIR>] [--quiet]
------------------------
opencli user-restore
Description: Restores a single OpenPanel user account from a full account .tar.gz backup.
Usage: opencli user-restore --file <ARCHIVE> [--force] [--new-username=NAME] [--quiet] [--temp-dir=<PATH>]
------------------------
opencli user-check
Description: Performs comprehensive security checks on user files, rootless Podman instance and containers.
Usage: opencli user-check <USERNAME>
------------------------
opencli user-quota
Description: Enforce and recalculate disk and inodes for a user.
Usage: opencli user-quota <username|--all>
------------------------
opencli report
Description: Generate a system report and send it to OpenPanel support team.
Usage: opencli report [--public|--link|--upload] [--non-interactive]
------------------------
opencli websites-scan
Description: Scan user files for WP sites and add them to SiteManager interface.
Usage: opencli websites-scan <USERNAME|-all>
------------------------
opencli websites-all
Description: Lists all websites currently hosted on the server.
Usage: opencli websites-all [TYPE]
------------------------
opencli websites-user
Description: Lists all websites and domains owned by a specific user.
Usage: opencli websites-user <USERNAME> [--type=] [--domains=] [--json]
------------------------
opencli websites-pagespeed
Description: Check Google PageSpeed data for website(s)
Usage: opencli websites-pagespeed <DOMAIN> [-all]
------------------------
opencli websites-secure
Description: WP Manager security rules for domain.
Usage: opencli websites-secure <DOMAIN> [--rules='RULE1 RULE2' | --disable-all | --list-active-rules]
------------------------
opencli websites-vulnerability
Description: Scan WP site for vulnerabilities.
Usage: opencli websites-vulnerability <DOMAIN> [--all]
------------------------
opencli imunify
Description: Install and manage ImunifyAV service.
Usage: opencli imunify [status|start|stop|install|update|uninstall]
------------------------
opencli config
Description: Get or update a value in the OpenPanel configuration file.
Usage: opencli config [get|update] <parameter_name> [new_value]
------------------------
opencli docker
Description: Open an interactive shell inside a system or user container, picking the user and container via fzf if not given.
Usage: opencli docker [<user> [<container>]]
------------------------
opencli install
Description: Create cronjobs and configuration files needed for openpanel.
Usage: opencli install
------------------------
opencli imav-deploy
------------------------
opencli port
Description: View and change port for accessing openpanel.
Usage: opencli port [set <port>] 
------------------------
opencli plan-delete
Description: Delete hosting plan
Usage: opencli plan-delete <PLAN_NAME>
------------------------
opencli plan-list
Description: Display all plans: id, name, description, limits..
Usage: opencli plan-list [--json]
------------------------
opencli plan-edit
Description: Edit an existing hosting plan (Package) and modify its parameters.
Usage: opencli plan-edit --debug id=<ID> name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<DEFAULT>
------------------------
opencli plan-usage
Description: Display all users that are currently using the plan.
Usage: opencli plan-usage <PLAN_NAME> [--json]
------------------------
opencli plan-apply
Description: Change plan for a user and apply new plan limits.
Usage: opencli plan-apply <NEW_PLAN_ID> <USERNAME>
------------------------
opencli plan-create
Description: Create a new hosting plan (Package) and set its limits.
Usage: opencli plan-create name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME>
------------------------
opencli php-domain
Description: View or change the PHP version used for a single domain name.
Usage: opencli php-domain <domain_name>
------------------------
opencli php-ini
Description: View or change php.ini values of any php version for a user.
Usage: opencli php-ini <username> <action> <setting> [value]
------------------------
opencli php-default
Description: View or change the default PHP version used for new domains added by user.
Usage: opencli php-default <username>
------------------------
opencli admin
Description: Manage OpenAdmin service and Administrators.
Usage: opencli admin <command> [options]
------------------------
opencli waf
Description: Manage CorazaWAF
Usage: opencli waf <setting> 
------------------------
opencli files-calculate_resellers_storage
Description: Calculates total disk usage for all resellers.
Usage: opencli files-calculate_resellers_storage
------------------------
opencli files-fix_permissions
Description: Fix permissions for users /home directory files inside the container.
Usage: opencli files-fix_permissions <USERNAME|--all> [PATH] [--debug]
------------------------
opencli files-purge_trash
Description: Auto-purge .Trash folders for users.
Usage: opencli files-purge_trash [--user USERNAME] [--force] [--dry-run]
------------------------
opencli firewall-reset
Description: Deletes all container related ports from CSF and opens exposed ports.
------------------------
opencli domains-dnssec
Description: Enable DNSSEC for a domain and re-sign after changes in the zone.
Usage: opencli domains-dnssec <DOMAIN> [--update | --check]
------------------------
opencli domains-update_ns
Description: Change nameservers for a single or all dns zones.
Usage: opencli domains-update_ns <DOMAIN_NAME>
------------------------
opencli domains-all
Description: Lists all domain names currently hosted on the server.
Usage: opencli domains-all [--docroot] [--php_version] [--json]
------------------------
opencli domains-varnish
Description: Check Varnish status for domain, enable/disable Varnish caching.
Usage: opencli domains-varnish <DOMAIN-NAME> [on|off] [--short]
------------------------
opencli domains-hsts
Description: Manage HSTS for a domain
Usage: opencli domains-hsts <domain> [enable|disable]
------------------------
opencli domains-add
Description: Add a domain name for user.
Usage: opencli domains-add <DOMAIN_NAME> <USERNAME> [--docroot DOCUMENT_ROOT] [--php_version N.N] [--skip_caddy] [--skip_vhost] [--skip_containers] [--skip_dns] [--skip-sentinel] [--debug] [--hs_ed25519_public_key KEY --hs_ed25519_secret_key KEY]
------------------------
opencli domains-delete
Description: Delete a domain name.
Usage: opencli domains-delete <DOMAIN_NAME> --debug
------------------------
opencli domains-whoowns
Description: Check which username owns a certain domain name.
Usage: opencli domains-whoowns <DOMAIN-NAME> [--context] [--docroot]
------------------------
opencli domains-unsuspend
Description: Unsuspend a domain name
Usage: opencli domains-unsuspend <DOMAIN-NAME>
------------------------
opencli domains-dns
Description: Manage DNS zones and the DNS server via subcommands (create, delete, reload, list, check, start/stop/restart, etc).
Usage: opencli domains-dns <reconfig|check|reload|show|list|create|delete|default|count|config|start|restart|hard-restart|stop> [DOMAIN]
------------------------
opencli domains-user
Description: Lists all domain names currently owned by a specific user.
Usage: opencli domains-user <USERNAME> [--docroot|--php_version]
------------------------
opencli domains-suspend
Description: Suspend a domain name
Usage: opencli domains-suspend <DOMAIN-NAME>
------------------------
opencli domains-ssl
Description: Check SSL for domain, add custom certificate, view files.
Usage: opencli domains-ssl <DOMAIN_NAME> [status|info|logs|auto|custom] [path/to/fullchain.pem path/to/key.pem]
------------------------
opencli domains-edit
Description: Open a domain's docroot directory, or its webserver vhost config file in nano with --ws.
Usage: opencli domains-edit <DOMAIN_NAME> [--ws]
------------------------
opencli domains-docroot
Description: View and change docroot for a domain.
Usage: opencli domains-docroot <DOMAIN_NAME> [update </var/www/html/>] --debug
------------------------
opencli domains-stats
Description: Parse caddy access logs for users domains and generate static html
Usage: opencli domains-stats
------------------------
opencli proxy
Description: View and change proxy path '/openpanel' for accessing openpanel.
Usage: opencli proxy [set <path>|default]
------------------------
opencli error
Description: Displays information for specific error ID received in OpenPanel UI.
Usage: opencli error <ID_HERE>
------------------------
opencli locale
Description: Install locales (Languages) for OpenPanel UI.
Usage: opencli locale <CODE>
```
</details>
