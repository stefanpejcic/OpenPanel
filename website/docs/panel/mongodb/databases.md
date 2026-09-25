---
sidebar_position: 1
---

# Databases

![MongoDB Databases page listing databases with their size and the Delete button](/img/openpanel-screenshots/mongodb/databases-list.png#gh-light-mode-only)
![MongoDB Databases page listing databases with their size and the Delete button](/img/openpanel-screenshots/mongodb/databases-list_dark.png#gh-dark-mode-only)

MongoDB databases are used to store and manage your application's document-based data, making it accessible and organized for your applications and services.

On the Databases page, you can view and manage your MongoDB databases.

The page provides a table containing the user's databases, along with options related to them:

- **Databases:** Here, you can view database sizes.

Available options on the Databases page are:

- **View Databases**
- **Create a new database**
- **Import or export a database**
- **Delete a MongoDB Database**

System databases (`admin`, `local`, `config`) are never shown in the table and cannot be created, imported into, or deleted through OpenPanel.

## Create a MongoDB Database

To create a new MongoDB database, click on the "New Database" button and fill in the name of the new database.

![Create a MongoDB Database form with the database name field](/img/openpanel-screenshots/mongodb/new_db-form.png#gh-light-mode-only)
![Create a MongoDB Database form with the database name field](/img/openpanel-screenshots/mongodb/new_db-form_dark.png#gh-dark-mode-only)

MongoDB has no explicit "create database" operation - the database is created as soon as it holds data, so a placeholder collection is created for you automatically.

## Database Actions

Each database row has these actions:

- **Import**: only shown if the Import feature is enabled for your account. Opens the [Import](/docs/panel/mongodb/import/) page with this database pre-selected.
- **Export**: opens a small panel where you choose the export **Format**, `Archive` (`.archive`) or `GZIP` (`.archive.gz`), and **Destination** (download to your **Browser**, or save to a **Files** path under `/var/www/html/`), then click **Export**. The export is made with `mongodump` and can be imported back on the Import page, also into a database with a different name.
- **Delete**: permanently deletes the database.

## Delete a MongoDB Database

To delete an existing MongoDB database, click on the "Delete" button next to the database name in the table.

The button turns into **Confirm** with a 5-second countdown. Click it again before the countdown ends to delete the database; otherwise it reverts to **Delete**.

![Delete button of a MongoDB database turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/mongodb/databases-delete.png#gh-light-mode-only)
![Delete button of a MongoDB database turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/mongodb/databases-delete_dark.png#gh-dark-mode-only)

:::danger
⚠️ Deleting a MongoDB database will permanently delete all collections and data for that database.
:::

:::info
The **New Database** and **Database Wizard** actions are only available while the MongoDB service is running.
:::
