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

Administrators have the ability to enable or disable each feature (page) in the OpenPanel interface. To activate a feature, select it in the "Enable Features" section and click save. The change is immediate and necessitates the restart of the OpenPanel service to implement the modifications.

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


### disk_usage
### inodes
### usage
### terminal
### services
### webserver
### fix_permissions
### process_manager
### ip_blocker
### redis
### memcached
### login_history
### activity
### twofa
### domains_visitors

### FTP

FTP feature [adds Files > FTP page](/docs/panel/files/FTP/) which allows users to create and manage FTP accounts.

### Flarum

Flarum feature [adds Flarum to the AutoInstaller](/docs/panel/applications/) and allows users to manage WordPress websites from the SiteManager interface.

### malware_scan

malware_scan feature adds [Malware Scanner](/docs/panel/files/malware-scanner/) page to the Files section, where user can scan files using the [ClamAV®](https://www.clamav.net/) scanner.

### elasticsearch





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
