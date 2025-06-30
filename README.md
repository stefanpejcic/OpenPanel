## What is OpenPanel

OpenPanel is *probably* the most customizable web hosting control panel.

Available in an community-supported version, and a more feature-filled version with premium support, OpenPanel is the cost-effective and comprehensive solution to web hosting management.


## Why use OpenPanel to host websites?

**OpenPanel** offers a distinct advantage over other hosting panels by providing each user with an isolated environment and tools to fully manage it. This ensures that your users enjoy full control over their environment, simillar to a VPS experience. They can effortlessly run multiple PHP versions, modify server configurations, view domain logs, restart services, set limits, configure backups and perform numerous other advanced tasks.

- all domains have automatic SSL.
- WordPress autoinstaller and manager.
- nodejs and python apps
- standard stuff: DNS, emails, crons, phpMyAdmin, FTP..

## Why use OpenPanel for your hosting business?

- focus on security.
- WHMCS, FOSSBilling, Blesta, paymenter.org integrations.
- dedicated MySQL/MariaDB per user.
- detailed activity log of all user actions.
- low maintenance: each user manages their own services and backups.
- import cpanel accounts.
- white label


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

## Contributing

We welcome and appreciate all contributions - technical or not!

You don’t need to be a developer to make a meaningful impact.
Plese see [CONTRIBUTING.md](https://github.com/stefanpejcic/OpenPanel/blob/main/CONTRIBUTING.md)


## Repo Activity

![Alt](https://repobeats.axiom.co/api/embed/9904d020c32812f0aff8d8d69f52643d16f85007.svg "Repobeats analytics image")

## Copyright & license

- OpenAdmin and OpenPanel UI are distributed under EULA.
- OpenCLI and configuration files are distributed under Commons Attribution-NonCommercial (CC BY-NC) license.
