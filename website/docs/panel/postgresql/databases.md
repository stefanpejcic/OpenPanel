---
sidebar_position: 1
---

# Databases

![PostgreSQL Databases page listing databases with their size, assigned users and actions](/img/openpanel-screenshots/postgresql/databases-list.png#gh-light-mode-only)
![PostgreSQL Databases page listing databases with their size, assigned users and actions](/img/openpanel-screenshots/postgresql/databases-list_dark.png#gh-dark-mode-only)

PostgreSQL databases are used to store and manage your application's data, making it accessible and organized for your applications and services.

On the Databases page, you can view and manage your PostgreSQL databases.

The page provides a table containing the users databases, along with options related to them:

- **Databases:** Here, you can view database sizes and all users assigned to each database.

Available options on the Databases page are:

- **View Databases and assigned Users**
- **Create a new database**
- **Import or export a database**
- **Delete a PostgreSQL Database**

System databases (`postgres`, `template0`, `template1`) are always listed separately and cannot be deleted or imported into. Use the **Show system databases** switch to display or hide them in the table.

## Create a PostgreSQL Database

To create a new PostgreSQL database, click on the "New Database" button and fill in the name of the new database.

![Create PostgreSQL Database form with the database name field](/img/openpanel-screenshots/postgresql/new_db-form.png#gh-light-mode-only)
![Create PostgreSQL Database form with the database name field](/img/openpanel-screenshots/postgresql/new_db-form_dark.png#gh-dark-mode-only)

## Database Actions

Each database row has these actions:

- **Import**: only shown if the Import feature is enabled for your account. Opens the [Import](/docs/panel/postgresql/import/) page with this database pre-selected.
- **Export**: opens a small panel where you choose the export **Format**, `SQL` (`.sql`) or `GZIP` (`.sql.gz`), and **Destination** (download to your **Browser**, or save to a **Files** path under `/var/www/html/`), then click **Export**. The export is made with `pg_dump` and can be imported back on the Import page.
- **Delete**: permanently deletes the database.

## Delete a PostgreSQL Database

To delete an existing PostgreSQL database, click on the "Delete" button next to the database name in the table.

The button turns into **Confirm** with a 5-second countdown. Click it again before the countdown ends to delete the database; otherwise it reverts to **Delete**.

![Delete button turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/postgresql/databases-delete.png#gh-light-mode-only)
![Delete button turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/postgresql/databases-delete_dark.png#gh-dark-mode-only)

:::danger
⚠️ Deleting a PostgreSQL database will permanently delete all tables and data for that database.
:::

:::info
The **New Database** and **Database Wizard** actions are only available while the PostgreSQL service is running.
:::
