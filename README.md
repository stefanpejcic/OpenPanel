## What is OpenPanel?

OpenPanel is a powerful and flexible web hosting control panel for Linux systems. Available in a community-supported version and a feature-rich version with premium support, OpenPanel is a cost-effective and comprehensive solution for web hosting management.

**OpenPanel** offers a distinct advantage over other hosting panels by providing each user with an isolated environment and tools to fully manage it. This ensures full control over your environment, similar to a VPS experience. You can effortlessly run multiple PHP versions, modify server configurations, view domain logs, restart services, and perform numerous other advanced tasks.

Deliver a **VPS-like experience** to your users at a fraction of the cost, with all-inclusive features such as **resource limiting, user isolation, WP Manager, and enhanced security** seamlessly integrated for worry-free hosting.

[![OpenPanel Scheme](/website/static/img/admin/openpanel_scheme.png)](https://openpanel.com/docs/panel/intro/)

This panel is the culmination of years of experience in the hosting industry. Having spent decades working with various hosting panels, we made sure to include all features that simply make sense.

When designing OpenPanel, we prioritized features that are not only user-friendly for beginners but also advanced enough to alleviate maintenance tasks for system administrators and hosting support teams.

### Key Features

- **Web Servers**: Run [Nginx or Apache](https://openpanel.com/docs/admin/plans/hosting_plans/#list-hosting-plans).
- **Databases**: Use [MySQL or MariaDB](https://openpanel.com/docs/articles/docker/how-to-set-mysql-mariadb-per-user-in-openpanel/).
- **PHP Management**: [Install PHP versions](https://openpanel.com/docs/panel/advanced/server_settings#install-php-version), [edit php.ini files](https://openpanel.com/docs/panel/advanced/server_settings#phpini-editor), and set desired limits.
- **MySQL Settings**: [Control MySQL settings](https://openpanel.com/docs/panel/advanced/server_settings#mysql-settings), set limits, [enable remote MySQL access](https://openpanel.com/docs/panel/databases/remote), and more.
- **System Services**: [Update system services](https://openpanel.com/docs/panel/advanced/server_settings#service-status) and install new services as needed.
- **WordPress Management**: Easily manage WordPress websites with [WP Manager](https://openpanel.com/docs/panel/applications/wordpress).
- **Password-less Access**: Access [phpMyAdmin](https://openpanel.com/docs/panel/databases/phpmyadmin) and [Web Terminal](https://openpanel.com/docs/panel/advanced/terminal) without passwords.
- **Caching**: Built-in [Redis](https://openpanel.com/docs/panel/caching/Redis) and [Memcached](https://openpanel.com/docs/panel/caching/Memcached) object caching.

### Unique Features

- User and admin panels are completely isolated when needed.
- SSL certificates are automatically generated and renewed.
- Services auto-start only when needed, saving resources.
- Google PageSpeed data is automatically displayed for every website in the Site Manager.
- Users can export DNS zones, pause cron jobs, suspend websites, and more.
- Administrators receive daily usage reports.
- Users can add comments to DNS records, download files from URLs, save pages to Favorites, and view all active sessions.
- Users can share web terminal sessions with third parties.
- Detailed activity logs of all actions.
- Administrators can add custom messages per user.
- And much more! 🙌

---

## OpenPanel vs OpenAdmin

The **OpenAdmin** interface is designed for administrators to efficiently handle tasks such as creating and managing users, setting up hosting plans, configuring backups, and editing OpenPanel settings.

The **OpenPanel** interface is the client-level panel where end-users can manage their environment, edit settings, install services, create websites, and more.

[![OpenPanel vs OpenAdmin](/website/static/img/admin/openpanel_vs_openadmin.svg)](https://openpanel.com/docs/admin/intro/)

---

## Supported Operating Systems

OpenPanel is a truly [OS-agnostic](https://www.techtarget.com/whatis/definition/agnostic) control panel.

| **Officially Supported OS** | **Version Details** | **Community Supported OS**                 |
|-----------------------------|---------------------|-------------------------------------------|
| Ubuntu 24.04 LTS            | Noble Numbat        | Ubuntu 22.04 LTS (Jammy Jellyfish) *(Looking for maintainers)* |
| Ubuntu 23.10                | Mantic Minotaur     |                                           |
| Debian 12                   | Bookworm            | Debian 11 (Bullseye) *(Looking for maintainers)* |
| Debian 13                   | Trixie (Testing)    |                                           |
| AlmaLinux 9                 | 9.0 - 9.3           | AlmaLinux 8.x *(Looking for maintainers)* |
| RockyLinux 9                | 9.0 - 9.3           | RockyLinux 8.x *(Looking for maintainers)* |
| CentOS Stream 9             | All point releases  | CentOS Stream 8 *(Looking for maintainers)* |
| Oracle Linux 9              | 9.0 - 9.3           |                                           |
| RHEL 9                      | 9.0 - 9.3           |                                           |
| Fedora 39                   | All updates         | Fedora 38 *(Looking for maintainers)*     |

### Unsupported Operating Systems

- **MacOS**: Explicitly checked and unsupported.
- **Containers**: Running inside Docker or LXC containers is not supported.

### Architecture Support

- **x86_64**: Fully supported.
- **aarch64**: Supported with specific configurations (e.g., branch selection for OpenAdmin).

For more details on supported operating systems and configurations, please refer to the [installation guide](https://openpanel.com/install).

---

## What Do You Mean by 'Open'?

Open for business!

OpenPanel is the first truly modular panel where [absolutely everything is customizable](https://openpanel.com/docs/articles/dev-experience/customizing-openpanel-user-interface/) and works independently of the rest of the system. It is also OS-agnostic and works the same on all supported systems.

However, OpenPanel is **not** an open-source project. The primary reason behind this decision is our commitment to maintaining the highest standards of security for our users, which we can only achieve with closed-source and controlled contributions.

While OpenPanel itself is not 100% open source, we are committed to transparency and security:

- [Who We Are](https://openpanel.com/about)
- [What We Are Planning](https://openpanel.com/roadmap)
- [What We Did So Far](https://openpanel.com/docs/changelog/intro/)

---

## Installation

### Self-Hosted VPS/Dedicated Server

```bash
bash <(curl -sSL https://openpanel.org)
```

For more details on configuring your server during installation, **please visit**: [Installation Guide](https://openpanel.com/install).

---

### 1-Click Deployment on DigitalOcean

[![Deploy on DigitalOcean](/website/static/img/do-btn-blue.svg)](https://marketplace.digitalocean.com/apps/openpanel?refcode=6498bfc47cd6&action=deploy)

#### DigitalOcean API Example

```bash
curl -X POST -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer '$TOKEN'' -d \
    '{"name":"choose_a_name","region":"nyc3","size":"s-2vcpu-4gb","image":"openpanel"}' \
    "https://api.digitalocean.com/v2/droplets"
```

---

## Support

Our [Community](https://community.openpanel.com/) serves as our virtual headquarters, where the community helps each other.

**Learn, share**, and **discuss** your questions with other community members.

---

## Troubleshooting

Here are some common issues and their solutions:

### Installation Issues

- **Error: "You must be root to execute this script."**
  - Ensure you are running the installation script as the root user or with `sudo`.

- **Error: "Unsupported Operating System."**
  - Verify that your OS is listed in the [Supported Operating Systems](#supported-operating-systems) section.

- **Error: "Running inside a container is not supported."**
  - OpenPanel does not support installation inside Docker or LXC containers. Use a dedicated VPS or server.

### Service Issues

- **OpenAdmin is not accessible on port 2087.**
  - Check if the service is running: `systemctl status admin`.
  - Ensure port 2087 is open in your firewall settings.

- **MySQL container fails to start.**
  - Verify that Docker is installed and running.
  - Check the logs for errors: `docker logs <container_id>`.

### General Issues

- **SSL certificate is not generated.**
  - Ensure your hostname is a valid FQDN and DNS is correctly configured.
  - Restart the Caddy service: `docker compose restart caddy`.

- **Quota setup fails.**
  - Ensure `usrquota` and `grpquota` are enabled in `/etc/fstab` and remount the filesystem.

For more help, visit our [Community Forums](https://community.openpanel.org/) or [Documentation](https://openpanel.com/docs/admin/intro/).
