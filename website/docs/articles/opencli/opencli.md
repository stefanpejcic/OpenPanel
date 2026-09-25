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
opencli admin
Description: Manage OpenAdmin service and Administrators.
Usage: opencli admin <command> [options]
------------------------
opencli api
Description: Check status, enable or disable OpenAdmin API access, and list available API endpoints.
Usage: opencli api <status|on|off|list>
------------------------
opencli backup
Description: Backs up (and restores) OpenPanel/OpenAdmin SYSTEM AND USER
Usage: opencli backup [--restore <ARCHIVE>] [--quiet]
------------------------
opencli config
Description: Get or update a value in the OpenPanel configuration file.
Usage: opencli config <get|update> <setting_name> [new_value]
------------------------
opencli docker
Description: Open an interactive shell inside a system or user container, picking the user and container via fzf if not given.
Usage: opencli docker [<user> [<container>]]
------------------------
opencli docker-autostart
Description: Prefetch (pull) container images for services listed as autostart, skipping if free disk space is low.
Usage: opencli docker-autostart [-f|--force]
------------------------
opencli docker-backup
Description: Generates a backup for all users.
Usage: opencli docker-backup
------------------------
opencli docker-collect_stats
Description: Collect resource usage information for user(s).
Usage: opencli docker-collect_stats <username|--all>
------------------------
opencli docker-logs
Description: Display log sizes for user and system containers
Usage: opencli docker-logs [--all|--system|--users|<USERNAME>]
------------------------
opencli docker-update
Description: Check for image updates across root's stack and every
Usage: opencli docker-update [-y|--yes] [--dry-run]
------------------------
opencli domain
Description: View and set domain/ip for accessing panels.
Usage: opencli domain [set <domain_name> | ip] [--debug] [--no-restart]
------------------------
opencli domains-add
Description: Add a domain name for user.
Usage: opencli domains-add <DOMAIN_NAME> <USERNAME> [--docroot DOCUMENT_ROOT] [--php_version N.N] [--skip_caddy] [--skip_vhost] [--skip_containers] [--skip_dns] [--skip-sentinel] [--debug] [--hs_ed25519_public_key KEY --hs_ed25519_secret_key KEY]
------------------------
opencli domains-all
Description: Lists all domain names currently hosted on the server.
Usage: opencli domains-all [--docroot] [--php_version] [--json]
------------------------
opencli domains-delete
Description: Delete a domain name.
Usage: opencli domains-delete <DOMAIN_NAME> [--debug]
------------------------
opencli domains-dns
Description: Manage DNS zones and the DNS server via subcommands (create, delete, reload, list, check, start/stop/restart, etc).
Usage: opencli domains-dns <reconfig|check|reload|show|list|create|delete|default|count|config|start|restart|hard-restart|stop> [DOMAIN] [-y]
------------------------
opencli domains-dnssec
Description: Enable DNSSEC for a domain, re-sign the zone after changes, or check its DNSSEC status.
Usage: opencli domains-dnssec <DOMAIN> [--update | --check]
------------------------
opencli domains-docroot
Description: View and change docroot for a domain.
Usage: opencli domains-docroot <DOMAIN_NAME> [update <DOCROOT>] [--debug]
------------------------
opencli domains-edit
Description: Open a domain's docroot directory, or its webserver vhost config file in nano with --ws.
Usage: opencli domains-edit <DOMAIN_NAME> [--ws]
------------------------
opencli domains-hsts
Description: Manage HSTS for a domain
Usage: opencli domains-hsts <domain> [enable|disable]
------------------------
opencli domains-ssl
Description: Check SSL for domain, add custom certificate, view files.
Usage: opencli domains-ssl <DOMAIN_NAME> [status|info|logs|auto|custom] [path/to/fullchain.pem path/to/key.pem]
------------------------
opencli domains-stats
Description: Parse caddy access logs for users domains and generate static html
Usage: opencli domains-stats [USERNAME] [--debug]
------------------------
opencli domains-suspend
Description: Suspend a domain name
Usage: opencli domains-suspend <DOMAIN-NAME> [--comment="<COMMENT>"]
------------------------
opencli domains-unsuspend
Description: Unsuspend a domain name
Usage: opencli domains-unsuspend <DOMAIN-NAME>
------------------------
opencli domains-update_ns
Description: Change nameservers for a single or all dns zones.
Usage: opencli domains-update_ns <DOMAIN_NAME | --all> [-y]
------------------------
opencli domains-user
Description: Lists all domain names currently owned by a specific user.
Usage: opencli domains-user <USERNAME> [--docroot|--php_version]
------------------------
opencli domains-varnish
Description: Check Varnish status for domain, enable/disable Varnish caching.
Usage: opencli domains-varnish <DOMAIN-NAME> [on|off] [--short]
------------------------
opencli domains-whoowns
Description: Check which username owns a certain domain name.
Usage: opencli domains-whoowns <DOMAIN-NAME> [--context] [--docroot]
------------------------
opencli email-manage
Description: Pass commands through to docker-mailserver's setup CLI inside the mailserver container.
Usage: opencli email-manage <COMMAND> [<ARGS>...]
------------------------
opencli email-quotas
Description: Fixes email permission issues for all domains.
Usage: opencli email-quotas
------------------------
opencli email-ratelimit
Description: Configure rate-limiting using postfwd for domains and users.
Usage: opencli email-ratelimit [--username=<user>] [--domain=<domain>] [--all-users] [--delete-user=<user>] [--delete-domain=<domain>] [--skip-reload]
------------------------
opencli email-server
Description: Manage mailserver
Usage: opencli email-server <status|config|install|uninstall|start|stop|restart|queue|flush|view|unhold|delete|fail2ban|ports|postconf|logs|login|supervisor|postfwd|pflogsumm|update-check|update-packages|versions> [args] [--debug]
------------------------
opencli email-setup
Description: Setup email addresses, forwarders, filters..
Usage: opencli email-setup <COMMAND> <ATTRIBUTES>
------------------------
opencli email-webmail
Description: Display or update the domain used for accessing webmail.
Usage: opencli email-webmail [--debug] [domain <domain>]
------------------------
opencli faq
Description: Display answers to most frequently asked questions.
Usage: opencli faq
------------------------
opencli files-calculate_resellers_storage
Description: Calculates total disk usage for all resellers.
Usage: opencli files-calculate_resellers_storage
------------------------
opencli files-fix_permissions
Description: Fix permissions for users /home directory files inside the container.
Usage: opencli files-fix_permissions <USERNAME|--all> [PATH] [--debug]
------------------------
opencli files-malware_scan
Description: Scan a user's (or every user's) html_data files for malware
Usage: opencli files-malware_scan <USERNAME> OR opencli files-malware_scan --all
------------------------
opencli files-purge_trash
Description: Auto-purge .Trash folders for users.
Usage: opencli files-purge_trash [--user USERNAME] [--force] [--dry-run]
------------------------
opencli ftp-add
Description: Create FTP sub-user for openpanel user.
Usage: opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME> [--debug]
------------------------
opencli ftp-connections
Description: Display all active FTP connection or for particular OpenPanel user.
Usage: opencli ftp-connections [OPENPANEL_USERNAME]
------------------------
opencli ftp-delete
Description: Delete FTP sub-user for openpanel user.
Usage: opencli ftp-delete <username> <openpanel_username> [--debug]
------------------------
opencli ftp-list
Description: List FTP sub-users for openpanel user.
Usage: opencli ftp-list [OPENPANEL_USERNAME] [--json]
------------------------
opencli ftp-logs
Description: View the FTP server log
Usage: opencli ftp-logs
------------------------
opencli ftp-password
Description: Change password for FTP sub-user of openpanel user.
Usage: opencli ftp-password <username> <new_password> <openpanel_username> [--debug]
------------------------
opencli ftp-path
Description: Change FTP path for a user.
Usage: opencli ftp-path <username> <path> <openpanel_username> [--debug]
------------------------
opencli imunify
Description: Install and manage ImunifyAV service.
Usage: opencli imunify [status|start|stop|install|update|uninstall]
------------------------
opencli license
Description: Manage OpenPanel Enterprise license.
Usage: opencli license [key|verify|info|delete|<LICENSE_KEY>] [--json] [--no-restart]
------------------------
opencli php-default
Description: View or change the default PHP version used for new domains added by user.
Usage: opencli php-default <USERNAME> [--update <PHP_VERSION>]
------------------------
opencli php-domain
Description: View or change the PHP version used for a single domain name.
Usage: opencli php-domain <DOMAIN_NAME> [--update <PHP_VERSION>]
------------------------
opencli phpmyadmin
Description: View and set domain/ip for phpmyadmin access.
Usage: opencli phpmyadmin [set <domain_name> | ip]
------------------------
opencli plan-apply
Description: Change plan for a user and apply new plan limits.
Usage: opencli plan-apply <PLAN_ID> <USERNAME>... [--all] [--cpu] [--ram] [--dsk] [--net] [--email] [--debug]
------------------------
opencli plan-create
Description: Create a new hosting plan (Package) and set its limits.
Usage: opencli plan-create name="<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT> max_hourly_email=<COUNT> [reseller=<USERNAME>]
------------------------
opencli plan-delete
Description: Delete hosting plan
Usage: opencli plan-delete <PLAN_NAME> [--json]
------------------------
opencli plan-edit
Description: Edit an existing hosting plan (Package) and modify its parameters.
Usage: opencli plan-edit id=<ID> name="<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT> max_hourly_email=<COUNT> [--debug]
------------------------
opencli plan-list
Description: Display all plans: id, name, description, limits..
Usage: opencli plan-list [--json]
------------------------
opencli plan-usage
Description: Display all users that are currently using the plan.
Usage: opencli plan-usage <PLAN_NAME> [--json]
------------------------
opencli port
Description: View and change port for accessing openpanel.
Usage: opencli port [set <port> | default]
------------------------
opencli proxy
Description: View and change proxy path '/openpanel' for accessing openpanel.
Usage: opencli proxy [set <path>|default]
------------------------
opencli report
Description: Generate a system report and send it to OpenPanel support team.
Usage: opencli report [--public|--link|--upload] [--non-interactive] [--user <USERNAME>]
------------------------
opencli sentinel
Description: Check system services, traffic and resource usage, and log/send custom notifications on request.
Usage: opencli sentinel [--startup] [--report] [--action=<name> --title=<title> --message=<msg>]
------------------------
opencli server-logrotate
Description: Configures logrotate for caddy, openpanel, syslog.
Usage: opencli server-logrotate
------------------------
opencli server-migrate
Description: Migrates all data from this server to another.
Usage: opencli server-migrate -h <DESTINATION_IP> --user root --password <DESTINATION_PASSWORD> [--force] [--exclude-* options]
------------------------
opencli update
Description: Check if update is available, install updates.
Usage: opencli update [--check | --force | --admin | --panel | --cli | --translations | --system | --modules | --compose | --env | --php]
------------------------
opencli user-2fa
Description: Check or disable 2FA for a user.
Usage: opencli user-2fa <username> [disable]
------------------------
opencli user-add
Description: Create a new user with the provided plan_name.
Usage: opencli user-add <USERNAME> <PASSWORD|generate> <EMAIL> "<PLAN_NAME>" [--send-email] [--debug] [--webserver="<nginx|apache|openresty|openlitespeed|litespeed|varnish+nginx|varnish+apache|varnish+openresty|varnish+openlitespeed>"] [--sql=<mysql|mariadb>] [--reseller=<RESELLER_USERNAME>] [--private-note="this user.."] [--no-sentinel]
------------------------
opencli user-backup
Description: Creates a full account .tar.gz backup of a single OpenPanel user account.
Usage: opencli user-backup --account <USER> [--output <DIR>] [--quiet]
------------------------
opencli user-block_ip
Description: View, add, or remove blocked IP addresses for a user's domains.
Usage: opencli user-block_ip <username> [--list='ip_here another_ip' | --delete-all]
------------------------
opencli user-change_plan
Description: Change plan for a user and apply new plan limits.
Usage: opencli user-change_plan <USERNAME> <NEW_PLAN_NAME>
------------------------
opencli user-check
Description: Performs comprehensive security checks on user files, rootless Podman instance and containers.
Usage: opencli user-check <USERNAME>
------------------------
opencli user-delete
Description: Delete user account and permanently remove all their data.
Usage: opencli user-delete <username> [-y]
------------------------
opencli user-email
Description: Change email for user
Usage: opencli user-email <USERNAME> <NEW_EMAIL>
------------------------
opencli user-ip
Description: Assign or remove dedicated IP to a user.
Usage: opencli user-ip <USERNAME> <IP | DELETE> [-y] [--debug]
------------------------
opencli user-list
Description: Display all users: id, username, email, plan, registered date,
Usage: opencli user-list [--json] [--total] [--quota]
------------------------
opencli user-login
Description: Generate an auto-login link for OpenPanel user.
Usage: opencli user-login <username> [--open|--delete]
------------------------
opencli user-loginlog
Description: View user's .lastlogin file with last 20 successfull logins.
Usage: opencli user-loginlog <USERNAME> [--table|--text|--json]
------------------------
opencli user-password
Description: Reset password for a user.
Usage: opencli user-password <USERNAME> <NEW_PASSWORD | random>
------------------------
opencli user-quota
Description: Report or set disk and inodes for users.
Usage: opencli user-quota <username|--all>
------------------------
opencli user-rename
Description: Rename username.
Usage: opencli user-rename <old_username> <new_username>
------------------------
opencli user-restore
Description: Restores a single OpenPanel user account from a full account .tar.gz backup.
Usage: opencli user-restore --file <ARCHIVE> [--force] [--new-username=NAME] [--quiet] [--temp-dir=<PATH> ]
------------------------
opencli user-suspend
Description: Suspend user: stop all containers and suspend domains.
Usage: opencli user-suspend <USERNAME> [-y] [--debug]
------------------------
opencli user-transfer
Description: Transfers a single user account from this server to another.
Usage: opencli user-transfer --account <OPENPANEL_USER> --host <DESTINATION_IP> --username <DESTINATION_SSH_USERNAME> --password <DESTINATION_SSH_PASSWORD> [--port <SSH_PORT>] [--force] [--live-transfer]
------------------------
opencli user-unsuspend
Description: Unsuspend user: start all containers and unsuspend domains
Usage: opencli user-unsuspend <USERNAME>
------------------------
opencli user-varnish
Description: Enable/disable Varnish Caching for user and display current status.
Usage: opencli user-varnish <USERNAME> [enable|disable|status]
------------------------
opencli version
Description: Displays the current (installed) version of OpenPanel container image.
Usage: opencli version
------------------------
opencli waf
Description: Manage CorazaWAF
Usage: opencli waf <status|enable|disable|domain|tags|ids|update|stats|count> [options]
------------------------
opencli websites-all
Description: Lists all websites currently hosted on the server.
Usage: opencli websites-all [TYPE]
------------------------
opencli websites-pagespeed
Description: Check Google PageSpeed data for website(s)
Usage: opencli websites-pagespeed <DOMAIN | -all>
------------------------
opencli websites-scan
Description: Scan user files for WP sites and add them to SiteManager interface.
Usage: opencli websites-scan <USERNAME|-all>
------------------------
opencli websites-secure
Description: WP Manager security rules for domain.
Usage: opencli websites-secure <DOMAIN> [--rules='RULE1 RULE2' | --disable-all | --list-active-rules]
------------------------
opencli websites-user
Description: Lists all websites and domains owned by a specific user.
Usage: opencli websites-user <USERNAME> [--type=] [--domains=] [--json]
------------------------
opencli websites-vulnerability
Description: Scan WP site for vulnerabilities.
Usage: opencli websites-vulnerability <DOMAIN | --all>
------------------------
opencli error
Description: Displays information for specific error ID received in OpenPanel UI.
Usage: opencli error <ID_HERE>
------------------------
opencli locale
Description: Install locales (Languages) for OpenPanel UI.
Usage: opencli locale <CODE>
------------------------
```
</details>
