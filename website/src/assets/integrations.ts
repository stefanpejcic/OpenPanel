import { IntegrationsType } from "../types/integrations";
import {
    Search,
    Languages,
    Keyboard,
    Dark,
    ServerInfo,
    Responsive,
    Branding,
    Suspend,
    Upgrade,
    UserServer,
    UserIP,
    UserLogin,
    Firewall,
    SSL,
    ModSec,
    Factor,
    CPU,
    Box,
    IPblock,
    Terminal,
    AdminOff,
    Ports,
    Separate,
    WHMCS,
    API,
    Notifications,
    Services,
    Usage,
    Bolt,
    Configuration,
    Download,
    Domain,
    Folder,
    PHP,
    Python,
    WP,
    MySQL,
    PostgreSql,
    MongoDB,
    ServerCog,
    CronJobs,
    Cache,
    DNS,
    Visitors,
    Activity,
    AccountCog,
    Apache,
    Nginx,
    Mail,
    Backup,
    Resellers,
    Migration,
    FeatureManager,
    DynamicDns,
    DnsCluster,
} from "./integration-icons";

export const integrations: IntegrationsType = {
    "ui-framework-packages": [
        {
            name: "Apache",
            icon: Apache,
            description:
                "Run Apache web server per user. Apache's support for .htaccess files enables users to customize and override global configuration settings on a per-directory basis.",
            url: "/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/#apache",
            status: "stable",
        },
        {
            name: "Nginx",
            icon: Nginx,
            description:
                "Nginx per user, with FastCGI cache, cache purging, cache exclusions and rewrite rules set per domain.",
            url: "/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/#nginx",
            status: "stable",
        },
        {
            name: "OpenResty",
            icon: Nginx,
            description:
                "OpenResty allows per-user customization by defining user-specific location blocks or dynamic Lua logic in the main config. Configuration is centralized and preloaded for performance ",
            url: "/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/#openresty",
            status: "stable",
        },
        {
            name: "LiteSpeed",
            icon: Apache,
            description:
                "LiteSpeed Enterprise reads Apache configuration and .htaccess files, runs PHP in its own LSPHP and caches pages with LSCache. Each account starts on a trial license and the user adds their serial number from OpenPanel.",
            url: "/docs/panel/webserver/webserver_type/#litespeed",
            status: "enterprise",
        },
        {
            name: "OpenLiteSpeed",
            icon: Apache,
            description:
                "OpenLiteSpeed reads .htaccess and has its own page cache. Each user sets their own security, cache and rewrite rules.",
            url: "/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/#openlitespeed",
            status: "enterprise",
        },
        {
            name: "Varnish",
            icon: Nginx,
            description:
                "Varnish in front of the user's web server, with WordPress cache rules and purging already set up.",
            url: "/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/",
            status: "enterprise",
         },
    ],
    "data-provider-packages": [
        {
            name: "Domain Names",
            icon: Domain,
            description:
                "Add domain names with automatic SSL renewals, include aliases and subdomains (Internationalized domains are supported), create redirects, enforce HTTPS, and edit vhost files.",
            url: "/docs/panel/domains/",
            status: "stable",
        },
        {
            name: "File Manager",
            icon: Folder,
            description:
                "Upload many files at once with no size limit, edit files in the browser and change permissions, copy or move files.",
            url: "/docs/panel/files/",
            status: "stable",
        },
        {
            name: "Email Hosting",
            icon: Mail,
            description:
                "Create and manage email accounts, aliases and forwarders, spam filters, and webmail access, with tools to monitor and improve deliverability.",
            url: "/docs/panel/emails/",
            status: "enterprise",
        },
        {
            name: "Backups",
            icon: Backup,
            description:
                "Users configure their own backup schedules, choose exactly what to back up, and restore from a wide range of supported destinations.",
            url: "/docs/panel/backups/",
            status: "enterprise",
        },
        {
            name: "FTP",
            icon: Folder,
            description:
                "Create FTP sub-users that share the FTP service with other users on the server, and connect using the server IP address and port 21.",
            url: "/docs/panel/files/ftp/",
            status: "enterprise",
        },
        {
            name: "PHP versions",
            icon: PHP,
            description:
                "Users can use different PHP versions for each domain, run multiple versions, set a default version for new domains, change limits by editing php.ini files.",
            url: "/docs/panel/php/domains/",
            status: "stable",
        },
        {
            name: "NodeJS, Python, Ruby & Java",
            icon: Python,
            description:
                "Run Node.js, Python, Ruby and Java apps in their own containers and point a domain at them.",
            url: "/docs/panel/applications/",
            status: "enterprise",
        },
        {
            name: "WP Manager",
            icon: WP,
            description: "Install WordPress, log in to wp-admin without a password, edit options, turn on debug mode and make a backup when needed.",
            url: "/docs/panel/applications/wordpress/",
            status: "stable",
        },
        {
            name: "One-Click App Installer",
            icon: Download,
            description:
                "Install Drupal, Joomla, Moodle, NextCloud, Matomo, OpenCart, PrestaShop, MediaWiki, DokuWiki, phpBB and more with a single click using the Autoinstaller.",
            url: "/docs/panel/applications/autoinstaller/",
            status: "stable",
        },
        {
            name: "Disk & Inodes Usage",
            icon: Usage,
            description:
                "Browse current disk and inodes usage per directory to quickly identify what is taking up space on the account.",
            url: "/docs/panel/statistics/disk_usage/",
            status: "enterprise",
        },
        {
            name: "MySQL and phpMyAdmin",
            icon: MySQL,
            description:
                "Create MySQL databases and users, open phpMyAdmin without logging in, allow remote access and edit the server's limits.",
            url: "/docs/panel/mysql/databases/",
            status: "stable",
        },
        {
            name: "PostgreSQL",
            icon: PostgreSql,
            description:
                "Create PostgreSQL databases and users, allow remote access and edit the server's limits.",
            url: "",
            status: "enterprise",
        },
        {
            name: "MongoDB",
            icon: MongoDB,
            description:
                "Create MongoDB databases and users, assign roles per database and import mongodump archives.",
            url: "/docs/panel/mongodb/databases/",
            status: "enterprise",
        },
        {
            name: "SSL certificates",
            icon: SSL,
            description:
                "Let's Encrypt certificates are issued and renewed automatically. Users can also upload their own certificate.",
            url: "/docs/panel/domains/SSL/",
            status: "stable",
        },
        {
            name: "Object Caching",
            icon: Cache,
            description:
                "Each user can run their own Redis, Valkey or Memcached, set its memory limit, restart it and read its logs.",
            url: "/docs/panel/caching/",
            status: "stable",
        },
        {
            name: "Web Terminal",
            icon: Terminal,
            description:
                "A terminal in the browser for running commands inside the user's containers.",
            url: "/docs/panel/containers/terminal/",
            status: "enterprise",
        },
        {
            name: "Server settings",
            icon: Configuration,
            description:
                "Users can start, stop and restart their own services and edit their configuration files.",
            url: "/docs/panel/webserver/webserver_settings/",
            status: "stable",
        },
        {
            name: "Cron Jobs",
            icon: CronJobs,
            description:
                "Add and edit cron jobs from the panel.",
            url: "/docs/panel/cronjobs/",
            status: "stable",
        },
        {
            name: "DNS Zone Editor",
            icon: DNS,
            description:
                "Edit DNS zones and add A, AAAA, CNAME, MX, TXT and other records.",
            url: "/docs/panel/domains/dns/",
            status: "stable",
        },
        {
            name: "Dynamic DNS",
            icon: DynamicDns,
            description:
                "Keep DNS records automatically in sync with changing IP addresses using a secure token URL, ideal for home servers and routers.",
            url: "/docs/panel/domains/dynamic-dns/",
            status: "enterprise",
        },
        {
            name: "Resource usage",
            icon: Usage,
            description:
                "Current CPU and memory usage per account, plus history, so users can see what slowed their site down.",
            url: "/docs/panel/statistics/resource_usage/",
            status: "stable",
        },
        {
            name: "Visitor reports",
            icon: Visitors,
            description:
                "GoAccess reports built from the access logs: visitors by country, top pages, IPs and error pages.",
            url: "/docs/panel/statistics/goaccess/",
            status: "stable",
        },
        {
            name: "Activity logs",
            icon: Activity,
            description:
                "Every action in the panel is logged, so users can see who changed what.",
            url: "/docs/panel/security/account_activity/",
            status: "stable",
        },
        {
            name: "Account settings",
            icon: AccountCog,
            description:
                "Users can change their email and password, turn on 2FA, pick a language and switch to dark mode.",
            url: "/docs/panel/account/",
            status: "stable",
        },
        {
            name: "Passkeys",
            icon: Factor,
            description:
                "Users can register passkeys for their account and sign in without a password, using their device's biometrics or security key.",
            url: "/docs/panel/security/passkeys/",
            status: "enterprise",
        },
        {
            name: "Login History",
            icon: UserLogin,
            description:
                "Users can view the last 20 IP addresses that logged into their account to spot any suspicious activity.",
            url: "/docs/panel/security/login_history/",
            status: "stable",
        },
    ],
    "community-data-provider-packages": [
        {
            name: "Multi-Server",
            icon: ServerCog,
            description:
                "Manage several OpenPanel servers from one OpenAdmin: see users, domains and services of all servers, apply settings everywhere and move accounts between servers.",
            url: "/docs/admin/system/multi-server/",
            status: "enterprise",
        },
        {
            name: "Simple server deployment",
            icon: Download,
            description:
                "Install OpenPanel in minutes. Add new servers to your multi-server cluster right from the install command.",
            url: "/docs/admin/intro/#installation",
            status: "stable",
        },
        {
            name: "cPanel Migration",
            icon: Migration,
            description:
                "Create an OpenPanel account directly from a full cPanel account backup, making it easy to move existing customers over.",
            url: "/docs/admin/backups/cpanel/",
            status: "enterprise",
        },
        {
            name: "Feature Manager",
            icon: FeatureManager,
            description:
                "Enable or disable specific features across the OpenPanel UI to customize the control panel experience per user role or hosting plan.",
            url: "/docs/admin/plans/feature-manager/",
            status: "enterprise",
        },
        {
            name: "DNS Clustering",
            icon: DnsCluster,
            description:
                "Synchronize DNS records from one OpenPanel server to other machines running BIND9, keeping DNS in sync across a cluster.",
            url: "/docs/admin/domains/dns-cluster/",
            status: "enterprise",
        },
        {
            name: "Resource usage management",
            icon: CPU,
            description: "Limit the CPU, I/O bandwidth, IOPS, nproc and memory on a per-user basis to ensure consistent performance for all your hosted websites.",
            url: "/",
            status: "stable",
        },
        {
            name: "Switch web servers",
            icon: Bolt,
            description:
                "With the Docker module, users have an option to switch to their preferred webserver: Nginx, Apache, LiteSpeed, OpenLiteSpeed, OpenResty.",
            url: "/docs/panel/webserver/webserver_type/",
            status: "enterprise",
        },
        {
            name: "Edit configuration",
            icon: Configuration,
            description:
                "Set the domain and port for OpenPanel and OpenAdmin and change other panel settings.",
            url: "/docs/category/settings/",
            status: "enterprise",
        },
        {
            name: "Resource usage statistics",
            icon: Usage,
            description:
                "Monitor CPU, Memory usage, Network, Disk and load with real time monitoring.",
            url: "/docs/admin/dashboard/#usage",
            status: "stable",
        },
        {
            name: "Service management",
            icon: Services,
            description:
                "See which services are running, restart them and read their logs from OpenAdmin.",
            url: "/docs/admin/services/status/",
            status: "stable",
        },
        {
            name: "Smart Notifications",
            icon: Notifications,
            description:
                "Get notified about reboots, high load, attacks on websites, failed services and similar events.",
            url: "/docs/admin/notifications/",
            status: "enterprise",
        },
        {
            name: "Webhooks",
            icon: Notifications,
            description:
                "Administrators can set up webhooks for Discord, Slack and other services, or receive email alerts instead.",
            url: "/docs/admin/notifications/",
            status: "enterprise",
        },
        {
            name: "Terminal Commands",
            icon: Terminal,
            description:
                "More than 100 opencli commands for administrators, useful for scripts and automation.",
            url: "https://dev.openpanel.com/cli/commands.html",
            status: "stable",
        },
    ],
    frameworks: [
        {
            name: "Branding",
            icon: Branding,
            description: "Fully customise the OpenPanel with colours, logos, fonts and more that mirror the look and feel of your hosting company.",
            url: "/docs/admin/settings/openpanel/#branding",
            status: "enterprise",
        },
        {
            name: "Fully responsive",
            icon: Responsive,
            description:
                "Access OpenPanel on desktop, tablet and mobile without any feature limitations.",
            url: "",
            status: "stable",
        },
        {
            name: "Dark Mode",
            icon: Dark,
            description:
                "Users can switch the panel to dark mode with one click.",
            url: "/docs/panel/dashboard/dark-mode/",
            status: "stable",
        },
        {
            name: "Server Info",
            icon: ServerInfo,
            description:
                "View real-time usage data, IP address, nameservers, and other important server information directly within the OpenPanel interface.",
            url: "/docs/panel/server_info/",
            status: "stable",
        },
        {
            name: "Advanced Search",
            icon: Search,
            description:
                "Search across pages and settings in the panel.",
            url: "/docs/panel/dashboard",
            status: "enterprise",
        },
        {
            name: "Multi-language support",
            icon: Languages,
            description:
                "OpenPanel is translation ready. Each login can view OpenPanel in their preferred language.",
            url: "https://dev.openpanel.com/localization.html",
            status: "stable",
        },
        {
            name: "Keyboard Shortcuts",
            icon: Keyboard,
            description:
                "More than 20 keyboard shortcuts for moving around the panel without a mouse.",
            url: "/docs/articles/dev-experience/openadmin-keyboard-shortcuts/",
            status: "stable",
        },
    ],
    integrations: [
        {
            name: "Suspend / Delete accounts",
            icon: Suspend,
            description:
                "Suspend customer accounts to instantly disable their OpenPanel access and websites. Delete accounts when they are no longer required.",
            url: "/docs/admin/accounts/users/#suspend",
            status: "stable",
        },
        {
            name: "Upgrade / Downgrade package",
            icon: Upgrade,
            description:
                "Move a customer to a bigger or smaller hosting plan.",
            url: "/docs/admin/accounts/users/#statistics",
            status: "stable",
        },
        {
            name: "Reseller accounts",
            icon: Resellers,
            description:
                "Create Reseller users that act as sub-administrators, managing their own set of user accounts within limits defined by the root administrator.",
            url: "/docs/admin/accounts/resellers/",
            status: "enterprise",
        },
        {
            name: "Dedicated IP address",
            icon: UserIP,
            description:
                "Allocate an IPv4 address to users, providing them with a dedicated IP for their websites and services.",
            url: "/docs/admin/accounts/users/#edit",
            status: "enterprise",
        },
        {
            name: "Impersonation",
            icon: UserLogin,
            description: "Auto-login to access a customer's account and see exactly what they see without having to leave your account.",
            url: "/docs/admin/accounts/users/#statistics",
            status: "stable",
        },
        {
            name: "Webserver per user",
            icon: UserServer,
            description:
                "Pick Apache, Nginx, OpenResty, LiteSpeed or OpenLiteSpeed as the default for new users. Users on the same server can run different web servers.",
            url: "/docs/articles/containers/how-to-set-nginx-apache-varnish-per-user-in-openpanel/",
            status: "stable",
        },
        {
            name: "Database server per user",
            icon: UserServer,
            description:
                "Pick MySQL, Percona or MariaDB for each user. Some users can run MySQL and others MariaDB on the same server.",
            url: "/docs/articles/containers/how-to-set-mysql-mariadb-per-user-in-openpanel/",
            status: "stable",
        },
    ],
    "live-providers": [
        {
            name: "Containerized services",
            icon: Box,
            description:
                "Every user service is containerised. Containers have no access to other users or server files.",
            url: "/docs/panel/intro/",
            status: "stable",
        },
        {
            name: "Resource limiting",
            icon: CPU,
            description:
                "Administrators have the ability to set specific limits per plan, including port speed, disk usage, inodes, the number of websites, MySQL databases, and domains.",
            url: "/docs/admin/plans/hosting_plans/#create-a-plan",
            status: "stable",
        },
        {
            name: "SSL / TLS",
            icon: SSL,
            description:
                "Automatically provision and renewal of Let's Encrypt certificates. Custom SSL certificates are also supported for domains that need them.",
            url: "/docs/panel/domains/ssl/",
            status: "stable",
        },
        {
            name: "CorazaWAF",
            icon: ModSec,
            description:
                "Turn on CorazaWAF with the OWASP rules from OpenAdmin. Users can switch it off per domain or disable single rules.",
            url: "/docs/admin/security/waf/",
            status: "enterprise",
        },
        {
            name: "Firewall",
            icon: Firewall,
            description:
                "Manage Sentinel, a fork of ConfigServer Security & Firewall (CSF), from OpenAdmin. Only the ports users need are open.",
            url: "/docs/admin/security/firewall/",
            status: "stable",
        },
        {
            name: "ImunifyAV",
            icon: Services,
            description:
                "OpenAdmin fully supports ImunifyAV.",
            url: "/docs/admin/security/imunify/",
            status: "enterprise",
        },
        {
            name: "Two-Factor Authentication",
            icon: Factor,
            description:
                "Users can turn on two-factor login. Admins can require it or turn it off for any user.",
            url: "/docs/panel/security/2fa/",
            status: "stable",
        },
        {
            name: "ClamAV Malware Scanner",
            icon: Services,
            description:
                "Scan account files for malware using ClamAV directly from the user panel.",
            url: "/docs/panel/security/malware-scanner/",
            status: "enterprise",
        },
        {
            name: "IP Blocker",
            icon: IPblock,
            description:
                "Users can block IP addresses from reaching their websites.",
            url: "/docs/panel/security/ip-blocker/",
            status: "enterprise",
        },
        {
            name: "Disable admin panel",
            icon: AdminOff,
            description:
                "Turn off OpenAdmin with one click. OpenPanel and the websites keep running.",
            url: "/docs/admin/security/disable-admin/",
            status: "stable",
        },
        {
            name: "Separate services",
            icon: Separate,
            description:
                "OpenAdmin and OpenPanel use their own database and web server, separate from the user websites.",
            url: "",
            status: "stable",
        },
        {
            name: "Custom ports",
            icon: Ports,
            description:
                "Change the default port (2083) and the /openpanel path.",
            url: "/docs/admin/settings/general/#ports",
            status: "enterprise",
        },
        {
            name: "Custom domains",
            icon: Domain,
            description:
                "Administrators can set different domains for the OpenPanel and OpenAdmin services.",
            url: "/docs/admin/settings/general/#domain",
            status: "enterprise",
        },
    ],
    "community-packages": [
        {
            name: "WHMCS",
            icon: WHMCS,
            description: "Integrate with the leading web hosting management and billing software.",
            url: "/docs/articles/extensions/openpanel-and-whmcs/",
            status: "enterprise",
        },
        {
            name: "Blesta",
            icon: WHMCS,
            description: "Provision OpenPanel accounts from blesta.",
            url: "https://blesta.club/extensions/openpanel",
            status: "enterprise",
        },
        {
            name: "FOSSBilling",
            icon: WHMCS,
            description: "Provision OpenPanel accounts from FOSSBilling software.",
            url: "https://openpanel.com/docs/articles/extensions/openpanel-and-fossbilling/",
            status: "enterprise",
        },
        {
            name: "ClientExec",
            icon: WHMCS,
            description: "Provision OpenPanel accounts from ClientExec.",
            url: "/docs/articles/extensions/openpanel-and-clientexec/",
            status: "enterprise",
        },
        {
            name: "WISECP",
            icon: WHMCS,
            description: "Provision OpenPanel accounts from WISECP.",
            url: "/docs/articles/extensions/openpanel-and-wisecp/",
            status: "enterprise",
        },
        {
            name: "ClientFence",
            icon: WHMCS,
            description: "Provision OpenPanel accounts from ClientFence, free open-source billing.",
            url: "https://clientfence.com/",
            status: "enterprise",
        },
        {
            name: "API",
            icon: API,
            description:
                "OpenAdmin API allows you to integrate with 3rd party systems you already use.",
            url: "/docs/admin/api/",
            status: "enterprise",
        },
        {
            name: "MCP",
            icon: API,
            description:
                "Built-in MCP server lets AI agents and assistants interact with OpenPanel directly.",
            url: "/docs/panel/account/mcp/",
            status: "enterprise",
        },
    ],
};
