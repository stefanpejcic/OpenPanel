---
sidebar_position: 2
---

# PostgreSQL Databases

The PostgreSQL tab lists the PostgreSQL databases of every user on the server.

Use **OpenAdmin > Databases > PostgreSQL** to access it.

![Databases page on the PostgreSQL tab with the Refresh button and the table of PostgreSQL databases from running services](/img/openadmin-screenshots/databases/postgresql-list.png#gh-light-mode-only)
![Databases page on the PostgreSQL tab with the Refresh button and the table of PostgreSQL databases from running services](/img/openadmin-screenshots/databases/postgresql-list_dark.png#gh-dark-mode-only)

PostgreSQL doesn't keep database names in a form that can be read from its files, so OpenAdmin asks each user's PostgreSQL service for them. Only users with a running service are checked, a few at a time, and the result is kept for 2 minutes. Click **Refresh** to check again right away.

| Column | Description |
|---|---|
| **Database** | Name of the database. |
| **Owner** | The OpenPanel user the database belongs to. |
| **Role** | The PostgreSQL role that owns the database. |
| **Size** | Size of the database. |

Users who have PostgreSQL but whose service is stopped are listed above the table under **Not running**. Their service is not started to list the databases. Users whose service didn't answer are listed with the error.

Resellers can open this tab too, and only see the databases of their own users.
