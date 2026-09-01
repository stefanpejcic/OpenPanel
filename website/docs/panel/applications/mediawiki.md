---
sidebar_position: 11
---

# MediaWiki Manager

Install and manage [MediaWiki](https://www.mediawiki.org/) wikis in an existing domain, via MediaWiki's own non-interactive CLI installer — deliberately simpler than the WordPress Manager: no dedicated MediaWiki Manager sidebar page, no scanning the filesystem for existing installs, and no hardening rules. Install, a read-only overview with live status, error logs, one-time admin login, cloning, a full backup/restore system, and self-update.

---

## Install MediaWiki

Navigate to **OpenPanel > AutoInstaller** and click **Install MediaWiki**.

On the install page, configure:

* **Domain** – The domain (and optional subfolder) to install into. Leave the subfolder blank to install at the domain root.
* **MediaWiki version** – Latest, or pin to a specific release, fetched live by scraping [releases.wikimedia.org's directory listing](https://releases.wikimedia.org/mediawiki/) for release branches (e.g. `1.42/`) and each branch's patch-level tarballs (e.g. `mediawiki-1.42.7.tar.gz`).
* **Site name** – The wiki's display name.
* **Admin username / password / email** – The initial administrator account MediaWiki creates during install. Leave the password blank to generate one.

Behind the scenes, this downloads the matching `mediawiki-<version>.tar.gz` packaged release, extracts it flat into the docroot, creates a dedicated MySQL/MariaDB database, and runs MediaWiki's own `maintenance/install.php` non-interactive CLI installer against it. The install is blocked with a clear error up front if the domain's PHP version is older than the release's own minimum (read from its `composer.json` — MediaWiki's minimum PHP requirement climbs across branches). A per-minute cron job running `maintenance/runJobs.php` is registered automatically, since MediaWiki's job queue (link tables, search index, email) does nothing without it. Progress is streamed live as each step completes.

## Manage a wiki

Every MediaWiki install shows up on the general **Site Manager** page (`/sites`) alongside your other websites. Click **Manage** on a MediaWiki site to open its overview page:

* **Screenshot** and **MediaWiki / PHP / MariaDB / Created** status cards, matching the other CMS managers.
* **Files** – Docroot path, a File Manager shortcut, and live folder size.
* **Database** – Host, name, user, and prefix (read live from `LocalSettings.php`, not stored anywhere separately), plus a phpMyAdmin shortcut.
* **PHP version** – Editable for root-level installs; subdirectory installs inherit the domain's PHP version.

### Login as Admin

The **Login as Admin** button generates a one-time login link and opens the wiki already authenticated as an administrator, without needing the admin password. MediaWiki core has no CLI command for this (unlike Drupal's `drush uli`), so OpenPanel deploys a small login helper file (`openpanel-login.php`) into the docroot at install time; each click looks up an active sysop account, issues a fresh token valid for 10 minutes and usable once, and the helper binds that admin user to the browser's session via MediaWiki's own `User::setCookies()` API.

### Logs

The **Logs** tab shows the tail (last 300 lines) of the PHP-FPM container's most recently modified `logs/*.log` file under the docroot — MediaWiki writes no application log file by default unless debug logging is explicitly configured.

### Update

The **Update** tab checks the latest available version and, with one click, downloads and replaces the core files (leaving `LocalSettings.php` and the `images/` upload directory untouched) then runs `maintenance/update.php --quick` — MediaWiki's own documented manual-update procedure.

### Clone

The **Clone** tab copies a wiki — files and database — to another domain or subfolder. You can optionally set the destination database name, user, and password; otherwise they're auto-generated. Cloning copies the docroot, dumps and imports the database, rewrites `$wgDBname`/`$wgDBuser`/`$wgDBpassword`/`$wgServer`/`$wgScriptPath` in the new `LocalSettings.php` (MediaWiki hardcodes its base URL there, unlike Drupal), registers the clone's own per-minute job-queue cron job, and runs a search-and-replace across the database for any remaining hardcoded references to the old domain in page content.

### Backups

The **Backups** tab generates on-demand backups (database, files, or both) into a timestamped folder, and lists existing backups for one-click restore. A files backup is a `tar.gz` of the docroot; a database backup is a plain SQL dump, resolved from `LocalSettings.php`. There's no scheduled/automatic backup — each one is triggered manually.

### Remove

The **Remove** tab's **Delete Application** button fully uninstalls the site: drops the database, deletes every file in the docroot, removes the per-minute job-queue cron job, and removes it from Site Manager. This cannot be undone.

---

## Not included

To keep this feature simple, the following are **not** part of MediaWiki support (unlike the WordPress Manager):

- A dedicated "MediaWiki Manager" sidebar page — manage installed sites from Site Manager instead
- Scanning the filesystem for untracked installations
- Security hardening rules
- Only MySQL/MariaDB databases are supported
