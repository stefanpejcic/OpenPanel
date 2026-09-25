---
sidebar_position: 8
---

# Import

Import data into a MongoDB database using a `mongodump` archive file (`.archive` or `.archive.gz`), for example one made with **Export** on the [Databases](/docs/panel/mongodb/databases/#database-actions) page. The archive can be imported into a database with a different name than the one it was exported from. This feature is useful for restoring backups, migrating data, or setting up initial database structures.

To import data into a database, navigate to **MongoDB > Import**:

1. **Select a Database**
   Choose the target database into which you want to import the data.

2. **Select a File**
   Click the **Select** button to choose a `.archive` or `.archive.gz` file from your device.

3. **Upload**
   Click the **Upload & Import** button and wait for the import process to complete.

:::info
Importing replaces any existing collections in the target database that are present in the archive (`mongorestore --drop`).
:::

If your archive file is very large, we recommend using the **Containers > Terminal** interface instead for more reliable import handling.
