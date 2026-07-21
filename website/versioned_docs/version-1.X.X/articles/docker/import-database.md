# Importing a Database

You can import tables into a MySQL/MariaDB database using one of the following methods:

* **Import Database** feature (via OpenPanel)
* **phpMyAdmin** interface
* **Terminal** feature (via OpenPanel)

---

## 1. Import Database (via OpenPanel)

The **Import Database** feature must be enabled on your plan for the option to appear in your OpenPanel dashboard.
Once available, follow the guide here: [Database Import Documentation](/docs/panel/mysql/import/)

---

## 2. Using phpMyAdmin

If the **phpMyAdmin** feature is enabled:

* Open phpMyAdmin from your dashboard.
* Select your target database.
* Use the **Import** tab to upload and execute your `.sql` file.

Alternatively, you can upload the SQL file using the **File Manager** to the `/var/www/html` directory. The uploaded files will then be visible on the phpMyAdmin import page.

Documentation: [phpMyAdmin Import Guide](https://openpanel.com/docs/panel/mysql/phpmyadmin/)

---

## 3. Using Terminal (Docker)

To use terminal access:

* The **Docker** feature must be enabled on your plan for container access to be visible in OpenPanel.
* Once enabled, select your MySQL/MariaDB container from the dashboard.
* Open the terminal and run the following commands to import your SQL file:

```bash
mysql

USE your_database_name;

wget https://example.com/some_file.sql

SOURCE ./your_file.sql;
```
