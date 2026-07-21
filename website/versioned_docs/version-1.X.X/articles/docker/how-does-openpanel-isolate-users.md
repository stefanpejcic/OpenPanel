# User Isolation

OpenPanel is built with **security-first architecture**, enforcing multiple levels of isolation to protect user environments.

Isolation Layers:

* **System Isolation**
  Each OpenPanel user corresponds to a system user on the host machine. These users have no login access or passwords and are strictly used for enforcing **storage quotas and ownership**.

* **User Isolation**
  Every OpenPanel account runs inside a **dedicated Docker context**. This ensures one user cannot access, interfere with, or even detect the existence of other users or their containers.

* **Service Isolation**
  Each service within a user’s environment (e.g., PHP, MySQL, Redis) runs in its **own container**. Services are sandboxed even from each other, so if one is compromised (e.g., MySQL), it **cannot affect other containers**, even within the same user.

* **Network Isolation**
  User services are segmented into **internal Docker networks**, such as:

  * `www` for web-facing components (PHP, Nginx, file manager)
  * `db` for databases (MySQL, MariaDB, PostgreSQL)
  * `none` for isolated services (Redis, Memcached)

  This design allows fine-grained control over what services can communicate and also supports **per-user bandwidth throttling**.


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
│  • Nginx + Varnish              │ │  • Apache                       │ │  • OpenResty + Varnish          │
│                                 │ │                                 │ │                                 │
│  ⚡ Applications:               │ │  ⚡ Applications:               │ │  ⚡ Applications:               │
│  • site1.com → PHP 8.4          │ │  • api.site.com → Node.js 20.1  │ │  • classic.com → PHP 7.0        │
│  • site2.com → PHP 8.2          │ │  • main.site.com → PHP 7.4      │ │  • modern.com → PHP 8.1         │
│  • legacy.com → PHP 7.0         │ │                                 │ │  • vintage.com → PHP 5.6        │
│                                 │ │                                 │ │  • api.site.com → Python 3.11   │
│                                 │ │                                 │ │                                 │
│  🗄️  Databases:                 │ │  🗄️  Databases:                 │ │  🗄️  Databases:                 │
│  • MySQL 8.0                    │ │  • MariaDB 10.11                │ │  • PostgreSQL                   │
│  • phpMyAdmin                   │ │  • phpMyAdmin                   │ │                                 │
├─────────────────────────────────┤ ├─────────────────────────────────┤ ├─────────────────────────────────┤
│  📊 Resource Limits:            │ │  📊 Resource Limits:            │ │  📊 Resource Limits:            │
│  • CPU: 2 cores                 │ │  • CPU: 4 cores                 │ │  • CPU: 1 core                  │
│  • RAM: 4 GB                    │ │  • RAM: 8 GB                    │ │  • RAM: 2 GB                    │
│  • Storage: 50 GB SSD           │ │  • Storage: 100 GB SSD          │ │  • Storage: 25 GB SSD           │
└─────────────────────────────────┘ └─────────────────────────────────┘ └─────────────────────────────────┘
```
