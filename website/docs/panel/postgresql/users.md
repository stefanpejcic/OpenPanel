---
sidebar_position: 4
---

# PostgreSQL Users

![PostgreSQL Users page listing database users with Change Password and Delete buttons](/img/openpanel-screenshots/postgresql/users-list.png#gh-light-mode-only)
![PostgreSQL Users page listing database users with Change Password and Delete buttons](/img/openpanel-screenshots/postgresql/users-list_dark.png#gh-dark-mode-only)

This section lists all your PostgreSQL users and offers options to reset a user's password or delete a user.

Available options on the Users page are:

- **Create a new user**
- **Assign a user to a database**
- **Remove a user from a database**
- **Reset a user's password**
- **Delete a user**

## Create a New Database User

PostgreSQL users (roles) are essential for controlling who can access and interact with your databases, ensuring data security and controlled access to your application's information.

To create a new database user, click on the "Create User" button and fill in the username and password for the new user.

![Create PostgreSQL User form with username and password fields and a password strength bar](/img/openpanel-screenshots/postgresql/new_user-form.png#gh-light-mode-only)
![Create PostgreSQL User form with username and password fields and a password strength bar](/img/openpanel-screenshots/postgresql/new_user-form_dark.png#gh-dark-mode-only)

## Assign a User to a Database

For a PostgreSQL user to be allowed to connect to a database, they need to be added (assigned) to that database. Assigning a user will grant them all privileges over the database. To assign a user to a specific database, click on the "Assign User" button and select a username and database.

![Assign User to Database form with user and database dropdowns](/img/openpanel-screenshots/postgresql/assign-form.png#gh-light-mode-only)
![Assign User to Database form with user and database dropdowns](/img/openpanel-screenshots/postgresql/assign-form_dark.png#gh-dark-mode-only)

## Remove a User from a Database

To remove a user from a database, click on the "Remove User" button, and in the new page, select a username to be removed from a database.

![Remove User from Database form with user and database dropdowns](/img/openpanel-screenshots/postgresql/remove-form.png#gh-light-mode-only)
![Remove User from Database form with user and database dropdowns](/img/openpanel-screenshots/postgresql/remove-form_dark.png#gh-dark-mode-only)

Removing a user will immediately remove all permissions for that user to the database and is useful when you want to temporarily disable a user's access to a database without actually deleting the user.

## Change User Password

If you need to change a user's password, click on the "Change Password" button next to that user. A page will open where you can insert the new password, then click on the "Change Password" button to save it.

![Change User Password form with the username prefilled and a new password field](/img/openpanel-screenshots/postgresql/password-form.png#gh-light-mode-only)
![Change User Password form with the username prefilled and a new password field](/img/openpanel-screenshots/postgresql/password-form_dark.png#gh-dark-mode-only)

## Delete User

To delete a PostgreSQL user, click the **Delete** button next to the user in the Users table. The button turns into **Confirm** with a 5-second countdown; click it again before the countdown ends to delete the user.

![Delete button turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/postgresql/users-delete.png#gh-light-mode-only)
![Delete button turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/postgresql/users-delete_dark.png#gh-dark-mode-only)

:::danger
⚠️ Deleting a PostgreSQL user will immediately remove that user and revoke all privileges to databases.
:::

:::info
Built-in PostgreSQL roles (such as `postgres` and the predefined `pg_*` roles) are marked as **System User** and cannot be edited or deleted.
:::

## Bulk Actions

Tick the checkbox of one or more users, or the checkbox in the table header to select every user shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![Two PostgreSQL users selected with the bulk actions bar offering Change password, Add to database, Remove from database and Delete](/img/openpanel-screenshots/postgresql/users-bulk.png#gh-light-mode-only)
![Two PostgreSQL users selected with the bulk actions bar offering Change password, Add to database, Remove from database and Delete](/img/openpanel-screenshots/postgresql/users-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Change password** | Sets the same new password on the selected users. |
| **Add to database** | Gives the selected users all privileges on the database you pick. |
| **Remove from database** | Removes the selected users from the database you pick. |
| **Delete** | Permanently deletes the selected users. |

System users can't be selected.

Click an action, confirm it, and it runs on the selected users one after another. When it's done the page reloads with a notice listing the users it worked for, or which ones failed and why.
