import React from "react";
import clsx from "clsx";

type Props = {
    className?: string;
};

const OPENCLI_COMMANDS: { cmd: string; desc: string; usage: string }[] = [
    { cmd: `opencli php-default`, desc: `View or change the default PHP version used for new domains added by user.`, usage: `opencli php-default <username>` },
    { cmd: `opencli php-domain`, desc: `View or change the PHP version used for a single domain name.`, usage: `opencli php-domain <domain_name>` },
    { cmd: `opencli files-fix_permissions`, desc: `Fix permissions for users /home directory files inside the container.`, usage: `opencli files-fix_permissions <USERNAME> [PATH]` },
    { cmd: `opencli files-purge_trash`, desc: `Auto-purge .Trash folders for users.`, usage: `opencli files-purge_trash --user [USERNAME]` },
    { cmd: `opencli files-calculate_resellers_storage`, desc: `Calculates total disk usage for all resellers.`, usage: `opencli files-calculate_resellers_storage` },
    { cmd: `opencli proxy`, desc: `View and change proxy path '/openpanel' for accessing openpanel.`, usage: `opencli port [set <path>]` },
    { cmd: `opencli update`, desc: `Check if update is available, install updates.`, usage: `opencli update [--check | --force | --admin | --panel | --cli]` },
    { cmd: `opencli domains-all`, desc: `Lists all domain names currently hosted on the server.`, usage: `opencli domains-all [--docroot|--php_version]` },
    { cmd: `opencli domains-delete`, desc: `Delete a domain name.`, usage: `opencli domains-delete <DOMAIN_NAME> --debug` },
    { cmd: `opencli domains-docroot`, desc: `View and change docroot for a domain.`, usage: `opencli domains-docroot <DOMAIN_NAME> [update </var/www/html/>] --debug` },
    { cmd: `opencli domains-unsuspend`, desc: `Unsuspend a domain name`, usage: `opencli domains-unsuspend <DOMAIN-NAME>` },
    { cmd: `opencli domains-add`, desc: `Add a domain name for user.`, usage: `opencli domains-add <DOMAIN_NAME> <USERNAME> [--docroot DOCUMENT_ROOT] [--php_version N.N] [--skip_caddy --skip_vhost --skip_containers --skip_dns] --debug` },
    { cmd: `opencli domains-dns`, desc: `Manage DNS for a domain.`, usage: `opencli domains-dns <DOMAIN>` },
    { cmd: `opencli domains-edit`, desc: `Enter docroot for a domain.`, usage: `opencli domains-edit <DOMAIN_NAME>` },
    { cmd: `opencli domains-ssl`, desc: `Check SSL for domain, add custom certificate, view files.`, usage: `opencli domains-ssl <DOMAIN_NAME> [status|info|logs|auto|custom] [path/to/fullchain.pem path/to/key.pem]` },
    { cmd: `opencli domains-suspend`, desc: `Suspend a domain name`, usage: `opencli domains-suspend <DOMAIN-NAME> [--comment="<COMMENT>"]` },
    { cmd: `opencli domains-dnssec`, desc: `Enable DNSSEC for a domain and re-sign after changes in the zone.`, usage: `opencli domains-dnssec <DOMAIN> [--update | --check]` },
    { cmd: `opencli domains-hsts`, desc: `Manage HSTS for a domain`, usage: `opencli hsts <domain> [on|off]` },
    { cmd: `opencli domains-user`, desc: `Lists all domain names currently owned by a specific user.`, usage: `opencli domains-user <USERNAME> [--docroot|--php_version]` },
    { cmd: `opencli domains-stats`, desc: `Parse caddy access logs for users domains and generate static html`, usage: `opencli domains-stats` },
    { cmd: `opencli domains-update_ns`, desc: `Change nameservers for a single or all dns zones.`, usage: `opencli domains-update_ns <DOMAIN_NAME>` },
    { cmd: `opencli domains-varnish`, desc: `Check Varnish status for domain, enable/disable Varnish caching.`, usage: `opencli domains-varnish <DOMAIN-NAME> [on|off] [--short]` },
    { cmd: `opencli domains-whoowns`, desc: `Check which username owns a certain domain name.`, usage: `opencli domains-whoowns <DOMAIN-NAME> [--context] [--docroot]` },
    { cmd: `opencli imunify`, desc: `Install and manage ImunifyAV service.`, usage: `opencli imunify [status|start|stop|install|update|uninstall]` },
    { cmd: `opencli report`, desc: `Generate a system report and send it to OpenPanel support team.`, usage: `opencli report` },
    { cmd: `opencli admin`, desc: `Manage OpenAdmin service and Administrators.`, usage: `opencli admin <command> [options]` },
    { cmd: `opencli waf`, desc: `Manage CorazaWAF`, usage: `opencli waf <setting>` },
    { cmd: `opencli user-change_plan`, desc: `Change plan for a user and apply new plan limits.`, usage: `opencli user-change_plan <USERNAME> <NEW_PLAN_NAME>` },
    { cmd: `opencli user-delete`, desc: `Delete user account and permanently remove all their data.`, usage: `opencli user-delete <username> [-y]` },
    { cmd: `opencli user-unsuspend`, desc: `Unsuspend user: start all containers and unsuspend domains`, usage: `opencli user-unsuspend <USERNAME>` },
    { cmd: `opencli user-email`, desc: `Change email for user`, usage: `opencli user-email <USERNAME> <NEW_EMAIL>` },
    { cmd: `opencli user-add`, desc: `Create a new user with the provided plan_name.`, usage: `opencli user-add <USERNAME> <PASSWORD|generate> <EMAIL> "<PLAN_NAME>" [--send-email] [--debug]  [--webserver="<nginx|apache|openresty|openlitespeed|litespeed|varnish+nginx|varnish+apache|varnish+openresty|varnish+openlitespeed>"] [--sql=<mysql|mariadb>] [--RESELLER=<RESELLER_USERNAME>][--server=<IP_ADDRESS>]  [--key=<SSH_KEY_PATH>] [--private-note="this user.."] [--no-sentinel]` },
    { cmd: `opencli user-quota`, desc: `Report or set disk and inodes for users.`, usage: `opencli user-quota <username|--all>` },
    { cmd: `opencli user-2fa`, desc: `Check or disable 2FA for a user.`, usage: `opencli user-2fa <username> [disable]` },
    { cmd: `opencli user-suspend`, desc: `Suspend user: stop all containers and suspend domains.`, usage: `opencli user-suspend <USERNAME>` },
    { cmd: `opencli user-backup`, desc: `Creates a full account .tar.gz backup of a single OpenPanel user account.`, usage: `opencli user-backup --account <USER> [--output <DIR>] [--quiet]` },
    { cmd: `opencli user-transfer`, desc: `Transfers a single user account from this server to another.`, usage: `opencli user-transfer --account <OPENPANEL_USER> --host <DESTINATION_IP> --username <DESTINATION_SSH_USERNAME> --password <DESTINATION_SSH_PASSWORD> [--live-transfer]` },
    { cmd: `opencli user-list`, desc: `Display all users: id, username, email, plan, registered date.`, usage: `opencli user-list [--json]` },
    { cmd: `opencli user-ip`, desc: `Assing or remove dedicated IP to a user.`, usage: `opencli user-ip <USERNAME> <IP | DELETE> [-y] [--debug]` },
    { cmd: `opencli user-restore`, desc: `Restores a single OpenPanel user account from a full account .tar.gz backup.`, usage: `opencli user-restore --file <ARCHIVE> [--force] [--new-username=NAME] [--quiet] [--temp-dir=<PATH> ]` },
    { cmd: `opencli user-rename`, desc: `Rename username.`, usage: `opencli user-rename <old_username> <new_username>` },
    { cmd: `opencli user-password`, desc: `Reset password for a user.`, usage: `opencli user-password <USERNAME> <NEW_PASSWORD | random>` },
    { cmd: `opencli user-block_ip`, desc: `Block IP addresses from accessng user domains.`, usage: `opencli user-block_ip <username> [--list='ip_here another_ip' | --delete-all]` },
    { cmd: `opencli user-varnish`, desc: `Enable/disable Varnish Caching for user and display current status.`, usage: `opencli user-varnish <USERNAME> [on|off]` },
    { cmd: `opencli user-loginlog`, desc: `View user's .lastlogin file with last 20 successfull logins.`, usage: `opencli user-loginlog <USERNAME> [--table|--text|--json]` },
    { cmd: `opencli user-login`, desc: `Generate an auto-login link for OpenPanel user.`, usage: `opencli user-login <username> [--open|--delete]` },
    { cmd: `opencli user-check`, desc: `Performs comprehensive security checks on user files, Docker daemon and containers.`, usage: `opencli user-check <USERNAME>` },
    { cmd: `opencli license`, desc: `Manage OpenPanel Enterprise license.`, usage: `opencli license verify` },
    { cmd: `opencli docker`, desc: `Manage OpenPanel system or user containers with lazydocker.`, usage: `opencli docker [<user> [<container>]]` },
    { cmd: `opencli faq`, desc: `Display answers to most frequently asked questions.`, usage: `opencli faq` },
    { cmd: `opencli websites-vulnerability`, desc: `Scan WP site for vulnerabilities.`, usage: `opencli websites-vulnerability <DOMAIN> [--all]` },
    { cmd: `opencli websites-all`, desc: `Lists all websites currently hosted on the server.`, usage: `opencli websites-all` },
    { cmd: `opencli websites-secure`, desc: `WP Manager security rules for domain.`, usage: `opencli websites-secure <DOMAIN> [--rules='RULE1 RULE2' | --disable-all | --list-active-rules]` },
    { cmd: `opencli websites-pagespeed`, desc: `Check Google PageSpeed data for website(s)`, usage: `opencli websites-pagespeed <DOMAIN> [-all]` },
    { cmd: `opencli websites-scan`, desc: `Scan user files for WP sites and add them to SiteManager interface.`, usage: `opencli websites-scan $username` },
    { cmd: `opencli websites-user`, desc: `Lists all websites and domains owned by a specific user.`, usage: `opencli websites-user <USERNAME> [--type=] [--domains=] [--json]` },
    { cmd: `opencli port`, desc: `View and change port for accessing openpanel.`, usage: `opencli port [set <port>]` },
    { cmd: `opencli ftp-delete`, desc: `Delete FTP sub-user for openpanel user.`, usage: `opencli ftp-delete <username> <openpanel_username> [--debug]` },
    { cmd: `opencli ftp-add`, desc: `Create FTP sub-user for openpanel user.`, usage: `opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME>` },
    { cmd: `opencli ftp-path`, desc: `Change FTP path for a user.`, usage: `opencli ftp-path <username> <path> <openpanel_username> [--debug]` },
    { cmd: `opencli ftp-logs`, desc: `View the FTP server log`, usage: `opencli ftp-log` },
    { cmd: `opencli ftp-connections`, desc: `Display all active FTP connection or for particular OpenPanel user.`, usage: `opencli ftp-add <NEW_USERNAME> <NEW_PASSWORD> <FOLDER> <OPENPANEL_USERNAME>` },
    { cmd: `opencli ftp-list`, desc: `List FTP sub-users for openpanel user.`, usage: `opencli ftp-list <OPENPANEL_USERNAME> [--json]` },
    { cmd: `opencli ftp-password`, desc: `Change password for FTP sub-user of openpanel user.`, usage: `opencli ftp-password <username> <new_password> <openpanel_username> [--debug]` },
    { cmd: `opencli version`, desc: `Displays the current (installed) version of OpenPanel docker image.`, usage: `opencli version` },
    { cmd: `opencli domain`, desc: `View and set domain/ip for accessing panels.`, usage: `opencli domain [set <domain_name> | ip] [--debug]` },
    { cmd: `opencli docker-autostart`, desc: `Set services to auto-start for user on acocunt creation.`, usage: `opencli docker-autostart` },
    { cmd: `opencli docker-collect_stats`, desc: `Collect resource usage information for user(s).`, usage: `opencli docker-collect_stats` },
    { cmd: `opencli docker-logs`, desc: `Display log sizes for user and sytem containers`, usage: `opencli docker-logs [--all|system|<USERNAME>]` },
    { cmd: `opencli docker-backup`, desc: `Generates a backup for all users.`, usage: `opencli docker-backup` },
    { cmd: `opencli docker-images`, desc: `Check and auto-update Docker images per user.`, usage: `opencli docker-images [--all|<USERNAME>] [--dry-run] [--force-update]` },
    { cmd: `opencli docker-limits`, desc: `Set global docker limits for all containers combined.`, usage: `opencli docker-limits [--apply | --apply SIZE | --read]` },
    { cmd: `opencli server-logrotate`, desc: `Configures logrotate for caddy, openpanel, syslog.`, usage: `opencli server-logrotate` },
    { cmd: `opencli server-migrate`, desc: `Migrates all data from this server to another.`, usage: `opencli server-migrate -h <DESTINATION_IP> --user root --password <DESTINATION_PASSWORD>` },
    { cmd: `opencli patch`, desc: `Download and install a patch`, usage: `opencli patch <NAME>` },
    { cmd: `opencli sentinel`, desc: `Check system services, traffic and resource usage.`, usage: `opencli sentinel` },
    { cmd: `opencli email-ratelimit`, desc: `Configure rate-limiting using postfwd for domains and users.`, usage: `opencli email-ratelimit` },
    { cmd: `opencli email-server`, desc: `Manage mailserver`, usage: `opencli email-server <install|start|restart|stop|uninstall> [--debug]` },
    { cmd: `opencli email-setup`, desc: `Setup email addresses, forwarders, filters..`, usage: `opencli email-setup <COMMAND> <ATTRIBUTES>` },
    { cmd: `opencli email-webmail`, desc: `Display webmail domain or choose webmail software`, usage: `opencli email-webmail [--debug]` },
    { cmd: `opencli email-manage`, desc: `Manage mailserver configuration and overview.`, usage: `opencli email-manage <COMMAND> <ATTRIBUTES>` },
    { cmd: `opencli email-quotas`, desc: `Fixes email permission issues for all domains.`, usage: `opencli email-quotas` },
    { cmd: `opencli api`, desc: `On/Off OpenAdmin API access and list available API endpoints.`, usage: `opencli api <status|on|off|list>` },
    { cmd: `opencli plan-delete`, desc: `Delete hosting plan`, usage: `opencli plan-delete <PLAN_NAME>` },
    { cmd: `opencli plan-apply`, desc: `Change plan for a user and apply new plan limits.`, usage: `opencli plan-apply <NEW_PLAN_ID> <USERNAME>` },
    { cmd: `opencli plan-edit`, desc: `Edit an existing hosting plan (Package) and modify its parameters.`, usage: `opencli plan-edit --debug id=<ID> name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<DEFAULT> max_email_quota=<COUNT> max_hourly_email=<COUNT>` },
    { cmd: `opencli plan-list`, desc: `Display all plans: id, name, description, limits..`, usage: `opencli plan-list [--json]` },
    { cmd: `opencli plan-create`, desc: `Create a new hosting plan (Package) and set its limits.`, usage: `opencli plan-create name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT> max_hourly_email=<COUNT>` },
    { cmd: `opencli plan-usage`, desc: `Display all users that are currently using the plan.`, usage: `opencli plan-usage <PLAN_NAME> [--json]` },
    { cmd: `opencli config`, desc: `View / change configuration for users and set defaults for new accounts.`, usage: `opencli config get <setting_name>` },
    { cmd: `opencli phpmyadmin`, desc: `View and set domain/ip for phpmyadmin access.`, usage: `opencli phpmyadmin [set <domain_name> | ip]` },
    { cmd: `opencli error`, desc: `Displays information for specific error ID received in OpenPanel UI.`, usage: `opencli error <ID_HERE>` },
    { cmd: `opencli locale`, desc: `Install locales (Languages) for OpenPanel UI.`, usage: `opencli locale <CODE>` },
];

const CommandBlock = ({
    cmd,
    desc,
    usage,
}: {
    cmd: string;
    desc: string;
    usage: string;
}) => (
    <div className={clsx("px-6", "py-3")}>
        <div className={clsx("flex", "items-baseline", "gap-2")}>
            <span className={clsx("text-refine-cyan-alt")}>$</span>
            <span className={clsx("text-gray-0", "font-semibold")}>{cmd}</span>
        </div>
        <div className={clsx("pl-4", "text-gray-500")}>
            <span>Description: </span>
            <span className={clsx("text-gray-400")}>{desc}</span>
        </div>
        <div className={clsx("pl-4", "text-gray-500")}>
            <span>Usage: </span>
            <span className={clsx("text-gray-400")}>{usage}</span>
        </div>
    </div>
);

const CommandList = () => (
    <div>
        {OPENCLI_COMMANDS.map((command) => (
            <CommandBlock key={command.cmd} {...command} />
        ))}
    </div>
);

export const LandingHeroShowcaseOpenCliTerminal = ({ className }: Props) => {
    return (
        <div
            className={clsx(
                className,
                "h-full",
                "flex",
                "flex-col",
                "bg-gray-900",
                "overflow-hidden",
            )}
        >
            <div
                className={clsx(
                    "shrink-0",
                    "h-10",
                    "flex",
                    "items-center",
                    "gap-1.5",
                    "px-4",
                    "bg-[#1B1B29]",
                )}
            >
                <span
                    className={clsx("w-2.5 h-2.5 rounded-full")}
                    style={{ background: "#FF5F57" }}
                />
                <span
                    className={clsx("w-2.5 h-2.5 rounded-full")}
                    style={{ background: "#FEBC2E" }}
                />
                <span
                    className={clsx("w-2.5 h-2.5 rounded-full")}
                    style={{ background: "#28C840" }}
                />
                <span
                    className={clsx(
                        "flex-1",
                        "text-center",
                        "text-gray-400",
                        "text-xs",
                        "font-jetBrains-mono",
                        "pr-8",
                    )}
                >
                    root@server: /opencli
                </span>
            </div>
            <div
                className={clsx(
                    "flex-1",
                    "overflow-hidden",
                    "relative",
                    "landing-opencli-terminal-mask",
                )}
            >
                <div
                    className={clsx(
                        "animate-opencli-scroll",
                        "will-change-transform",
                        "text-[11px] landing-sm:text-xs",
                        "leading-[17px]",
                        "font-jetBrains-mono",
                        "select-none",
                    )}
                >
                    <CommandList />
                    <CommandList />
                </div>
            </div>
        </div>
    );
};
