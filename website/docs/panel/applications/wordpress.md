---
sidebar_position: 1
---

# WordPress Manager

![wp_manager_grid.png](/img/panel/v2/wpmanager.png)

The WordPress Manager is your all-in-one tool inside OpenPanel for installing and managing WordPress websites — without ever needing to log in to wp-admin. It makes handling multiple sites fast, simple, and efficient.

## Manage WordPress sites

The WordPress Manager lets you adjust settings, create backups, update plugins, toggle debugging, and more — all directly from OpenPanel. No need to open multiple dashboards or remember dozens of logins.
Perfect for agencies, developers, and anyone managing several WordPress sites at once.

### WP Manager overview

On the main WP Manager page you can:

- [View installations](#): see the domain, WordPress version, install date, admin email.
- [Refresh website data](/docs/panel/applications/wordpress#refresh-existing-data): if you’ve changed a domain, updated WordPress manually, or modified the admin email.
- [Manage themes and plugins sets](#themes-and-plugins-sets): define which themes and plugins are auto-installed on every new site.
- [Install WordPress](/docs/panel/applications/wordpress#install-wordpress): set up a fresh WordPress installation in a few clicks.
- [Scan for existing installations](/docs/panel/applications/wordpress#scan-import-installations): detect and import manually installed WordPress sites.
- [Change to Table/Grid view](/docs/panel/applications/wordpress#grid-vs-tabular-view): display sites in either Grid (default) or Table mode.

### Install WordPress

Installing WordPress is quick and automatic. OpenPanel takes care of everything — downloading WordPress from WordPress.org, creating the database, linking it to your domain, and configuring your new site.

1. Add your **domain name** first.
2. Open **Site Manager** from the sidebar and click **+ New Website**.
3. Choose I**nstall WordPress**.

![new_site_popup.png](/img/panel/v2/wpinstall.png)

Then fill in the form:

- Website name
- Site description (optional)
- Domain name (optionally a subfolder)
- Admin username
- Admin password
- WordPress version


Click **Start Installation** and you’re done.

![wp_install.png](/img/panel/v2/wpinstall2.png)

### Scanning (Importing) Installations

If you already have WordPress installed manually, you can import it into the WP Manager.
The system scans your hosting files for `wp-config.php` and automatically adds the found websites.

![wp_manager_scan.png](/img/panel/v2/wpscan.png)

### Themes and Plugins Sets

Tired of installing the same setup every time?
Create **Theme Set** and **Plugin Set** that automatically apply to new WordPress installs.

For example, you might set up a default combo like:

- Elementor theme + child theme
- Elementor plugin
- Classic Editor plugin

Every time you install a new site — boom, it’s ready with your preferred setup.

📘 Read the full guide: [WordPress Plugin & Theme Sets in OpenPanel](http://openpanel.com/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/)

### Refresh Website Data

If you’ve made manual changes to your site (like updating WordPress core or changing the admin email), click **Refresh Data** to sync everything with WP Manager.

![wp_manager_refresh_data.png](/img/panel/v2/wprefresh.png)

### Grid vs. Table View

You can view your sites in a **grid with screenshots** or a **simple table** view.
Switch views anytime using a button.

![wp_manager_table.png](/img/panel/v2/wptable.png)

---

## Managing a Website

![wp_manager_site.png](/img/panel/v2/wpmanage.png)

### General Information

Each website entry shows you:

- Site name & screenshot
- File path and folder size
- Database details (and size)
- Domain name
- WordPress version
- PHP & MySQL versions
- SSL status
- Quick links to phpMyAdmin, wp-admin login, PHP version settings

### View Database Info

Database credentials are displayed directly from wp-config.php file, click the blurred password to reveal it.
![wp_manager_site_database_password.png](/img/panel/v2/wpunblur.png)

### Detach a Website

Want to stop managing a site in WP Manager (without deleting it)?

Use **Detach** — your files and database remain untouched.

![wp_manager_site_detach_1.png](/img/panel/v2/wpdetach.png)

Confirm the action in the popup:

![wp_manager_site_detach_2.png](/img/panel/v2/wpdetach2.png)


### Auto Login to wp-admin

Use **Login as Admin** for one-click secure access to your WordPress dashboard — no password needed.

![wp_manager_site_login_admin.png](/img/panel/v2/wpautolog.png)


### Preview with Temporary Link

Preview your site even before your domain is connected or SSL is ready.
Temporary links last 15 minutes.

Click **Preview** to generate one:

![website_temporary_url_openpanel.gif](/img/panel/v2/wppreview.png)

### Edit WordPress Settings

CLick on the field to edit information, and click on Save.

#### General Settings

General settings that can be edited for a website:

- Website name
- Blog Description
- Enable/Disable user registrations
- Admin Email that is used for receiving all information
- Allow/Block pingbacks from other websites that mention you
- Block/Allow search engines like Google, Bing, etc.

![wp_manager_site_edit_1.png](/img/panel/v2/wpgeneral.png)

#### Update Preferences

Here you can set the update preferences for WordPress core, plugins and themes.

By default only WordPress core updates to minor versions are enabled.

![wp_manager_site_edit_2.png](/img/panel/v2/wpupdate.png)

#### Update WordPress core

If a newer WordPress core version is available, you will see 'Click to update' button which when clicked will perform WordPress update to the newest version available.

#### Debug Preferences

These options allow you to manage the native WordPress debugging tools, enabling and disabling these tools in the wp-config.php file. It is not recommended to use these options on production websites since they are meant for development and test installations. Refer to [Debugging in WordPress article](https://wordpress.org/documentation/article/debugging-in-wordpress/) for more information on these options.


Here you can enable:

- WP_DEBUG
- WP_DEBUG_LOG
- WP_DEBUG_DISPLAY
- SHOW_DEBUG
- SAVEQUERIES

![wp_manager_site_edit_3.png](/img/panel/v2/wpdebug.png)

### Refresh website screenshot

Website screenshots are periodically re-generated every 24h, if you need to manually refresh the screenshot click on the icon in top left corner of the screenshot.

![wp_manager_site_refresh_screenshot.png](/img/panel/v1/applications/wp_manager_site_refresh_screenshot.png)


### Uninstall WordPress

To uninstall WordPress and permanently delete all website files and database, click on the 'Uninstall' button.

![wp_manager_site_remove_1.png](/img/panel/v2/wpdetach.png)

On the popup click on 'Confirm Uninstall' button to confirm:

![wp_manager_site_remove_2.png](/img/panel/v2/wpdetach2.png)


### Backup and Restore
You have the options to perform manual backups of WordPress files or databases as needed, and easily restore them when required.

#### Create a backup

To generate a new backup click on the 'Backup' button.

![wp_manager_site_backup_1.png](/img/panel/v2/wpbackup.png)

On the modal select to backup both files and database, just a database or just files.
Click on the 'Run Backup' to start the backup process:

![wp_manager_site_backup_2.png](/img/panel/v1/applications/wp_manager_site_backup_2.png)

After backup process is finished you will receive a notification.

![wp_manager_site_backup_done.png](/img/panel/v1/applications/wp_manager_site_backup_done.png)

#### Restore from backup

To restore website from a backup created with OpenPanel WP Manager backup option simply click on the 'Restore' button for that website in WP Manager:

![wp_manager_site_restore_1.png](/img/panel/v1/applications/wp_manager_site_restore_1.png)

In the modal, select the backup date from which to restore the website. In the brackets next to each date you can view if the backup contains only database, only files or both. 

![wp_manager_site_restore_2.png](/img/panel/v1/applications/wp_manager_site_restore_2.png)

After selecting a date, confirm the restore process by clicking on the 'Confirm Restore (Click Again)' button.

![wp_manager_site_restore_confirm.png](/img/panel/v1/applications/wp_manager_site_restore_confirm.png)

When the restore process is complete you will receive a notification:

![wp_manager_site_restore_done.png](/img/panel/v1/applications/wp_manager_site_restore_done.png)


### Maintenance mode

Using WordPress manager you can enable and disable Maintenance mode for your website and also view the current website status regarding maintenance.

You can also begin editing the maintenance.php file through WordPress manager.

![wp_manager_maintenance.png](/img/panel/v2/wpmaint.png)

### Security

Within the security tab you can shuffle WordPress salts, check the integrity of WordPress core files and also reinstall the core files if needed.

![wp_manager_security.png](/img/panel/v2/wpsec.png)

