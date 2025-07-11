## What is OpenPanel

OpenPanel is *probably* the most customizable web hosting control panel.


```
╔════════════════════════════════════════════════════════════════╗
║                      🖥️  HOST SERVER                          ║
╠════════════════════════════════════════════════════════════════╣
║  Management Layer:                                             ║
║  • 🎛️  OpenPanel (Control Panel)                              ║
║  • ⚙️  OpenAdmin (Admin Interface)                            ║
║                                                                ║
║  Infrastructure Services:                                      ║
║  • 🌐 Caddy (Reverse Proxy & SSL Termination)                 ║
║  • 🔍 BIND9 (DNS Server)                                      ║
║  • 🗄️  Global MySQL (User Management & Metadata)             ║
║  • 🐳 Docker Engine (Container Orchestration)                 ║
╚════════════════════════════════════════════════════════════════╝
                                   │
        ┌──────────────────────────┼──────────────────────────┐
        │                          │                          │
        ▼                          ▼                          ▼
┌─────────────────────────────────┐  ┌─────────────────────────────────┐  ┌─────────────────────────────────┐
│        👤 USER 1 TENANT         │  │        👤 USER 2 TENANT         │  │        👤 USER 3 TENANT         │
│     🐳 Docker Context           │  │     🐳 Docker Context           │  │     🐳 Docker Context           │
├─────────────────────────────────┤  ├─────────────────────────────────┤  ├─────────────────────────────────┤
│                                 │  │                                 │  │                                 │
│  🌐 Web Server Layer:           │  │  🌐 Web Server Layer:           │  │  🌐 Web Server Layer:           │
│  • Nginx (Load Balancer)        │  │  • Apache (HTTP Server)         │  │  • OpenResty (Nginx + Lua)      │
│  • Varnish (HTTP Cache)         │  │  • Varnish (HTTP Cache)         │  │  • Varnish (HTTP Cache)         │
│                                 │  │                                 │  │                                 │
│  ⚡ Application Layer:          │  │  ⚡ Application Layer:          │  │  ⚡ Application Layer:          │
│  🐘 PHP containers:   │  │  🐘 PHP containers:   │  │  🐘 PHP containers:   │
│  • PHP 8.4 (php84-fpm)         │  │  • PHP 7.4 (php74-fpm)         │  │  • PHP 5.6 (php56-fpm)         │
│  • PHP 8.2 (php82-fpm)         │  │                                 │  │  • PHP 7.0 (php70-fpm)         │
│  • PHP 7.0 (php70-fpm)         │  │  📱 Node.js Applications:        │  │  • PHP 8.1 (php81-fpm)         │
│                                 │  │  • Node.js App 1 (v18.17.0)    │  │                                 │
│  🎯 Application Routing:        │  │  • Node.js App 2 (v20.5.1)     │  │  🐍 Python Environment:         │
│  • site1.com → PHP 8.4         │  │                                 │  │  • Python 3.11 (Flask/Django)  │
│  • site2.com → PHP 8.2         │  │  🎯 Application Routing:        │  │                                 │
│  • legacy.com → PHP 7.0        │  │  • api.site.com → Node.js App1 │  │  🎯 Application Routing:        │
│                                 │  │  • app.site.com → Node.js App2 │  │  • modern.com → PHP 8.1        │
│  📦 Version Management:         │  │  • main.site.com → PHP 7.4     │  │  • classic.com → PHP 7.0       │
│  • phpbrew (PHP switcher)       │  │                                 │  │  • vintage.com → PHP 5.6       │
│  • Composer (per PHP version)   │  │  📦 Version Management:         │  │  • api.site.com → Python App   │
│                                 │  │  • nvm (Node Version Manager)   │  │                                 │
│                                 │  │  • npm/yarn (Package Managers)  │  │  📦 Version Management:         │
│                                 │  │  • Composer (PHP 7.4)          │  │  • phpbrew (PHP switcher)       │
│                                 │  │  • PM2 (Process Manager)        │  │  • pyenv (Python versions)      │
│                                 │  │                                 │  │  • pip/poetry (Python packages) │
│                                 │  │                                 │  │  • Composer (per PHP version)   │
│                                 │  │                                 │  │                                 │
│  🗄️  Database Layer:            │  │  🗄️  Database Layer:            │  │  🗄️  Database Layer:            │
│  • MySQL 8.0 (Primary DB)      │  │  • MariaDB 10.11 (Primary DB)  │  │  • MySQL 5 (Primary DB)      │
│  • phpMyAdmin (DB Management)  │  │  • phpMyAdmin (DB Management)  │  │  • phpMyAdmin (DB Management)  │
│                                 │  │  • MongoDB (NoSQL for Node.js) │  │  • PostgreSQL (Python apps)    │
│                                 │  │                                 │  │                                 │
│  🔧 Operations Layer:           │  │  🔧 Operations Layer:           │  │  🔧 Operations Layer:           │
│  • Backup Tool (Automated)     │  │  • Backup Tool (Automated)     │  │  • Backup Tool (Automated)     │
│  • Cron Jobs (Task Scheduler)  │  │  • Cron Jobs (Task Scheduler)  │  │  • Cron Jobs (Task Scheduler)  │
│                                 │  │                                 │  │                                 │
│  📊 Resource Limits:            │  │  📊 Resource Limits:            │  │  📊 Resource Limits:            │
│  • CPU: 2 cores                │  │  • CPU: 4 cores                │  │  • CPU: 1 core                 │
│  • RAM: 4GB                    │  │  • RAM: 8GB                    │  │  • RAM: 2GB                    │
│  • Storage: 50GB SSD           │  │  • Storage: 100GB SSD          │  │  • Storage: 25GB SSD           │
│                                 │  │                                 │  │                                 |
└─────────────────────────────────┘  └─────────────────────────────────┘  └─────────────────────────────────┘
```

Available in an community-supported version, and a more feature-filled version with premium support, OpenPanel is the cost-effective and comprehensive solution to web hosting management.

## Why use OpenPanel to host websites?

**OpenPanel** offers a distinct advantage over other hosting panels by providing each user with an isolated environment and tools to fully manage it. This ensures that your users enjoy full control over their environment, simillar to a VPS experience. They can effortlessly run multiple PHP versions, modify server configurations, view domain logs, restart services, set limits, configure backups and more.

## Why use OpenPanel for your hosting business?

- focus on security
- billing integrations: [FOSSBilling](https://openpanel.com/docs/articles/extensions/openpanel-and-fossbilling/), [WHMCS](https://openpanel.com/docs/articles/extensions/openpanel-and-whmcs/), [Blesta](https://openpanel.com/docs/articles/extensions/openpanel-and-blesta/)
- dedicated [MySQL or MariaDB per user](https://openpanel.com/docs/articles/docker/how-to-set-mysql-mariadb-per-user-in-openpanel/)
- dedicated [Apache or Nginx + Varnish per user](https://openpanel.com/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/)
- detailed activity log of all user actions.
- low maintenance: each user manages their own services and [backups](https://openpanel.com/docs/panel/files/backups/)
- [import cPanel accounts](https://openpanel.com/docs/articles/transfers/import-cpanel-backup-to-openpanel/)
- [white label](https://openpanel.com/docs/articles/dev-experience/customizing-openpanel-user-interface/)


----

This panel is the culmination of years of experience in the hosting industry, having spent decades working with various hosting panels we made sure to include all features that simply make sense.

When we designed OpenPanel, we prioritized features that are not only user-friendly for beginners but also advanced enough to alleviate maintenance tasks for system administrators and hosting support teams.

Some of the features worth mentioning are:

- All services are containerized.
- Webserver per user: Nginx, Apache, OpenResty and/or Varnish
- MySQL or MariaDB per user.
- Users can switch webserver and mysql type.
- Users set PHP version per domain.
- Users set CPU and Memory limits for services.
- Users configure their own backup destinations.

And unique features that simply made sense 💁 to us:
- User and admin panels are completelly isolated
- SSL is automatically generated and renewed
- Services auto-start only when needed so resources are not wasted
- Gooogle PageSpeed data is automatically displayed for every website in Site Manager
- Users can export DNS zones easily
- Users can suspend websites
- Administrators can receive daily usage reports
- Users can add comments for DNS records
- Download files from URL and drag-and-drop file upload in File Manager
- Users can save pages to Favorites for quick navigation
- Users can view/terminate their active sessions
- Detailed activity log of all actions
- Admins can add custom message per user
- and a lot more 🙌

## OpenPanel vs OpenAdmin

The **OpenAdmin** offers an administrator-level interface where you can efficiently handle tasks such as creating and managing users, setting up hosting plans, and editing OpenPanel settings.

The **OpenPanel** interface is the client-level panel where end-users can manage their containers: edit settings, configure limits, manage backups, create websites and more.

## Supported OS

OpenPanel is a truly [OS-agnostic](https://www.techtarget.com/whatis/definition/agnostic) control panel.


Supported OS:
- Ubuntu 22.04 and 24.04 (recommended)
- Debian [10](https://voidnull.es/instalacion-de-openpanel-en-debian-10/), 11 and 12
- AlmaLinux 9.5 *(recommended for ARM cpu)
- RockyLinux 9.3
- CentOS 9.5

## Installation

To install on self-hosted VPS/Dedicated server: 

```bash
bash <(curl -sSL https://openpanel.org)
```

To see more details to configure server on installation, **please visit**: https://openpanel.com/install 

## Support

Our [Community](https://community.openpanel.org/) serves as our virtual Headquater, where the community helps each other.

**Learn, share** and **discuss** with other community members your questions.

## Version

Latest OpenPanel version is: **1.4.7** - [View Changelog](https://openpanel.com/docs/changelog/1.4.7/)

![Alt](https://repobeats.axiom.co/api/embed/9904d020c32812f0aff8d8d69f52643d16f85007.svg "Repobeats analytics image")

## Copyright & license

- OpenAdmin and OpenPanel UI are distributed under EULA.
- OpenCLI and configuration files are distributed under Commons Attribution-NonCommercial (CC BY-NC) license.

## Contributing

We welcome and appreciate all contributions - technical or not!

You don’t need to be a developer to make a meaningful impact.
Plese see [CONTRIBUTING.md](https://github.com/stefanpejcic/OpenPanel/blob/main/CONTRIBUTING.md)
