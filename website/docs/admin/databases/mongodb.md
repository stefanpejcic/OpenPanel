---
sidebar_position: 3
---

# MongoDB Databases

The MongoDB tab lists the MongoDB databases of every user on the server.

Use **OpenAdmin > Databases > MongoDB** to access it.

![Databases page on the MongoDB tab with the Refresh button and the table of MongoDB databases from running services](/img/openadmin-screenshots/databases/mongodb-list.png#gh-light-mode-only)
![Databases page on the MongoDB tab with the Refresh button and the table of MongoDB databases from running services](/img/openadmin-screenshots/databases/mongodb-list_dark.png#gh-dark-mode-only)

MongoDB doesn't keep database names in a form that can be read from its files, so OpenAdmin asks each user's MongoDB service for them. Only users with a running service are checked, a few at a time, and the result is kept for 2 minutes. Click **Refresh** to check again right away.

| Column | Description |
|---|---|
| **Database** | Name of the database. |
| **Owner** | The OpenPanel user the database belongs to. |
| **Collections** | Number of collections in the database. |
| **Size** | Size of the database on disk. |

The `admin`, `local` and `config` databases are not listed. Users who have MongoDB but whose service is stopped are listed above the table under **Not running**. Their service is not started to list the databases. Users whose service didn't answer are listed with the error.

Resellers can open this tab too, and only see the databases of their own users.
