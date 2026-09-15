---
sidebar_position: 1
---

# Databases

MongoDB databases are used to store and manage your application's document-based data, making it accessible and organized for your applications and services.

On the Databases page, you can view and manage your MongoDB databases.

The page provides a table containing the user's databases, along with options related to them:

- **Databases:** Here, you can view database sizes.

Available options on the Databases page are:

- **View Databases**
- **Create a new database**
- **Delete a MongoDB Database**

System databases (`admin`, `local`, `config`) are never shown in the table and cannot be created, imported into, or deleted through OpenPanel.

## Create a MongoDB Database

To create a new MongoDB database, click on the "New Database" button and fill in the name of the new database.

MongoDB has no explicit "create database" operation - the database is created as soon as it holds data, so a placeholder collection is created for you automatically.

## Delete a MongoDB Database

To delete an existing MongoDB database, click on the "Delete" button next to the database name in the table.

Then click on the same 'Confirm' button.

:::danger
⚠️ Deleting a MongoDB database will permanently delete all collections and data for that database.
:::

:::info
The **New Database** and **Database Wizard** actions are only available while the MongoDB service is running.
:::
