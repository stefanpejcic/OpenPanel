---
sidebar_position: 4
---

# phpMyAdmin

phpMyAdmin is an advanced MySQL database management tool and is recommended only for experienced users.

## Manage

To access phpMyAdmin, go to **Databases > phpMyAdmin** in the sidebar, or from **Databases** page click on the 'Open phpMyAdmin' icon next to the database name.


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

phpMyAdmin allows you to import `.sql` files from your device:
1. Open phpMyAdmin and select the database you want to import into.  
2. Click **Import** in the top menu.  
3. Under **File to Import**, click **Choose File** and select the `.sql` file from your computer.  
4. Click **Import** at the bottom and wait for the process to complete.
