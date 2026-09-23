---
sidebar_position: 6
---

# Assign User to DB

Grant a MongoDB user a role scoped to an existing database, so they can execute queries, modify collections, and manage data.

This is a critical step after creating a database and user - without it, the user won't be able to interact with the database.

To assign an existing user to a database, navigate to **MongoDB > Users > Assign User**:

![Assign User to Database form with the user, database and role dropdowns](/img/openpanel-screenshots/mongodb/assign-form.png#gh-light-mode-only)
![Assign User to Database form with the user, database and role dropdowns](/img/openpanel-screenshots/mongodb/assign-form_dark.png#gh-dark-mode-only)

1. **Select a User**
   Choose the MongoDB user you want to grant access to.

2. **Select a Database**
   Choose the existing database the user should be assigned to.

3. **Select a Role**
   Choose the role to grant on the database:
   - `read` - read-only access
   - `readWrite` - read and write access (default)
   - `dbAdmin` - administrative tasks on the database (indexes, stats, schema)
   - `dbOwner` - full control over the database, combining `readWrite` and `dbAdmin`

4. **Click 'Assign User to Database'**
   After selecting the user, database, and role, click the **Assign User to Database** button.
