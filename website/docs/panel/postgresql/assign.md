---
sidebar_position: 6
---

# Assign User to DB

Assign a PostgreSQL user to an existing database to grant permissions for executing queries, modifying tables, and managing data.

This is a critical step after creating a database and user - without it, the user won't be able to interact with the database.

To assign an existing user to a database, navigate to **PostgreSQL > Users > Assign user**:

1. **Select a user**
   Choose the PostgreSQL user you want to grant access to.

2. **Select a database**
   Choose the existing database the user should be assigned to.

3. **Click 'Assign User to Database'**
   After selecting both user and database, click the **Assign user to database** button.
   The user will be granted **all privileges** on the selected database.

![Assign User to Database form with user and database dropdowns](/img/openpanel-screenshots/postgresql/assign-form.png#gh-light-mode-only)
![Assign User to Database form with user and database dropdowns](/img/openpanel-screenshots/postgresql/assign-form_dark.png#gh-dark-mode-only)
