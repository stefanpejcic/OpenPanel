---
sidebar_position: 5
---

# Drupal Manager

Install and manage [Drupal](https://www.drupal.org/) sites in an existing domain, via [Composer](https://getcomposer.org/) and [Drush](https://www.drush.org/) — deliberately simpler than the WordPress Manager: no cloning, no filesystem scan for existing installs, no hardening rules, no drush passthrough console, and no dedicated backup system. Install, a read-only overview, and uninstall.

---

## Install Drupal

Navigate to **OpenPanel > AutoInstaller** and click **Install Drupal**, or go directly to **Drupal Manager** in the sidebar and click **Install Drupal**.

On the install page, configure:

* **Domain** – The domain (and optional subfolder) to install into. Leave the subfolder blank to install at the domain root.
* **Site name** – Displayed as the site's title.
* **Drupal version** – Latest, or pin to a specific major version (Drupal 11 or Drupal 10).
* **Admin username / password / email** – The initial administrator account Drupal creates during install. Leave the password blank to generate one.

Behind the scenes, this runs `composer create-project drupal/recommended-project`, requires `drush/drush`, creates a dedicated MySQL/MariaDB database, and runs `drush site:install standard` with the admin details you provided. Progress is streamed live as each step completes.

## Drupal Manager list

**Drupal Manager** in the sidebar lists every Drupal site on your account. Installations also show up on the general **Site Manager** page (`/sites`) alongside your other websites.

## Manage a Drupal site

Click **Manage** on a Drupal site (from either Site Manager or Drupal Manager) to open its overview page:

* **Docroot** – Where the Drupal codebase lives.
* **Drupal version** – The exact installed version, read live from `composer.lock`.
* **Database name / user / host** – Read live from `settings.php`, not stored anywhere separately.
* **Created** – When the site was installed.

### Remove

The **Remove** tab's **Delete Application** button fully uninstalls the site: drops the database and database user, deletes every file in the docroot, and removes it from Site Manager. This cannot be undone.

---

## Not included

To keep this feature simple, the following are **not** part of the Drupal Manager (unlike the WordPress Manager):

- Cloning a site
- Scanning the filesystem for untracked installations
- Security hardening rules
- A Drush command passthrough console
- A dedicated backup/restore system (use the account-level [Backups](/docs/panel/files/backups) feature instead)
- Only MySQL/MariaDB databases are supported (not PostgreSQL)
