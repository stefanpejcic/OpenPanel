---
sidebar_position: 3
---

# WordPress Manager

![Site Manager listing websites grouped by type with their version, creation date and PageSpeed scores](/img/openpanel-screenshots/applications/sites-list.png#gh-light-mode-only)
![Site Manager listing websites grouped by type with their version, creation date and PageSpeed scores](/img/openpanel-screenshots/applications/sites-list_dark.png#gh-dark-mode-only)

The WordPress Manager is your all-in-one tool inside OpenPanel for installing and managing WordPress websites — without ever needing to log in to wp-admin. It makes handling multiple sites fast, simple, and efficient.

## Manage WordPress sites

The WordPress Manager lets you adjust settings, create backups, update plugins, toggle debugging, and more — all directly from OpenPanel. No need to open multiple dashboards or remember dozens of logins.
Perfect for agencies, developers, and anyone managing several WordPress sites at once.

### WP Manager

On the main WP Manager page you can:

- [View installations](#wp-manager): see the domain, WordPress version, install date, admin email.
- [Refresh website data](#refresh-website-data): if you’ve changed a domain, updated WordPress manually, or modified the admin email.
- [Manage themes and plugins sets](#themes-and-plugins-sets): define which themes and plugins are auto-installed on every new site.
- [Install WordPress](#install-wordpress): set up a fresh WordPress installation in a few clicks.
- [Scan for existing installations](#scanning-importing-installations): detect and import manually installed WordPress sites.
- [Change to Table/Grid view](#grid-vs-table-view): display sites in either Grid (default) or Table mode.

### Install WordPress

Installing WordPress is quick and automatic. OpenPanel takes care of everything — downloading WordPress from WordPress.org, creating the database, linking it to your domain, and configuring your new site.

1. Add your **domain name** first.
2. Open **Websites** from the sidebar and click **+ New Website**.
3. Choose I**nstall WordPress**.

![Auto Installer page with cards for WordPress, Joomla, Drupal, Website Builder, PrestaShop, OpenCart and other applications](/img/openpanel-screenshots/applications/autoinstaller-page.png#gh-light-mode-only)
![Auto Installer page with cards for WordPress, Joomla, Drupal, Website Builder, PrestaShop, OpenCart and other applications](/img/openpanel-screenshots/applications/autoinstaller-page_dark.png#gh-dark-mode-only)

Then fill in the form:

- Website name
- Site description (optional)
- Domain name (optionally a subfolder)
- Admin username
- Admin password
- WordPress version


Click **Start Installation** and you’re done.

![Install WordPress form with site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/wp_install-form.png#gh-light-mode-only)
![Install WordPress form with site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/wp_install-form_dark.png#gh-dark-mode-only)

📘 Read the full guide: [How to Install WordPress® With OpenPanel](/docs/articles/websites/how-to-install-wordpress-with-openpanel/#install-wordpress-via-wp-manager)

### Scanning (Importing) Installations

If you already have WordPress installed manually, you can import it into the WP Manager.
The system scans your hosting files for `wp-config.php` and automatically adds the found websites.

![Scan for Existing Installations button next to New Installation on the WordPress Manager page](/img/openpanel-screenshots/applications/wp_manager-scan.png#gh-light-mode-only)
![Scan for Existing Installations button next to New Installation on the WordPress Manager page](/img/openpanel-screenshots/applications/wp_manager-scan_dark.png#gh-dark-mode-only)

📘 Read the full guide: [How to Migrate a WordPress® Installation to OpenPanel](/docs/articles/websites/how-to-upload-wordpress-website-to-openpanel/)

### Themes and Plugins Sets

Tired of installing the same setup every time?
Create **Theme Set** and **Plugin Set** that automatically apply to new WordPress installs.

![Themes and Plugins buttons for managing the sets that are installed on every new WordPress site](/img/openpanel-screenshots/applications/wp_manager-sets.png#gh-light-mode-only)
![Themes and Plugins buttons for managing the sets that are installed on every new WordPress site](/img/openpanel-screenshots/applications/wp_manager-sets_dark.png#gh-dark-mode-only)

For example, you might set up a default combo like:

- Elementor theme + child theme
- Elementor plugin
- Classic Editor plugin

Every time you install a new site — boom, it’s ready with your preferred setup.

📘 Read the full guide: [WordPress Plugin & Theme Sets in OpenPanel](/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/)

### Refresh Website Data

If you’ve made manual changes to your site (like updating WordPress core or changing the admin email), click **Refresh Data** to sync everything with WP Manager.

![Refresh Data button on the WordPress Manager page](/img/openpanel-screenshots/applications/wp_manager-refresh.png#gh-light-mode-only)
![Refresh Data button on the WordPress Manager page](/img/openpanel-screenshots/applications/wp_manager-refresh_dark.png#gh-dark-mode-only)

### Grid vs. Table View

You can view your sites in a **grid with screenshots** or a **simple table** view.
Switch views anytime using a button.

![Switch to Table view button on the WordPress Manager page](/img/openpanel-screenshots/applications/wp_manager-view.png#gh-light-mode-only)
![Switch to Table view button on the WordPress Manager page](/img/openpanel-screenshots/applications/wp_manager-view_dark.png#gh-dark-mode-only)

---

## Site Manager

![WordPress site manager with the screenshot, versions, files and database details](/img/openpanel-screenshots/applications/wordpress-site.png#gh-light-mode-only)
![WordPress site manager with the screenshot, versions, files and database details](/img/openpanel-screenshots/applications/wordpress-site_dark.png#gh-dark-mode-only)


### Auto Login to wp-admin

Use **Login as Admin** for one-click secure access to your WordPress dashboard — no password needed.

![Site header with the Live Preview and Login as Admin buttons](/img/openpanel-screenshots/applications/wordpress-header.png#gh-light-mode-only)
![Site header with the Live Preview and Login as Admin buttons](/img/openpanel-screenshots/applications/wordpress-header_dark.png#gh-dark-mode-only)

### Temporary Link

Preview your site even before your domain is connected or SSL is ready.
Temporary links last 15 minutes.

Click **Live Preview** to generate one:


### Screenshot

Website screenshots refresh automatically every 24 hours.
Need it sooner? Click the refresh icon over the screenshot.

### Versions

* **WordPress Version** – The WordPress version is retrieved from the database and verified via an AJAX request to the website itself, ensuring the displayed version is accurate. If an update is available, a badge will appear next to the version number.
* **PHP Version** – The PHP version is read from the domain’s VirtualHost configuration file, guaranteeing that the version shown matches the one actually configured for the domain.
* **MySQL/MariaDB Version** – Displays whether the site uses MySQL or MariaDB, along with the version number obtained directly from the terminal.
* **Created** – Indicates the date and time when the website was first added to WP Manager.

![WordPress, PHP and MariaDB version cards and the creation date](/img/openpanel-screenshots/applications/wordpress-versions.png#gh-light-mode-only)
![WordPress, PHP and MariaDB version cards and the creation date](/img/openpanel-screenshots/applications/wordpress-versions_dark.png#gh-dark-mode-only)

### Speed

Website performance is monitored daily using **Google PageSpeed Insights**. For both mobile and desktop devices, you can view the check time along with key metrics such as **First Contentful Paint**, **Speed Index**, and **Time to Interactive**.

You can also [add your own PageSpeed Insights API key](/docs/articles/websites/google-pagespeed-insights-api-key/#adding-the-api-key-in-openpanel) to customize the data collection.

![Speed card with desktop and mobile PageSpeed scores and First Contentful Paint, Speed Index and Time to Interactive](/img/openpanel-screenshots/applications/wordpress-speed.png#gh-light-mode-only)
![Speed card with desktop and mobile PageSpeed scores and First Contentful Paint, Speed Index and Time to Interactive](/img/openpanel-screenshots/applications/wordpress-speed_dark.png#gh-dark-mode-only)

### Safe Browsing

Checks your domain against the **Google Safe Browsing** API for malware, social engineering, unwanted software, and other flagged threats. Results are cached for 12 hours.

![Google Safe Browsing section of the Security tab with the result of the check against the Google Safe Browsing list](/img/openpanel-screenshots/applications/wordpress-safe-browsing.png#gh-light-mode-only)
![Google Safe Browsing section of the Security tab with the result of the check against the Google Safe Browsing list](/img/openpanel-screenshots/applications/wordpress-safe-browsing_dark.png#gh-dark-mode-only)

### Vulnerability Scan

Scans the site's WordPress core, plugin, and theme versions for known vulnerabilities. A fresh scan runs automatically if no cached report exists yet, or can be triggered manually.

![WP Vulnerabilities section of the Security tab with the number of detected vulnerabilities, the last check time and the Scan for vulnerabilities button](/img/openpanel-screenshots/applications/wordpress-vulnerabilities.png#gh-light-mode-only)
![WP Vulnerabilities section of the Security tab with the number of detected vulnerabilities, the last check time and the Scan for vulnerabilities button](/img/openpanel-screenshots/applications/wordpress-vulnerabilities_dark.png#gh-dark-mode-only)

### Cache

Cache widget displays the current [wp cache type](https://developer.wordpress.org/cli/commands/cache/type/) on your website and an option to purge the cache.

![Cache card with the cache type and the Clear Cache button](/img/openpanel-screenshots/applications/wordpress-cache.png#gh-light-mode-only)
![Cache card with the cache type and the Clear Cache button](/img/openpanel-screenshots/applications/wordpress-cache_dark.png#gh-dark-mode-only)

### Firewall

If CorazaWAF is enabled on the server, and your account has access to the WAF feature, you will see a *Firewall* widget displaying current status for the domain, an option to change it and number of denied/challenged requests in the last hour.

![Firewall card showing the firewall as active with denied and challenged request counts](/img/openpanel-screenshots/applications/wordpress-firewall.png#gh-light-mode-only)
![Firewall card showing the firewall as active with denied and challenged request counts](/img/openpanel-screenshots/applications/wordpress-firewall_dark.png#gh-dark-mode-only)


### Overview

Under *Overview* tab you can view:
- Files: Folder path and Folder Size
- Database: Size, Host, Name, Table Prefix, User, Password and link to open phpMyAdmin

![Files and Database cards of the Overview tab with the folder path and size, disk usage, and the database name, user, host, size and phpMyAdmin link](/img/openpanel-screenshots/applications/wordpress-files-database.png#gh-light-mode-only)
![Files and Database cards of the Overview tab with the folder path and size, disk usage, and the database name, user, host, size and phpMyAdmin link](/img/openpanel-screenshots/applications/wordpress-files-database_dark.png#gh-dark-mode-only)

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

![Options tab with the site URL, site name, email, registration, SEO visibility and pingback settings](/img/openpanel-screenshots/applications/wordpress-options.png#gh-light-mode-only)
![Options tab with the site URL, site name, email, registration, SEO visibility and pingback settings](/img/openpanel-screenshots/applications/wordpress-options_dark.png#gh-dark-mode-only)

### Maintenance mode

Enable or disable maintenance mode directly from WP Manager.
You can even edit the maintenance.php file right from the panel.

![Maintenance tab with the maintenance mode toggle](/img/openpanel-screenshots/applications/wordpress-maintenance.png#gh-light-mode-only)
![Maintenance tab with the maintenance mode toggle](/img/openpanel-screenshots/applications/wordpress-maintenance_dark.png#gh-dark-mode-only)

### Security

Keep your site safe with built-in security tools.

From here, you can:
- Shuffle WordPress salts
- Check core file integrity
- Reinstall WordPress core if needed

![Security tab with vulnerability report, Safe Browsing, salts, integrity check, malware scan and reinstall](/img/openpanel-screenshots/applications/wordpress-security.png#gh-light-mode-only)
![Security tab with vulnerability report, Safe Browsing, salts, integrity check, malware scan and reinstall](/img/openpanel-screenshots/applications/wordpress-security_dark.png#gh-dark-mode-only)

### Updates

Control how WordPress handles updates for the core, plugins, and themes.
By default, only minor core updates are auto-enabled.

![Updates tab listing core, plugin and theme update status](/img/openpanel-screenshots/applications/wordpress-updates.png#gh-light-mode-only)
![Updates tab listing core, plugin and theme update status](/img/openpanel-screenshots/applications/wordpress-updates_dark.png#gh-dark-mode-only)

If a newer WordPress core version is available, you will see 'Click to update WordPress core' button which when clicked will perform WordPress update to the newest version available.

### Debugging

Toggle WordPress’s built-in debugging tools (WP_DEBUG, WP_DEBUG_LOG, etc.) directly from WP Manager.

These are great for testing or development sites — not recommended for production.
For details, check [Debugging in WordPress](https://wordpress.org/documentation/article/debugging-in-wordpress/) for more information on these options.

![Debugging tab with toggles for WP_DEBUG, WP_DEBUG_LOG, WP_DEBUG_DISPLAY, SCRIPT_DEBUG and SAVEQUERIES](/img/openpanel-screenshots/applications/wordpress-debugging.png#gh-light-mode-only)
![Debugging tab with toggles for WP_DEBUG, WP_DEBUG_LOG, WP_DEBUG_DISPLAY, SCRIPT_DEBUG and SAVEQUERIES](/img/openpanel-screenshots/applications/wordpress-debugging_dark.png#gh-dark-mode-only)

### Backups

Create and restore backups anytime — files, database, or both.

Create a Backup:
- Choose what to back up (files, database, or both).
- Click **Generate Backup**.

![Backups tab with the Create a Backup and Restore from Backups sections](/img/openpanel-screenshots/applications/wordpress-backups.png#gh-light-mode-only)
![Backups tab with the Create a Backup and Restore from Backups sections](/img/openpanel-screenshots/applications/wordpress-backups_dark.png#gh-dark-mode-only)

Restore a Backup:
To restore, click Restore, pick a backup date, and confirm.

### Clone

Create a clone (copy files and database tables, replace links in database, fluch wp cache and rewrite rules).

Create a clone:
- Under 'Target' select the desired domain and optionally 'Database' to be used.
- Click **Clone Website**.

![Clone tab with the destination location and database for the copy](/img/openpanel-screenshots/applications/wordpress-clone.png#gh-light-mode-only)
![Clone tab with the destination location and database for the copy](/img/openpanel-screenshots/applications/wordpress-clone_dark.png#gh-dark-mode-only)

### Remove

Want to stop managing a site in WP Manager (without deleting it)?

Use **Detach** — your files and database remain untouched.


To completely remove a website — files, database, and all — click **Uninstall**, then confirm.

![Remove tab with the Detach and Uninstall options](/img/openpanel-screenshots/applications/wordpress-remove.png#gh-light-mode-only)
![Remove tab with the Detach and Uninstall options](/img/openpanel-screenshots/applications/wordpress-remove_dark.png#gh-dark-mode-only)
