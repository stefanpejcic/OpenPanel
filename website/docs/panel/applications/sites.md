---
sidebar_position: 2.5
---

# Site Manager

The **Site Manager** (`/sites`) lists every website on your account in one place — static sites, WordPress installs, Node.js/Python/Ruby/Java applications, and sites created via other installers (Drupal, Joomla, OpenCart, etc.) — grouped by type.

To create a new site, click **+ New Website**, which opens the [Auto Installer](/docs/panel/applications/autoinstaller/).

## Table

Each row shows:

- **Type** – The site's application/framework (WordPress, static HTML, Node.js, Drupal, etc.).
- **Site Name** – Domain (or subfolder) the site is installed on, with its favicon.
- **Created** – Date the site was added to Site Manager.
- **Version** – The installed application version, when applicable. If a newer version is available upstream, an **Update available** badge appears next to it (checked for WordPress, Drupal, Joomla, and OpenCart sites).
- **Speed** – A link to check the site on Google PageSpeed Insights.
- **Actions** – Per-row buttons, shown depending on the site type:
  - **Login as Admin** – One-click login to the application's admin area (WordPress sites).
  - **Manage** (files icon) – Opens the domain's files in [File Manager](/docs/panel/files/files/).
  - **Manage** (site icon) – Opens the site's management page (the deeper per-site tabs, e.g. the [WordPress Manager](/docs/panel/applications/wordpress/#site-manager) tabs for WordPress sites), or the [Website Builder](/docs/panel/applications/builder/) editor for builder sites.
  - **Clone** – Copies the site's files and database to another domain.
  - **Backups** – Jumps to the site's backup tab.
  - **Delete** – Removes the site.

Use the **Table/Grid** toggle and column sort headers (Type, Site Name, Created, Version) to organize the list.

## Scanning for Existing Installations

If an application was installed manually (outside Site Manager), click **Scan** to detect it. The scan looks for installation files across every supported application type on disk, repairs the database host in the detected config if it still points to `localhost` instead of the actual database container, verifies the database connection, and imports any installation it finds as a new Site Manager entry.

## Bulk Actions

Select multiple sites using the checkboxes to:

- **Update** – Update the selected sites to the latest available version (where supported).
- **Detach** – Remove the sites from Site Manager only. Files and databases are left untouched.
- **Delete** – Permanently remove the selected sites, including their files and databases.

Each bulk action asks for confirmation before running.
