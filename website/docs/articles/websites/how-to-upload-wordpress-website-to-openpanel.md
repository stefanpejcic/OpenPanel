# How to Migrate a WordPress® Installation to OpenPanel

This guide will walk you through the process of uploading your WordPress website to **OpenPanel**, including domain setup, database configuration, file uploads, and final testing.

If you’re starting fresh with a new site, follow: [How to Install WordPress® With OpenPanel](/docs/articles/websites/how-to-install-wordpress-with-openpanel/)

---

## 1. Domain

### 1.1 Add a Domain Name

1. Log in to your **OpenPanel** dashboard.
2. Navigate to **Domains** → **Add Domain**.
3. Enter your domain name (e.g., `example.com`).
4. Save.

![add domain](https://i.postimg.cc/mkDT1Mhh/add-domain.png)

Make sure to point your domain’s DNS records (A record) to your OpenPanel server IP.

---

## 2. Database

### 2.1 Create Database and User

1. Go to **MySQL** → **Database Wizard**.
2. Set **Database Name**, **Database User** and set a strong password.
3. Click on the **Create DB, User and Grant Privileges** button.

![mysql db wizard](https://i.postimg.cc/rm4Pnwbg/db-wizard.png)

### 2.2 Import SQL File to Database

There are two ways to upload a database:

- Using 'Import Database' option if available.
- Using phpMyAdmin interface.

**Import SQL file using Import Database**:

1. Go to **MySQL** → **Import Database**.
2. Select **Database** and choose file from your device.
3. Click on **Upload & Import**.

![import sql file](https://i.postimg.cc/Vfwnd6sR/db-import.png)


**2.3 Import SQL file using phpMyAdmin**:

For uploading SQL files using phpMyAdmin interface, please follow this guide: [How to Import SQL Files](/docs/panel/mysql/phpmyadmin/#import-sql-files)

---

## 3. Files

### 3.1 Upload Files via File Manager

Before uploading, ensure your backup is archived (`.zip`, `.tar` or `.tar.gz`). If it isn’t, compress the files into one archive first.

1. Open **File Manager** in OpenPanel.
2. Navigate to your domain’s root directory (e.g., `/var/www/html/pejcic.rs/`).
3. Click the **Upload** button.
4. Select and upload your backup archive.
5. Once uploaded, go back to **File Manager**, select the archive, and click **Extract** → confirm extraction.
6. After extraction is complete, delete the uploaded archive by selecting it, clicking **Delete**, and confirming.

---

### 3.2 Change wp-config.php File

1. In File Manager, locate the **`wp-config.php`** file.
   ![edit wp config file](https://i.postimg.cc/mkLgNk21/edit-wp-config.png)
2. Edit the database settings:

   ```php
   define('DB_NAME', 'your_database_name');
   define('DB_USER', 'your_database_user');
   define('DB_PASSWORD', 'your_database_password');
   define('DB_HOST', 'mysql'); // mariadb or mysql - depending on your current setting
   ```
   ![example wp config ph](https://i.postimg.cc/3NtJLhdS/edit-wp-config-file.png)
3. Save changes.

---

## 4. Test

### 4.1 Test via Domain Name

1. Visit your domain (e.g., `https://example.com`).
2. If the DNS has propagated, your WordPress website should load.

### 4.2 Test via Live Preview

1. If DNS is not yet pointed, use OpenPanel’s **Live Preview** feature.
2. Go to WordPress Manager in OpenPanel → click **Scan for Existing Installation** and wait for process to complete.
3. Refresh the page and click on your website.
4. Click on the **Live Preview** button to check your website.

![wp scan](https://i.postimg.cc/npmRW3S9/scan-wp.png)
