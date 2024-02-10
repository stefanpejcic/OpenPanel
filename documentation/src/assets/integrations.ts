import { IntegrationsType } from "../types/integrations";
import {
    Ably,
    Airtable,
    Antd,
    Appwrite,
    Chakra,
    Directus,
    Dp,
    Elide,
    ElideGraphql,
    EntRefine,
    Firebase,
    Graphql,
    Hasura,
    HookForm,
    Hygraph,
    JSONApi,
    Kbar,
    Mantine,
    Medusa,
    Mui,
    Nest,
    NestQuery,
    Nextjs,
    React,
    Remix,
    Rest,
    Sanity,
    SQLite,
    Strapi,
    Supabase,
    UseGenerated,
    Kinde,
} from "./integration-icons";

export const integrations: IntegrationsType = {
    "ui-framework-packages": [
        {
            name: "Nginx",
            icon: Mui,
            description:
                "Run Nginx web server is a lightweight, open-source solution. The OpenPanel version of the Nginx web server enables configuration of cache exclusion, cache purging, URL rewriting, and FastCGI cache on a per-domain basis.",
            url: "/docs/admin/",
            status: "stable",
        },
        {
            name: "Apache",
            icon: Antd,
            description:
                "Run Apache web server per user. Apache's support for .htaccess files enables users to customize and override global configuration settings on a per-directory basis.",
            url: "/docs/admin/",
            status: "stable",
        },
    ],
    "data-provider-packages": [
        {
            name: "Domain Names",
            icon: Rest,
            description:
                "Add domain names with automatic SSL renewals, include aliases and subdomains (Internationalized domains are supported), create redirects, enforce HTTPS, and edit vhost files.",
            url: "/",
            status: "stable",
        },
        {
            name: "File Manager",
            icon: Rest,
            description:
                "Effortlessly upload multiple files simultaneously without any upload limits. Edit files, adjust permissions, copy or move files, add new files, and perform various file management tasks.",
            url: "/",
            status: "stable",
        },
        {
            name: "PHP versions",
            icon: Rest,
            description:
                "Users can use different PHP versions for each domain, install new versions, set a default version for new domains, change limits by editing php.ini files.",
            url: "/",
            status: "stable",
        },
        {
            name: "NodeJS and Python",
            icon: Rest,
            description:
                "Effortlessly create and manage NodeJS and Python applications. Proxy websites to display content from these applications seamlessly.",
            url: "/",
            status: "stable",
        },
        {
            name: "WP Manager",
            icon: Graphql,
            description: "Automatic WordPress installer with features like auto-login to wp-admin, option editing, on-demand backup, debugging, and a variety of additional functionalities.",
            url: "/",
            status: "stable",
        },
        {
            name: "MySQL and phpMyAdmin",
            icon: Nest,
            description:
                "Create and manage MySQL databases and users easily. Automatically log in to phpMyAdmin, enable remote MySQL access, and adjust configuration settings with desired limits.",
            url: "/",
            status: "stable",
        },
        {
            name: "SSL certificates",
            icon: Strapi,
            description:
                "Automatic SSL generation and renewal ensures that your website's security is effortlessly managed, providing continuous protection with up-to-date SSL certificates.",
            url: "/",
            status: "stable",
        },
        {
            name: "Object Caching",
            icon: Strapi,
            description:
                "Implement object caching using dedicated REDIS and Memcached instances. Set memory limits, start/stop services, and view logs efficiently.",
            url: "/",
            status: "stable",
        },
        {
            name: "SSH and Web Terminal",
            icon: Strapi,
            description:
                "Access the terminal remotely through SSH or log in automatically to the Web Terminal. Comes with preinstalled WPCLI and NPM for added convenience!",
            url: "/",
            status: "stable",
        },
        {
            name: "Server settings",
            icon: Strapi,
            description:
                "Each user has complete control over their server configuration, enabling them to install or restart services, edit system configurations, and perform various other administrative tasks.",
            url: "/",
            status: "stable",
        },
        {
            name: "Cron Jobs",
            icon: Strapi,
            description:
                "Schedule and edit cronjobs directly from the OpenPanel interface to efficiently plan and manage scheduled actions.",
            url: "/",
            status: "stable",
        },
        {
            name: "DNS Zone Editor",
            icon: Strapi,
            description:
                "Easily edit DNS zone files for your domains and add various records such as A, AAAA, CNAME, MX, TXT, etc., through the OpenPanel interface.",
            url: "/",
            status: "stable",
        },
        {
            name: "Resource usage",
            icon: Supabase,
            description:
                "Monitor real-time CPU and memory usage, check historical trends, and adjust server configuration as needed to optimize performance.",
            url: "/",
            status: "stable",
        },
        {
            name: "Visitor reports",
            icon: Hasura,
            description:
                "Access automatically generated, visually appealing visitor reports from your website's access logs. Explore visitor locations, accessed pages, IPs, error pages, and more with ease.",
            url: "/",
            status: "stable",
        },
        {
            name: "Activity logs",
            icon: Hasura,
            description:
                "Every action in the OpenPanel interface is recorded, allowing users to easily track who did what and when, eliminating the need to sift through server logs.",
            url: "/",
            status: "stable",
        },
        {
            name: "Account settings",
            icon: Hasura,
            description:
                "Users can change their email address and password, enable 2FA, adjust language preferences, and activate dark mode for a personalized experience.",
            url: "/",
            status: "stable",
        },
    ],
    "community-data-provider-packages": [
        {
            name: "Simple server deployment",
            icon: Directus,
            description:
                "Install OpenPanel in minutes. Provision new servers to your cluster with a single command.",
            url: "/",
            status: "stable",
        },
        {
            name: "Resource usage management",
            icon: Firebase,
            description: "Limit the CPU, I/O bandwidth, IOPS, nproc and memory on a per-user basis to ensure consistent performance for all your hosted websites.",
            url: "/",
            status: "stable",
        },
        {
            name: "Switch web servers",
            icon: Hygraph,
            description:
                "Currently, only Nginx is supported as the webserver, but upcoming support for LiteSpeed will provide administrators the option to choose their preferred webserver.",
            url: "/",
            status: "stable",
        },
        {
            name: "Edit configuration",
            icon: Sanity,
            description:
                "Administrators have the capability to designate domains for panel access, modify ports, and edit settings for both the OpenPanel and OpenAdmin interfaces, providing flexibility in configuring the system to suit specific requirements.",
            url: "/",
            status: "stable",
        },
        {
            name: "Resource usage statistics",
            icon: Elide,
            description:
                "Monitor CPU and Memory usage, Network and load with real time monitoring.",
            url: "/",
            status: "stable",
        },
        {
            name: "Service management",
            icon: ElideGraphql,
            description:
                "Monitor services, initiate restarts, view logs, and perform additional management tasks efficiently from the admin interface.",
            url: "/",
            status: "stable",
        },
        {
            name: "Smart Notifications",
            icon: EntRefine,
            description:
                "Receive notifications for events such as reboots, high resource usage, website attacks, failed services, and other critical occurrences to stay informed about the status of your server.",
            url: "/",
            status: "stable",
        },
    ],
    frameworks: [
        {
            name: "Branding",
            icon: Airtable,
            description: "Fully customise the OpenPanel with colours, logos, fonts and more that mirror the look and feel of your hosting company.",
            url: "/",
            status: "stable",
        },
        {
            name: "Fully responsive",
            icon: Medusa,
            description:
                "Access OpenPanel on desktop, tablet and mobile without any feature limitations.",
            url: "/link",
            status: "stable",
        },
        {
            name: "Dark Mode",
            icon: Appwrite,
            description:
                "OpenPanel features a built-in dark mode that users can activate with a single click directly from the interface, enhancing user experience.",
            url: "",
            status: "stable",
        },
        {
            name: "Advanced Search",
            icon: NestQuery,
            description:
                "Quickly and easily find what you are looking for with a powerful search functionality.",
            url: "",
            status: "stable",
        },
        {
            name: "Multi-language support",
            icon: Appwrite,
            description:
                "OpenPanel is translation ready. Each login can view OpenPanel in their preferred language.",
            url: "",
            status: "stable",
        },
    ],
    integrations: [
        {
            name: "Suspend / Delete accounts",
            icon: React,
            description:
                "Suspend customer accounts to instantly disable their OpenPanel access and websites. Delete accounts when they are no longer required.",
            url: "https://www.npmjs.com/package/@refinedev/react-table",
            status: "stable",
        },
        {
            name: "Upgrade / Downgrade package",
            icon: HookForm,
            description:
                "Seamlessly upgrade or downgrade a customer's package to another of your hosting packages.",
            url: "/",
            status: "stable",
        },
        {
            name: "Dedicated IP address",
            icon: Sanity,
            description:
                "Allocate an IPv4 address to users, providing them with a dedicated IP for their websites and services.",
            url: "/",
            status: "stable",
        },
        {
            name: "Impersonation",
            icon: Kbar,
            description: "Auto-login to access a customer's account and see exactly what they see without having to leave your account.",
            url: "/",
            status: "stable",
        },
        {
            name: "Apache or nginx per user",
            icon: React,
            description:
                "Administrators can select either Apache or Nginx as the web server for each user. This flexibility allows admins to accommodate a mix of users utilizing Apache and others using Nginx, all within the same server.",
            url: "/",
            status: "stable",
        },
    ],
    "live-providers": [
        {
            name: "User containerisation",
            icon: Ably,
            description:
                "Every user account is containerised. Containers have no access to other users or server files.",
            url: "/",
            status: "stable",
        },
        {
            name: "Resource limiting",
            icon: Ably,
            description:
                "Administrators have the ability to set specific limits per plan, including port speed, disk usage, inodes, the number of websites, MySQL databases, and domains.",
            url: "/",
            status: "stable",
        },
        {
            name: "SSL / TLS",
            icon: Ably,
            description:
                "Automatically provision Let's Encrypt certificates. Users can generate new certificates and seamlessly redirect all website traffic to HTTPS.",
            url: "/",
            status: "stable",
        },
        {
            name: "ModSecurity",
            icon: Ably,
            description:
                "Administrators can activate ModSecurity with a single click and configure the OWASP ruleset. Additionally, individual users have the flexibility to enable or disable ModSecurity per domain.",
            url: "/",
            status: "stable",
        },
        {
            name: "Firewall",
            icon: Ably,
            description:
                "Administrators can manage (UFW) firewall rules directly from the admin interface. Only needed ports are open for users, and administrators have full control over them.",
            url: "/",
            status: "stable",
        },
        {
            name: "Two-Factor Authentication",
            icon: Ably,
            description:
                "Users can enhance security by enabling Two-Factor Authentication for their OpenPanel account. Administrators have the flexibility to enforce or disable 2FA for any user.",
            url: "/",
            status: "stable",
        },
        {
            name: "Limited shell access",
            icon: Ably,
            description:
                "End users are restricted from root-level access to their container. Additionally, to increase security, all services within a user's container operate under distinct user accounts.",
            url: "/",
            status: "stable",
        }, 
        {
            name: "IP blocking",
            icon: Ably,
            description:
                "Each user has the capability to configure a domain-specific IP block list, providing a personalized means to restrict access to websites by blocking specific IP addresses.",
            url: "/",
            status: "stable",
        },
        {
            name: "Disable admin panel",
            icon: Ably,
            description:
                "Administrators can effortlessly disable the OpenAdmin interface with a single click, while preserving the core functionality of OpenPanel.",
            url: "/",
            status: "stable",
        },
        {
            name: "Separate services",
            icon: Ably,
            description:
                "Both OpenAdmin and OpenPanel employ separate databases and webservers, maintaining full independence from user websites.",
            url: "/",
            status: "stable",
        },
        {
            name: "Custom ports",
            icon: Ably,
            description:
                "Administrators have the flexibility to customize the default port (e.g., change from 2083) and alter the directory path (e.g., from /openpanel) to cater to specific preferences.",
            url: "/",
            status: "stable",
        },
    ],
    "community-packages": [
        {
            name: "WHMCS Module",
            icon: Dp,
            description: "Integrate with the leading web hosting management and billing software.",
            url: "https://www.npmjs.com/package/data-provider-customizer",
            status: "stable",
        },
        {
            name: "REST API",
            icon: Kinde,
            description:
                "Our powerful RESTful API allows you to integrate with 3rd party systems you already use.",
            url: "/",
            status: "stable",
        },
    ],
};
