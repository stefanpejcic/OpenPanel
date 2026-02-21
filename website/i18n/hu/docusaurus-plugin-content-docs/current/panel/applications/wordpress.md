---
sidebar_position: 1
---

# WordPress Manager

![wp_manager_grid.png](/img/panel/v2/wpmanager.png)

The WordPress Manager is your all-in-one tool inside OpenPanel for installing and managing WordPress websites — without ever needing to log in to wp-admin. It makes handling multiple sites fast, simple, and efficient.

## Manage WordPress sites

The WordPress Manager lets you adjust settings, create backups, update plugins, toggle debugging, and more — all directly from OpenPanel. No need to open multiple dashboards or remember dozens of logins.
Perfect for agencies, developers, and anyone managing several WordPress sites at once.

### WP Manager

On the main WP Manager page you can:

- [View installations](#wp-manage): see the domain, WordPress version, install date, admin email.
- [Refresh website data](#refresh-website-data): if you’ve changed a domain, updated WordPress manually, or modified the admin email.
- [Manage themes and plugins sets](#themes-and-plugins-sets): define which themes and plugins are auto-installed on every new site.
- [Install WordPress](#install-wordpress): set up a fresh WordPress installation in a few clicks.
- [Scan for existing installations](#scanning-importing-installations): detect and import manually installed WordPress sites.
- [Change to Table/Grid view](#grid-vs-table-view): display sites in either Grid (default) or Table mode.

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

📘 Read the full guide: [How to Install WordPress® With OpenPanel](/docs/articles/websites/how-to-install-wordpress-with-openpanel/#install-wordpress-via-wp-manager)

### Scanning (Importing) Installations

If you already have WordPress installed manually, you can import it into the WP Manager.
The system scans your hosting files for `wp-config.php` and automatically adds the found websites.

📘 Read the full guide: [How to Migrate a WordPress® Installation to OpenPanel](/docs/articles/websites/how-to-upload-wordpress-website-to-openpanel/)

### Themes and Plugins Sets

Tired of installing the same setup every time?
Create **Theme Set** and **Plugin Set** that automatically apply to new WordPress installs.

For example, you might set up a default combo like:

- Elementor theme + child theme
- Elementor plugin
- Classic Editor plugin

Every time you install a new site — boom, it’s ready with your preferred setup.

📘 Read the full guide: [WordPress Plugin & Theme Sets in OpenPanel](/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/)

### Refresh Website Data

If you’ve made manual changes to your site (like updating WordPress core or changing the admin email), click **Refresh Data** to sync everything with WP Manager.

### Grid vs. Table View

You can view your sites in a **grid with screenshots** or a **simple table** view.
Switch views anytime using a button.

---

## Site Manager

![wp_manager_site.png](/img/panel/v2/wpmanage.png)


### Auto Login to wp-admin

Use **Login as Admin** for one-click secure access to your WordPress dashboard — no password needed.

![wp_manager_autologin](/img/panel/v2/wpautolog.png)

### Temporary Link

Preview your site even before your domain is connected or SSL is ready.
Temporary links last 15 minutes.

Click **Live Preview** to generate one:

![website_temporary_url_openpanel.gif](/img/panel/v2/wppreview.png)

### Screenshot

Website screenshots refresh automatically every 24 hours.
Need it sooner? Click the refresh icon over the screenshot.

### Versions

* **WordPress Version** – The WordPress version is retrieved from the database and verified via an AJAX request to the website itself, ensuring the displayed version is accurate. If an update is available, a badge will appear next to the version number.
* **PHP Version** – The PHP version is read from the domain’s VirtualHost configuration file, guaranteeing that the version shown matches the one actually configured for the domain.
* **MySQL/MariaDB Version** – Displays whether the site uses MySQL or MariaDB, along with the version number obtained directly from the terminal.
* **Created** – Indicates the date and time when the website was first added to WP Manager.

![general](/img/panel/v2/general.png)

### Speed

Website performance is monitored daily using **Google PageSpeed Insights**. For both mobile and desktop devices, you can view the check time along with key metrics such as **First Contentful Paint**, **Speed Index**, and **Time to Interactive**.

You can also [add your own PageSpeed Insights API key](/docs/articles/websites/google-pagespeed-insights-api-key/#adding-the-api-key-in-openpanel) to customize the data collection.

![speed](/img/panel/v2/speed.png)

### Cache

Cache widget displays the current [wp cache type](https://developer.wordpress.org/cli/commands/cache/type/) on your website and an option to purge the cache.

![wp_cache](/img/panel/v2/wp_cache.png)

### Firewall

If CorazaWAF is enabled on the server, and your account has access to the WAF feature, you will see a *Firewall* widget displaying current status for the domain, an option to change it and number of denied/challenged requests in the last hour.

![wp_waf](/img/panel/v2/wp_waf.png)


### Overview

Under *Overview* tab you can view:
- Files: Folder path and Folder Size
- Database: Size, Host, Name, Table Prefix, User, Password and link to open phpMyAdmin

![overview](/img/panel/v2/overview.png)

### Options

*Options* tab displays current WordPress settings and allows you to change them.

Available options:

- Site URl
- Homepage URL
- Site Name
- Blog Description
- Administrator Email
- Enable New User Registration
- Enable SEO Visibility
- Enable Pingbacks

![options](/img/panel/v2/options.png)

### Maintenance mode

Enable or disable maintenance mode directly from WP Manager.
You can even edit the maintenance.php file right from the panel.

![wp_manager_maintenance](/img/panel/v2/wpmaint.png)

### Security

Keep your site safe with built-in security tools.

From here, you can:
- Shuffle WordPress salts
- Check core file integrity
- Reinstall WordPress core if needed

![wp_manager_security.png](/img/panel/v2/wpsec.png)

### Updates

Control how WordPress handles updates for the core, plugins, and themes.
By default, only minor core updates are auto-enabled.

![wp_manager_site_edit_2.png](/img/panel/v2/wpupdate.png)

If a newer WordPress core version is available, you will see 'Click to update WordPress core' button which when clicked will perform WordPress update to the newest version available.

### Debugging

Toggle WordPress’s built-in debugging tools (WP_DEBUG, WP_DEBUG_LOG, etc.) directly from WP Manager.

These are great for testing or development sites — not recommended for production.
For details, check [Debugging in WordPress](https://wordpress.org/documentation/article/debugging-in-wordpress/) for more information on these options.

![wp_manager_site_edit_3.png](/img/panel/v2/wpdebug.png)

### Backups

Create and restore backups anytime — files, database, or both.

Create a Backup:
- Choose what to back up (files, database, or both).
- Click *Generate Backup**.

![wp_manager_site_backup_1.png](/img/panel/v2/wpbackup.png)

Restore a Backup:
To restore, click Restore, pick a backup date, and confirm.

### Remove

Want to stop managing a site in WP Manager (without deleting it)?

Use **Detach** — your files and database remain untouched.

![detach](/img/panel/v2/detach.png)

To completely remove a website — files, database, and all — click **Uninstall**, then confirm.

![uninstall](/img/panel/v2/uninstall.png)
