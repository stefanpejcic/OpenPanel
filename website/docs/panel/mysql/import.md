---
sidebar_position: 8
---

# Import

Import tables into a MySQL database using a `.sql` export file, for example one made with **Export** on the [Databases](/docs/panel/mysql/databases/#database-actions) page. This feature is useful for restoring backups, migrating data, or setting up initial database structures.


To import tables into a database, navigate to **OpenPanel > MySQL > Import**:

1. **Select a Database**  
   Choose the target database into which you want to import the tables.

2. **Select a File**  
   Click the **Select** button to choose a `.sql` (or `.sql.gz`) export file from your device.

3. **Upload**  
   Click the **Upload** button and wait for the import process to complete.

![Import into Database form with a database dropdown and a file picker for the .sql file](/img/openpanel-screenshots/mysql/import-form.png#gh-light-mode-only)
![Import into Database form with a database dropdown and a file picker for the .sql file](/img/openpanel-screenshots/mysql/import-form_dark.png#gh-dark-mode-only)

If your SQL file is larger than **1 GB**, we recommend using the **Containers > Terminal** interface instead for more reliable import handling.
