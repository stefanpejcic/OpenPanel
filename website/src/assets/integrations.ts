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
    ServerCog,
    CronJobs,
    Cache,
    DNS,
    Visitors,
    Activity,
    AccountCog,
    Apache,
    Nginx,
} from "./integration-icons";

export const integrations: IntegrationsType = {
    "ui-framework-packages": [
        {
            name: "Apache",
            icon: Apache,
            description:
                "Run Apache web server per user. Apache's support for .htaccess files enables users to customize and override global configuration settings on a per-directory basis.",
            url: "/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/#apache",
            status: "stable",
        },
        {
            name: "Nginx",
            icon: Nginx,
            description:
                "Nginx web server is a lightweight, open-source solution. The OpenPanel version of the Nginx web server enables configuration of cache exclusion, cache purging, URL rewriting, and FastCGI cache on a per-domain basis.",
            url: "/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/#nginx",
            status: "stable",
        },
        {
            name: "OpenResty",
            icon: Nginx,
            description:
                "OpenResty allows per-user customization by defining user-specific location blocks or dynamic Lua logic in the main config. Configuration is centralized and preloaded for performance ",
            url: "/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/#openresty",
            status: "stable",
        },
        {
            name: "OpenLiteSpeed",
            icon: Apache,
            description:
                "OpenLiteSpeed is a high-performance, lightweight web server with built-in caching and .htaccess support. OpenPanel allows per-user configuration for security, caching, and rewrite rules.",
            url: "/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/#openlitespeed",
            status: "stable",
        },
        {
            name: "Varnish",
            icon: Nginx,
            description:
                "Varnish is a powerful HTTP accelerator designed for content-heavy websites. OpenPanel enables WordPress caching rules, purge logic, and backend configuration for optimal performance.",
            url: "/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/",
            status: "stable",
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
                "Effortlessly upload multiple files simultaneously without any upload limits. Edit files, adjust permissions, copy or move files, add new files, and perform various file management tasks.",
            url: "/docs/panel/files/",
            status: "stable",
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
            name: "NodeJS and Python",
            icon: Python,
            description:
                "Effortlessly create and manage containerized NodeJS and Python applications. Proxy websites to display content from these applications seamlessly.",
            url: "/docs/panel/applications/pm2/",
            status: "stable",
        },
        {
            name: "WP Manager",
            icon: WP,
            description: "Automatic WordPress installer with features like auto-login to wp-admin, option editing, on-demand backup, debugging, and a variety of additional functionalities.",
            url: "/docs/panel/applications/wordpress/",
            status: "stable",
        },
        {
            name: "MySQL and phpMyAdmin",
            icon: MySQL,
            description:
                "Create and manage MySQL databases and users easily. Automatically log in to phpMyAdmin, enable remote MySQL access, and adjust configuration settings with desired limits.",
            url: "/docs/panel/mysql/databases/",
            status: "stable",
        },
        {
            name: "SSL certificates",
            icon: SSL,
            description:
                "Automatic SSL generation and renewal ensures that your website's security is effortlessly managed, providing continuous protection with up-to-date SSL certificates.",
            url: "/docs/panel/domains/SSL/",
            status: "stable",
        },
        {
            name: "Object Caching",
            icon: Cache,
            description:
                "Implement object caching using dedicated REDIS and Memcached instances. Set memory limits, start/stop services, and view logs efficiently.",
            url: "/docs/panel/caching/",
            status: "stable",
        },
        {
            name: "Web Terminal",
            icon: Terminal,
            description:
                "Run commands inside Docker containers easily with WebTerminal — a simple interface to access and manage containers directly from your browser.",
            url: "/docs/panel/containers/terminal/",
            status: "stable",
        },
        {
            name: "Server settings",
            icon: Configuration,
            description:
                "Each user has complete control over their server configuration, enabling them to install or restart services, edit system configurations, and perform various other administrative tasks.",
            url: "/docs/panel/advanced/webserver_settings/",
            status: "stable",
        },
        {
            name: "Cron Jobs",
            icon: CronJobs,
            description:
                "Schedule and edit cronjobs directly from the OpenPanel interface to efficiently plan and manage scheduled actions.",
            url: "/docs/panel/advanced/cronjobs/",
            status: "stable",
        },
        {
            name: "DNS Zone Editor",
            icon: DNS,
            description:
                "Easily edit DNS zone files for your domains and add various records such as A, AAAA, CNAME, MX, TXT, etc., through the OpenPanel interface.",
            url: "/docs/panel/domains/dns/",
            status: "stable",
        },
        {
            name: "Resource usage",
            icon: Usage,
            description:
                "Monitor real-time CPU and memory usage, check historical trends, and adjust server configuration as needed to optimize performance.",
            url: "/docs/panel/advanced/resource_usage/",
            status: "stable",
        },
        {
            name: "Visitor reports",
            icon: Visitors,
            description:
                "Access automatically generated, visually appealing visitor reports from your website's access logs. Explore visitor locations, accessed pages, IPs, error pages, and more with ease.",
            url: "/docs/panel/domains/goaccess/",
            status: "stable",
        },
        {
            name: "Activity logs",
            icon: Activity,
            description:
                "Every action in the OpenPanel interface is recorded, allowing users to easily track who did what and when, eliminating the need to sift through server logs.",
            url: "/docs/panel/account/account_activity/",
            status: "stable",
        },
        {
            name: "Account settings",
            icon: AccountCog,
            description:
                "Users can change their email address and password, enable 2FA, adjust language preferences, and activate dark mode for a personalized experience.",
            url: "/docs/panel/account/",
            status: "stable",
        },
    ],
    "community-data-provider-packages": [
        {
            name: "Simple server deployment",
            icon: Download,
            description:
                "Install OpenPanel in minutes. Provision new servers to your cluster with a single command.",
            url: "/docs/admin/intro/#installation",
            status: "stable",
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
                "Currently, only Nginx is supported as the webserver, but upcoming support for LiteSpeed will provide administrators the option to choose their preferred webserver.",
            url: "/",
            status: "stable",
        },
        {
            name: "Edit configuration",
            icon: Configuration,
            description:
                "Administrators have the capability to designate domains for panel access, modify ports, and edit settings for both the OpenPanel and OpenAdmin interfaces, providing flexibility in configuring the system to suit specific requirements.",
            url: "/docs/category/settings/",
            status: "stable",
        },
        {
            name: "Resource usage statistics",
            icon: Usage,
            description:
                "Monitor CPU, Memory usage, Network, Disk and load with real time monitoring.",
            url: "/docs/admin/dashboard/#sse-usage",
            status: "stable",
        },
        {
            name: "Service management",
            icon: Services,
            description:
                "Monitor services, initiate restarts, view logs, and perform additional management tasks efficiently from the admin interface.",
            url: "/docs/admin/services/status/",
            status: "stable",
        },
        {
            name: "Smart Notifications",
            icon: Notifications,
            description:
                "Receive notifications for events such as reboots, high resource usage, website attacks, failed services, and other critical occurrences to stay informed about the status of your server.",
            url: "/docs/admin/notifications/",
            status: "stable",
        },
        {
            name: "Terminal Commands",
            icon: Terminal,
            description:
                "OpenCLI serves as the terminal interface for Administrators, allowing automation of diverse OpenPanel settings with access to over 100 available commands.",
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
            status: "stable",
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
                "OpenPanel features a built-in dark mode that users can activate with a single click directly from the interface, enhancing user experience.",
            url: "/docs/panel/dashboard/dark-mode/",
            status: "stable",
        },
        {
            name: "Server Info",
            icon: ServerInfo,
            description:
                "View real-time usage data, IP address, nameservers, and other important server information directly within the OpenPanel interface.",
            url: "/docs/panel/advanced/server_info/",
            status: "stable",
        },
        {
            name: "Advanced Search",
            icon: Search,
            description:
                "Quickly and easily find what you are looking for with a powerful search functionality.",
            url: "/docs/panel/dashboard",
            status: "stable",
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
                "OpenPanel was designed with a focus on advanced users, offering over 20 keyboard shortcuts to enhance your navigation speed through the interface.",
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
                "Seamlessly upgrade or downgrade a customer's package to another of your hosting packages.",
            url: "/docs/admin/accounts/users/#statistics",
            status: "stable",
        },
        {
            name: "Dedicated IP address",
            icon: UserIP,
            description:
                "Allocate an IPv4 address to users, providing them with a dedicated IP for their websites and services.",
            url: "/docs/admin/accounts/users/#edit",
            status: "stable",
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
                "Administrators can select Apache, Nginx, OpenResty or OpenLitespeed as the default web server for new users. This flexibility allows admins to accommodate a mix of users utilizing Apache and others using Nginx, all within the same server.",
            url: "/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/",
            status: "stable",
        },
        {
            name: "Database server per user",
            icon: UserServer,
            description:
                "Administrators can select either MySQL, Percona or MariaDB as the mysql server for each user. This flexibility allows admins to accommodate a mix of users utilizing MySQL and others using MariaDB, all within the same server.",
            url: "/docs/articles/docker/how-to-set-mysql-mariadb-per-user-in-openpanel/",
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
                "Automatically provision and renewal of Let's Encrypt certificates.",
            url: "/docs/panel/domains/ssl/",
            status: "stable",
        },
        {
            name: "CorazaWAF",
            icon: ModSec,
            description:
                "Administrators can activate CorazaWAF with a single click and configure the OWASP ruleset. Additionally, individual users have the flexibility to enable or disable WAF per domain or disable certain rules.",
            url: "/docs/admin/security/waf/",
            status: "stable",
        },
        {
            name: "Firewall",
            icon: Firewall,
            description:
                "Administrators can manage ConfigServer Firewall (CSF) directly from the admin interface. Only needed ports are open for users, and administrators have full control over them.",
            url: "/docs/admin/security/firewall/",
            status: "stable",
        },
        {
            name: "ImunifyAV",
            icon: Services,
            description:
                "OpenAdmin fully supports ImunifyAV.",
            url: "/docs/admin/security/imunify/",
            status: "stable",
        },
        {
            name: "Two-Factor Authentication",
            icon: Factor,
            description:
                "Users can enhance security by enabling Two-Factor Authentication for their OpenPanel account. Administrators have the flexibility to enforce or disable 2FA for any user.",
            url: "/docs/panel/account/2fa/",
            status: "stable",
        },
       // {
       //     name: "Shell access DEPRECATED",
       //     icon: Terminal,
       //     description:
       //         "End users are restricted from root-level access to their container. Additionally, to increase security, all services within a user's container operate under distinct user accounts.",
       //     url: "/docs/panel/advanced/ssh/",
       //     status: "stable",
       // }, 
       // {
       //     name: "IP blocking DEPRECATED",
       //     icon: IPblock,
       //     description:
       //         "Each user has the capability to configure a domain-specific IP block list, providing a personalized means to restrict access to websites by blocking specific IP addresses.",
       //     url: "",
       //    status: "stable",
       // },
        {
            name: "Disable admin panel",
            icon: AdminOff,
            description:
                "Administrators can effortlessly disable the OpenAdmin interface with a single click, while preserving the core functionality of OpenPanel.",
            url: "/docs/admin/security/disable-admin/",
            status: "stable",
        },
        {
            name: "Separate services",
            icon: Separate,
            description:
                "Both OpenAdmin and OpenPanel employ separate databases and webservers, maintaining full independence from user websites.",
            url: "",
            status: "stable",
        },
        {
            name: "Custom ports",
            icon: Ports,
            description:
                "Administrators have the flexibility to customize the default port (e.g., change from 2083) and alter the directory path (e.g., from /openpanel) to cater to specific preferences.",
            url: "/docs/admin/settings/openpanel/",
            status: "stable",
        },
    ],
    "community-packages": [
        {
            name: "WHMCS",
            icon: WHMCS,
            description: "Integrate with the leading web hosting management and billing software.",
            url: "/docs/articles/extensions/openpanel-and-whmcs/",
            status: "stable",
        },
        {
            name: "Blesta",
            icon: WHMCS,
            description: "Provision OpenPanel accounts from blesta.",
            url: "https://blesta.club/extensions/openpanel",
            status: "stable",
        },
        {
            name: "FOSSBilling",
            icon: WHMCS,
            description: "Provision OpenPanel accounts from FOSSBilling software.",
            url: "https://openpanel.com/docs/articles/extensions/openpanel-and-fossbilling/",
            status: "stable",
        },        
        {
            name: "API",
            icon: API,
            description:
                "OpenAdmin API allows you to integrate with 3rd party systems you already use.",
            url: "https://dev.openpanel.com/openadmin-api/",
            status: "stable",
        },
    ],
};
