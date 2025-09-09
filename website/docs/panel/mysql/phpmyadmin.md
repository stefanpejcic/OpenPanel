---
sidebar_position: 4
---

# phpMyAdmin

phpMyAdmin is an advanced MySQL database management tool and is recommended only for experienced users.

![databases_phpmyadmin.png](/img/panel/v1/databases/databases_phpmyadmin.png)

To access phpMyAdmin, go to **Databases > phpMyAdmin** in the sidebar. You will be automatically logged in and can view all existing databases and their tables.

:::danger
Editing or manipulating tables directly in phpMyAdmin can break your site if done incorrectly. If you are unsure, consult a developer before making changes.
:::

Once logged in, you can:
- View database tables  
- Run SQL queries  
- Drop tables  
- Import data  
- Export databases (e.g., your WordPress database)  
- And more

For more information about using phpMyAdmin, refer to [the official phpMyAdmin documentation](https://www.phpmyadmin.net/docs/).

---

## Import SQL Files

phpMyAdmin allows you to import `.sql` files from different sources.

### 1. Import from Your Device
1. Open phpMyAdmin and (optionally) select the database you want to import into.  
2. Click **Import** in the top menu.  
3. Under **File to Import**, click **Choose File** and select the `.sql` file from your computer.  
4. Click **Import** at the bottom and wait for the process to complete.

### 2. Import from the Server Upload Directory
You can also import files already uploaded to your server in `/var/www/html/`.

1. Upload the `.sql` file to `/var/www/html/` using File Manager.  
2. Open phpMyAdmin and (optionally) select the target database.  
3. Click **Import** in the top menu.  
4. Under *Select from the web server upload directory /html/:*, choose the file name.  
5. Click **Import** at the bottom and wait for the process to complete.


### 3. Import via Terminal (Docker)
If the **docker** feature is enabled, you can import directly from the terminal inside your `mysql` or `mariadb` container.

1. Open **Containers > Terminal**.  
2. Select the `mysql` or `mariadb` container.  
3. Download or place your `.sql` file in the container.  
4. Run the appropriate command:  
   ```bash
   mysql -u username -p database_name < file.sql
   ```
   Replace `username`, `database_name`, and `file.sql` with your actual details.

