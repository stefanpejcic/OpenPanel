---
sidebar_position: 5
---

# Switch MySQL Type

The **Docker > Switch MySQL Type** page allows you to switch your current mysql service between available options: **MySQL** and **MariaDB**.

The currently active type is displayed in the top-right corner of the page.

## Requirements

To access this feature:

- The **Docker** module must be enabled **server-wide** by an Administrator.
- Your account must have the **Docker** feature enabled.

## Usage

Before switching the databse type, please ensure the following:

- **All existing databases and users must be removed.**
- The current mysql container must be **stopped** before the new one can be started.

> ⚠️ If you already have databases configured, **back up all data**, remove all databases and users, then proceed with switching the mysql server.  
> To avoid downtime, it's best to make this change **before adding any databases**.

### Steps to Switch

1. In the OpenPanel menu, navigate to **Docker > Switch MySQL Type**.
2. From the dropdown menu, select the new type you want to use.
3. Click the **Switch** button to initiate the process.

After confirmation:

- The existing database container will be stopped and its data will be removed.
- The new mysql type server will be started.
- You can then re-add your database and users under the new server.
