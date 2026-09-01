---
sidebar_position: 16
---

# TinyPhotoGallery Manager

Install and manage [TinyPhotoGallery](https://github.com/stefanpejcic/tinyphotogallery) in an existing domain — the simplest of OpenPanel's app managers. TinyPhotoGallery is a single PHP file plus an empty `photos/` folder: no database, no admin account, no CLI installer, no versioning, and no update mechanism upstream. There is no dedicated TinyPhotoGallery Manager sidebar page, no cloning, no scanning for existing installs, no hardening rules, no maintenance mode, no admin auto-login, and no cache to clear.

---

## Install TinyPhotoGallery

Navigate to **OpenPanel > AutoInstaller** and click **Install TinyPhotoGallery**.

On the install page, configure:

* **Domain** – The domain (and optional subfolder) to install into. Leave the subfolder blank to install at the domain root.

Behind the scenes, this downloads `index.php` from the project's main branch (there are no tagged releases upstream) and creates an empty `photos/` folder next to it — the entire upstream install procedure per the project's own README. There's nothing else to configure: install is complete the moment those two filesystem items exist. Progress is streamed live as each step completes.

## Manage a TinyPhotoGallery site

Every TinyPhotoGallery install shows up on the general **Site Manager** page (`/sites`) alongside your other websites. Click **Manage** on a TinyPhotoGallery site to open its overview page:

* **Screenshot** and **TinyPhotoGallery / PHP / Created** status cards. TinyPhotoGallery always reports as "main" since there is no real versioning upstream.
* **Files** – Docroot path, a File Manager shortcut, and live folder size. There is no Database card — TinyPhotoGallery has no database.
* **PHP version** – Editable for root-level installs; subdirectory installs inherit the domain's PHP version.
* **Setup** – Shown until at least one photo exists in `photos/`, explaining that uploading photos there is all that's needed, with a link to open the site.

### Backups

The **Backups** tab generates and restores files-only backups (there is no database). Backups are stored under `backups/<domain>/<timestamp>/files.tar.gz` in the account's html-data volume, same layout every other CMS module uses for its own files backup.

### Remove

The **Remove** tab's **Delete Application** button fully uninstalls the site: deletes `index.php` and the `photos/` folder, and removes it from Site Manager. This cannot be undone.

---

## Not included

To keep this feature simple, the following are **not** part of TinyPhotoGallery support:

- A dedicated "TinyPhotoGallery Manager" sidebar page — manage installed sites from Site Manager instead
- Cloning a site
- Scanning the filesystem for untracked installations
- Security hardening rules
- Maintenance mode, admin auto-login, and cache-clearing — no such concepts exist in TinyPhotoGallery at all
- Version tracking — no tagged releases exist upstream, so the manager page always just shows "main"
- A database of any kind — TinyPhotoGallery is entirely flat-file
