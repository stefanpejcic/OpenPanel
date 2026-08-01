---
sidebar_position: 3
---

# pgAdmin

pgAdmin is an advanced PostgreSQL database management tool and is recommended only for experienced users.

## Manage

To access pgAdmin, go to **PostgreSQL > pgAdmin** in the sidebar, or from the **Databases** page click on the 'pgAdmin' icon next to the database name.

:::danger
Editing or manipulating tables directly in pgAdmin can break your application if done incorrectly. If you are unsure, consult a developer before making changes.
:::

Once logged in, you can:
- View database tables
- Run SQL queries
- Drop tables
- Import data
- Export databases
- And more

If pgAdmin has not been enabled server-wide by your hosting provider, the page will show that it is **Not currently enabled by Administrator** and you should contact your hosting provider.

## Settings

On the pgAdmin page you can view and edit:

- **Email** - the login email used to access the pgAdmin interface
- **Password** - the login password used to access the pgAdmin interface
- **Version** - the pgAdmin container image tag in use

## Port

The **Server** and **Port** cards show the IP and port used to reach the pgAdmin service directly.
