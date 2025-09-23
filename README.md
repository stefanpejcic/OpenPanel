## What is OpenPanel

OpenPanel is *probably* the most customizable web hosting control panel.

Available in an community-supported version, and a more feature-filled version with premium support, OpenPanel is the cost-effective and comprehensive solution to web hosting management.

```
╔════════════════════════════════════════════════════════════════╗
║                     🖥️  OPENPANEL SERVER                       ║
╠════════════════════════════════════════════════════════════════╣
║  • 🎛️ OpenPanel - user control panel                           ║
║  • ⚙️ OpenAdmin - administration panel                         ║
║  • 🌐 Caddy – Reverse Proxy & SSL                              ║
║  • 🔍 BIND9 – DNS Server                                       ║
║  • 🗄️ MySQL – User Management & Metadata                       ║
║  • 🐳 Docker Engine – Container Orchestration                  ║
╚════════════════════════════════════════════════════════════════╝
                                                   │   
        ┌──────────────────────────────────────────┼──────────────────────────────────────────┐
        │                                          │                                          │
        ▼                                          ▼                                          ▼
┌─────────────────────────────────┐ ┌─────────────────────────────────┐ ┌─────────────────────────────────┐
│           👤 USER 1             │ │           👤 USER 2             │ │           👤 USER 3             │ 
├─────────────────────────────────┤ ├─────────────────────────────────┤ ├─────────────────────────────────┤
│  🌐 Web Server:                 │ │  🌐 Web Server:                 │ │  🌐 Web Server:                 │
│  • Nginx + Varnish              │ │  • OpenLitespeed                │ │  • Apache + Varnish             │
│                                 │ │                                 │ │                                 │
│  ⚡ Applications:               │ │  ⚡ Applications:               │ │  ⚡ Applications:               │
│  • site1.com → PHP 8.4          │ │  • api.site.com → Node.js 20.1  │ │  • classic.com → PHP 7.0        │
│  • site2.com → PHP 8.2          │ │  • main.site.com → PHP 8.3      │ │  • modern.com → PHP 8.1         │
│  • legacy.com → PHP 7.0         │ │                                 │ │  • vintage.com → PHP 5.6        │
│                                 │ │                                 │ │  • api.site.com → Python 3.11   │
│                                 │ │                                 │ │                                 │
│  🗄️  Databases:                 │ │  🗄️  Databases:                 │ │  🗄️  Databases:                 │
│  • MySQL 8.0                    │ │  • MariaDB 10.11                │ │  • Percona MySQL                │
│  • phpMyAdmin                   │ │  • phpMyAdmin                   │ │  • PostgreSQL                   │
├─────────────────────────────────┤ ├─────────────────────────────────┤ ├─────────────────────────────────┤
│  📊 Resource Limits:            │ │  📊 Resource Limits:            │ │  📊 Resource Limits:            │
│  • CPU: 2 cores                 │ │  • CPU: 4 cores                 │ │  • CPU: 1 core                  │
│  • RAM: 4 GB                    │ │  • RAM: 8 GB                    │ │  • RAM: 2 GB                    │
│  • Storage: 50 GB SSD           │ │  • Storage: 100 GB SSD          │ │  • Storage: 25 GB SSD           │
└─────────────────────────────────┘ └─────────────────────────────────┘ └─────────────────────────────────┘
```

## Why use OpenPanel?

**OpenPanel** offers a distinct advantage over other hosting panels by providing each user with an isolated environment and tools to fully manage it. This ensures that your users enjoy full control over their environment, simillar to a VPS experience. They can effortlessly run multiple PHP versions, modify server configurations, view domain logs, restart services, set limits, configure backups and more.

**Why use OpenPanel for your hosting business?**

- focus on [security](https://openpanel.com/docs/articles/security/securing-openpanel/)
- billing integrations: [FOSSBilling](https://openpanel.com/docs/articles/extensions/openpanel-and-fossbilling/), [WHMCS](https://openpanel.com/docs/articles/extensions/openpanel-and-whmcs/), [Blesta](https://openpanel.com/docs/articles/extensions/openpanel-and-blesta/) paymenter.org[^1]
- dedicated [MySQL, Percona or MariaDB per user](https://openpanel.com/docs/articles/docker/how-to-set-mysql-mariadb-per-user-in-openpanel/)
- dedicated [Apache, Nginx, OpenLitespeed, Openresty + Varnish per user](https://openpanel.com/docs/articles/docker/how-to-set-nginx-apache-varnish-per-user-in-openpanel/)
- detailed activity log of all user actions.
- low maintenance: each user manages their own services and [backups](https://openpanel.com/docs/panel/files/backups/)
- [import cPanel accounts](https://openpanel.com/docs/articles/transfers/import-cpanel-backup-to-openpanel/)
- [white label](https://openpanel.com/docs/articles/dev-experience/customizing-openpanel-user-interface/)


[^1]: not actively maintained

----

## OpenPanel vs OpenAdmin

- The **OpenAdmin** offers an administrator-level interface where you can efficiently handle tasks such as creating and managing users, setting up hosting plans, and editing OpenPanel settings.
- The **OpenPanel** interface is the client-level panel where end-users can manage their containers: edit settings, configure limits, manage backups, create websites and more.

## Supported OS

OpenPanel is a truly [OS-agnostic](https://www.techtarget.com/whatis/definition/agnostic) control panel. Supported OS:

| Operating System       | Versions                             | Notes                                |
|------------------------|--------------------------------------|--------------------------------------|
| Ubuntu                 | 22.04, 24.04                         | **24.04 is recommended for AMD CPU**  |
| Debian                 | 10, 11, 12, 13                       |                     |
| AlmaLinux              | 9.5, 10                              | **9.5 is recommended for ARM CPU** |
| RockyLinux             | 9.6, 10                              | *On Rocky 10, you must manually switch from `nftables` to `iptables` first — see [#1472](https://github.com/docker/for-linux/issues/1472)* |
| CentOS                 | 9.5                                  |                                      |

## 📥 Installation

To install on self-hosted VPS/Dedicated server: 
```bash
bash <(curl -sSL https://openpanel.org)
```

To see more details to configure server on installation, **please visit**: https://openpanel.com/install 

## Documentation

- [OpenAdmin - Admin panel documentation](https://openpanel.com/docs/admin/intro/)
- [OpenPanel - End-user panel documentation](https://openpanel.com/docs/panel/intro/)
- [Guides and How-to](https://openpanel.com/docs/articles/intro/)
- [OpenCLI - Terminal commands](https://dev.openpanel.com/cli/)

## Support

Our [Community](https://community.openpanel.org/) serves as our virtual Headquater, where the community helps each other.

**Learn, share** and **discuss** with other community members your questions.

## Version

Latest OpenPanel version is: **1.6.1** - [View Changelog](https://openpanel.com/docs/changelog/1.6.1/)

[![Alt](https://repobeats.axiom.co/api/embed/9904d020c32812f0aff8d8d69f52643d16f85007.svg "Repobeats analytics image")](https://openpanel.com/statistics)

## Copyright & license

- OpenAdmin and OpenPanel UI are distributed under EULA.
- OpenCLI and configuration files are distributed under Commons Attribution-NonCommercial (CC BY-NC) license.

## Contributing

We welcome and appreciate all contributions - technical or not!

You don’t need to be a developer to make a meaningful impact.
Plese see [CONTRIBUTING.md](https://github.com/stefanpejcic/OpenPanel/blob/main/CONTRIBUTING.md)
