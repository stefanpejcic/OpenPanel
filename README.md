<div align="center">
<a href="https://openpanel.com/">
    <img alt="website" src="/website/static/img/openpanel_admin.gif">
</a>
  
<br/>
<br/>

<div align="center">
    <a href="https://openpanel.com">Website</a> |
    <a href="https://community.openpanel.com/">Forums</a> |
    <a href="https://openpanel.com/docs/admin/intro/">Documentation</a>
</div>
</div>

<br/>
<br/>

<div align="center"> 
    
Deliver a <strong>VPS-like experience</strong> to your users at a fraction of the cost, with all-inclusive features such as <strong>resource limiting, user isolation, WP Manager, and enhanced security</strong> seamlessly integrated for worry-free hosting.

<br />
<br />

</div>
<br/>

## What is OpenPanel

OpenPanel is a powerful and flexible web hosting control panel for Linux systems. Available in an community-supported version, and a more feature-filled version with premium support, OpenPanel is the cost-effective and comprehensive solution to web hosting management.

**OpenPanel** offers a distinct advantage over other hosting panels by providing each user with an isolated environment and tools to fully manage it. This ensures that you enjoy full control over your environment, simillar to a VPS experience. You can effortlessly install new PHP versions, modify server configurations, view domain logs, restart services, and perform numerous other advanced tasks.

[![openpanel scheme](/website/static/img/admin/openpanel_scheme.png)](https://openpanel.com/docs/panel/intro/)

This panel is the culmination of years of experience in the hosting industry, having spent decades working with various hosting panels we made sure to include all features that simply make sense.

When we designed OpenPanel, we prioritized features that are not only user-friendly for beginners but also advanced enough to alleviate maintenance tasks for system administrators and hosting support teams.

Some of the features worth mentioning are:

- Users can run [Nginx or Apache webserver](https://openpanel.com/docs/admin/plans/hosting_plans/#list-hosting-plans).
- Users can run MySQL or MariaDB
- Users can [install PHP versions](https://openpanel.com/docs/panel/advanced/server_settings#install-php-version) they need, [edit php.ini files](https://openpanel.com/docs/panel/advanced/server_settings#phpini-editor) and set desired limits.
- Control [MySQL settings](https://openpanel.com/docs/panel/advanced/server_settings#mysql-settings), set limits, [enable remote mysql access](https://openpanel.com/docs/panel/databases/remote) and much more.
- [Update system services](https://openpanel.com/docs/panel/advanced/server_settings#service-status) and even install new services that they need.
- Manage WordPress websites easily with [WP Manager](https://openpanel.com/docs/panel/applications/wordpress).
- Password-less login to [phpMyAdmin](https://openpanel.com/docs/panel/databases/phpmyadmin) and [Web Terminal](https://openpanel.com/docs/panel/advanced/terminal).
- Built-in [REDIS](https://openpanel.com/docs/panel/caching/Redis) and [Memcached](https://openpanel.com/docs/panel/caching/Memcached) object caching.

And unique features that simply made sense 💁 to us:
- User and admin panels are completelly isolated
- Services auto-start only when needed so resources are not wasted
- Gooogle PageSpeed data is automatically displayed for every website in Site Manager
- Users can export DNS zones easily
- Users can pause cronjobs when not needed
- Administrators can receive daily usage reports
- Users can add comments for DNS records
- Download files from URL in File Manager
- Users can save pages to Favorites
- Users can share web terminal session with third-parties
- Detailed activity log of all actions
- Admins can add custom message per user
- and a lot more 🙌

## OpenPanel vs OpenAdmin


The **OpenAdmin** offers an administrator-level interface where you can efficiently handle tasks such as creating and managing users, setting up hosting plans, configuring backups, and editing OpenPanel settings.

The **OpenPanel** interface is the client-level panel where end-users can manage their enviroment, edit settings, install services, create websites and more.

[![openpanel-vs-openadmin](/website/static/img/admin/openpanel_vs_openadmin.svg)](https://openpanel.com/docs/admin/intro/)

## Community VS Enterprise 

OpenPanel is available in two editions:

- **[Community Edition](https://openpanel.com/product/openpanel-free-control-panel/)** is a free hosting control panel for up to 3 user accounts, suitable for VPS and private use.
- **[Enterprise Edition](https://openpanel.com/product/openpanel-premium-control-panel/)** is a premium version that offers advanced features for user isolation and management, suitable for web hosting providers.

## Supported OS


Supported OS:

OpenPanel is a truly [OS-agnostic](https://www.techtarget.com/whatis/definition/agnostic) control panel. Currently supported:

- AlmaLinux 9
- RockyLinux 9
- Fedora 40
- CentOS 9.4
- Debian 11 12
- Ubuntu 22.04 24.04

## What do you mean by "open" ?

Open for business!

OpenPanel is the first truly modular panel where absolutely everything is customizable and works independently of the rest of the system. It is also OS-agnostic and works the same on all supported systems.

However, OpenPanel is **not** an open-source project. The primary reason behind this decision is our commitment to maintaining the highest standards of security for our users, which we can only achive with closed-source and controlled contributions.

While OpenPanel itself is not 100% open source, we are committed to transparency and security:

- [Who are we](https://openpanel.com/about)
- [What are we planning](https://openpanel.com/roadmap)
- [What we did so far](https://openpanel.com/docs/changelog/intro/)

## Installation

To install on self-hosted VPS/Dedicated server: 

```bash
bash <(curl -sSL https://openpanel.org)
```

Sometimes, you might encounter issues when the images (e.g. Docker) were already installed on your server. To prevent such conflicts, it is recommended to use the `--skip-images` flag.

```bash
bash <(curl -sSL https://openpanel.org) --skip-images
```

To see more details to configure server on installation, **please visit**: https://openpanel.com/install 

----

Spin a 1-click droplet on DigitalOcean: 

[![droplet](/website/static/img/do-btn-blue.svg)](https://marketplace.digitalocean.com/apps/openpanel?refcode=6498bfc47cd6&action=deploy)


DigitalOcean API:
```bash
curl -X POST -H 'Content-Type: application/json' \
     -H 'Authorization: Bearer '$TOKEN'' -d \
    '{"name":"choose_a_name","region":"nyc3","size":"s-2vcpu-4gb","image":"openpanel"}' \
    "https://api.digitalocean.com/v2/droplets"
```

## Support

Our [Community](https://community.openpanel.com/) serves as our virtual Headquater, where the community helps each other.

**Learn, share** and **discuss** with other community members your questions.
