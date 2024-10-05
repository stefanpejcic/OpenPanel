---
sidebar_position: 2
---

# OpenPanel

Edit nameservers, disable features and more.

![openadmin openpanel settings](/img/admin/adminpanel_openpanel_settings.png)

The OpenPanel Settings page allows you to edit settings and features available to users in their OpenPanel interface.

## Branding

To set a custom name visible in the OpenPanel sidebar and on login pages, enter the desired name in the "Brand name" option. Alternatively, to display a logo instead, provide the URL in the "Logo image" field and save the changes.

## Set nameservers

Before adding any domains its important to first create nameservers so that added domains will have valid dns zone files and be able to propagate.

Configuring nameservers involves two steps:

1. Create private nameservers (glue DNS records) for the domain through your domain registry.
2. Add the nameservers into the OpenPanel configuration.

Here are tutorials for some popular domain providers:
- [Cloudflare](https://developers.cloudflare.com/dns/additional-options/custom-nameservers/zone-custom-nameservers/)
- [GoDaddy](https://uk.godaddy.com/help/add-custom-hostnames-12320)
- [NameCheap](https://www.namecheap.com/support/knowledgebase/article.aspx/768/10/how-do-i-register-personal-nameservers-for-my-domain/#:~:text=Click%20on%20the%20Manage%20option,5.)

To add nameservers from OpenAdmin navgiate to Settings > OpenPanel and set nameservers in ns1 and ns2 fields and click on save:

![openpanel add nameservers](/img/admin/openadmin_add_ns.png)

Or from terminal run commands:
```bash
opencli config update ns1 your_ns1.domain.com
opencli config update ns2 your_ns2.domain.com
```

:::info
After creating nameservers it can take up to 12h for the records to be globally accessible. Use a tool sush as [whatsmydns.net](https://www.whatsmydns.net/) to monitor the status.

If you still experience problems after the propagation process, then please check this guide: [dns server not responding to reqeuests](https://community.openpanel.co/d/5-dns-server-does-not-respond-to-request-for-domain-zone).
:::


## Enable Features

Administrators have the ability to enable or disable each feature (page) in the OpenPanel interface. To activate a feature, select it in the "Enable Features" section and click save. The change is immediate and restarts OpenPanel service.

Once enabled, the feature becomes instantly available to all users, appearing in the OpenPanel interface sidebar, search results, and dashboard icons.

![openpanel enable modules](/img/admin/openpanel_settings_modules.png)

### DNS
DNS feature enables the [Domains > Edit DNS Zone](/docs/panel/domains/dns) feature for users to edit DNS records.

### Favorites
Favorites allows users to bookmark pages to [Favorites](/docs/panel/dashboard/#favorites).

### phpMyAdmin
The phpMyAdmin feature allows users to access [phpMyAdmin](/docs/panel/databases/phpmyadmin/) with automatic login directly from the OpenPanel interface. When enabled, links to phpMyAdmin are added for each website in the [SiteManager](/docs/panel/applications/) and for every database on the [Databases](/docs/panel/databases/) page.

### temporary_links
The temporary_links feature adds a 'Preview' button to websites in the SiteManager allowing users to [preview their site using a temporary openpanel.org subdomain for 15 minutes](/docs/panel/applications/wordpress/#preview-with-temporary-link). This is useful for testing when the domain's DNS is not yet pointed to the server or when the domain lacks an SSL certificate.

### SSH
SSH feature enables the [SSH](//docs/panel/advanced/ssh/) page where users can enable/disable remote SSH access, view ssh port and change root password.

### Crons
Crons feature enables the [Cronjobs](/dodocs/panel/advanced/cronjobs/) page where users can edit cronjobs via interface or directly edit the crontab file via text editor.

### Backups
Backups feature enables users to access the [Files > Backups](/docs/panel/files/backups/) page where they can view current backups, restore or download them.

### WordPress
WordPress feature [adds WordPress to the AutoInstaller](/docs/panel/applications/) and allows users to manage WordPress websites using the [SiteManager interface](/docs/panel/applications/wordpress/#wp-manager-overview). 

### PM2
PM2 feature [adds NodeJS and Python Applications to the AutoInstaller](//docs/panel/applications/pm2/) and allows users to manage Python/NodeJS websites using the SiteManager interface.

### Mautic
Mautic feature [adds Mautic to the AutoInstaller](/docs/panel/applications/) and allows users to manage Mautic websites from the SiteManager interface.

### Flarum
Flarum feature [adds Flarum to the AutoInstaller](/docs/panel/applications/) and allows users to manage WordPress websites from the SiteManager interface.

### Redis
Redis feature enables the [REDIS page](/docs/panel/caching/Redis/) where users can view REDIS port (`6379`), view logs, set memory limits and enable/disable service.

### Memcached
Memcached feature enables the [Memcached page](/docs/panel/caching/Memcached/) where users can view Memcached port (`11211`), view logs, set memory limits and enable/disable service.

### ElasticSearch
Memcached feature enables the [ElasticSearch page](/docs/panel/caching/elasticsearch/) where users can view WlasticSearch port (`9200`), view logs, set memory limits and enable/disable service.

### Resource Usage
usage feature enables the [Resource Usage page](/docs/panel/analytics/resource_usage/) where users can view real-time information about the server's CPU and RAM usage.

### WebTerminal
terminal feature enables the [WebTerminal page](/docs/panel/advanced/terminal/) where users can open web terminal or create a temporary session for 3rd parties.

### Services
Services feature enables the [Server > Services Status page](docs/panel/advanced/server_settings/#service-status) where users can view their installed services, their current status and versions.

### WebServer Configuration Editor
webserver feature enables the [Server > Nginx/Apache Configuration Editor page](/docs/panel/advanced/server_settings/#nginx--apache-settings) where users can edit the main configuration file for their webserver. Users with Nginx can edit nginx.conf file and users with Apache webserver edit the apache2.conf file.

### Process Manager
process_manager feature enables the  [Advanced > Process Manager](/docs/panel/advanced/process_manager/) page where users can view all running processes and kill (force stop the process).

### IP Blocker
ip_blocker feature enables the [IP Blocker](/#) page where users can block IP addresses per domain name.

### Login History
login_history feature enables the [Account > Login History page](/docs/panel/account/login_history/) where users can view up to 10 last logins.

### Activity Log
activity feature enables the [Account > Activity page](/docs/panel/analytics/account_activity/) where users can view their activity log.

### 2FA
twofa feature enables the [Account > Two-Factor Authentication page](/docs/panel/account/2fa/) where users can enable or disable 2FA for enhanced security.

### Domain Visitors
domains_visitors feature enables the [Analytics > Domain Visitors page](/docs/panel/account/2fa/) where users can view generated HTML reports from their domains access logs.

### Disk Usage
disk_usage feature enables the [Files > Disk Usage page](//docs/panel/files/disk_usage/) where users can view disk usage per directory.

### Inodes Explorer
inodes feature enables the [Files > Inodes Explorer page](//docs/panel/files/inodes_explorer/) where users can view their number of inodes per directory.

### Fix Permissions
fix_permissions feature enables the [Files > Fix Permissions page](/docs/panel/files/fix_permissions/) where users can fix permissions and ownership issues per directory.

### FTP
FTP feature [adds Files > FTP page](/docs/panel/files/FTP/) which allows users to create and manage FTP accounts.

### Malware Scanner
malware_scan feature adds [Malware Scanner](/docs/panel/files/malware-scanner/) page to the Files section, where user can scan files using the [ClamAV®](https://www.clamav.net/) scanner.



## Other settings

Additional settings available in the Settings > OpenPanel page include:

- **Logout URL:** Set the URL for redirecting users upon logout from the OpenPanel.
- **Avatar Type:** Choose to display Gravatar, Letter, or Icon as avatars for users.
- **Resource Usage Charts:** Opt to display 1, 2, or no charts on the Resource Usage page.
- **Default PHP Version:** Specify the default PHP version for domains added by users (users can override this setting).
- **Enable Password Reset:** Activate password reset on login forms (not recommended).
- **Display 2FA Nag:** Show a message in users' dashboards encouraging them to set up 2FA for added security.
- **Display How-to Guides:** Display how-to articles for users in their dashboard pages.
- **Login Records:** Set the number of login records to keep for each user.
- **Activities per Page:** Specify the number of activity items to display per page.
- **Usage per Page:** Specify the number of Resource Usage items to display per page.
- **Usage Retention:** Set the number of Resource Usage items to keep for each user.
- **Domains per Page:** Specify the number of domains to display per page.
