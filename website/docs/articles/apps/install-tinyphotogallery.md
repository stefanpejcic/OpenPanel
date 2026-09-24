---
sidebar_label: "TinyPhotoGallery"
description: "How to create a simple photo gallery website with TinyPhotoGallery and OpenPanel: one-click install, uploading photos by File Manager or FTP, and backups."
---

# How to Create a Simple Photo Gallery Website

[TinyPhotoGallery](https://github.com/stefanpejcic/tinyphotogallery) turns a folder of images into a clean photo gallery website. There's no database, no admin panel and nothing to configure - upload photos and they appear.

---

## Install TinyPhotoGallery

1. Add a domain or subdomain in **OpenPanel → Domains**, e.g. `photos.example.com`.
2. Go to **OpenPanel → Websites → Install App** and click **Install TinyPhotoGallery**.
3. Select the **domain** (and optional subfolder) and start the installation.

![Install TinyPhotoGallery form with the site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/tinyphotogallery-form.png#gh-light-mode-only)
![Install TinyPhotoGallery form with the site details, domain and location, and admin credentials](/img/openpanel-screenshots/applications/tinyphotogallery-form_dark.png#gh-dark-mode-only)

OpenPanel downloads the gallery's `index.php` and creates an empty `photos/` folder.

---

## Add Photos

Upload images into the `photos/` folder:

- with the [File Manager](/docs/panel/files/files/) (**Upload** button), or
- with [FTP](/docs/panel/files/FTP/) for many photos at once.

Open `https://photos.example.com` - the photos are shown immediately. The manage page shows a **Setup** hint until the first photo is uploaded.

:::tip
Resize large camera photos (e.g. to 2000 px wide) before uploading - the gallery loads much faster and uses less disk space.
:::

To make the gallery private, [password-protect the directory](/docs/articles/websites/password-protect-directory/).

---

## Managing the Gallery

The manage page in **OpenPanel → Websites → Sites** has **Backups** (files-only, including your photos) and **Remove**, which deletes the gallery and the `photos/` folder.

More: [TinyPhotoGallery in OpenPanel](/docs/panel/applications/tinyphotogallery/).

---

## Related

- [Install TinyFileManager](/docs/articles/apps/install-tinyfilemanager/)
- [Build a website with the Website Builder](/docs/articles/apps/website-builder/)
