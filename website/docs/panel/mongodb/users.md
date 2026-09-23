---
sidebar_position: 2
---

# MongoDB Users

![MongoDB Users page listing users with their roles and the Change Password and Delete buttons](/img/openpanel-screenshots/mongodb/users-list.png#gh-light-mode-only)
![MongoDB Users page listing users with their roles and the Change Password and Delete buttons](/img/openpanel-screenshots/mongodb/users-list_dark.png#gh-dark-mode-only)

This section lists all your MongoDB users and offers options to reset a user's password or delete a user.

Available options on the Users page are:

- **Create a new user**
- **Assign a user to a database**
- **Remove a user from a database**
- **Reset a user's password**
- **Delete a user**

## Create a New Database User

MongoDB users are essential for controlling who can access and interact with your databases, ensuring data security and controlled access to your application's information.

To create a new database user, click on the "Create User" button and fill in the username and password for the new user. New users are created with no roles - they can't access any database until they're assigned to one.

![Create MongoDB User form with the username and password fields](/img/openpanel-screenshots/mongodb/new_user-form.png#gh-light-mode-only)
![Create MongoDB User form with the username and password fields](/img/openpanel-screenshots/mongodb/new_user-form_dark.png#gh-dark-mode-only)

## Assign a User to a Database

For a MongoDB user to be allowed to connect to a database, they need to be granted a role on that database. To assign a user to a specific database, click on the "Assign User" button and select a username, database, and role.

![Assign User to Database form with the user, database and role dropdowns](/img/openpanel-screenshots/mongodb/assign-form.png#gh-light-mode-only)
![Assign User to Database form with the user, database and role dropdowns](/img/openpanel-screenshots/mongodb/assign-form_dark.png#gh-dark-mode-only)

## Remove a User from a Database

To remove a user's access from a database, click on the "Remove User" button, and in the new page, select a username, database, and role to revoke.

![Remove User access from Database form with the user, database and role dropdowns](/img/openpanel-screenshots/mongodb/remove-form.png#gh-light-mode-only)
![Remove User access from Database form with the user, database and role dropdowns](/img/openpanel-screenshots/mongodb/remove-form_dark.png#gh-dark-mode-only)

Removing a user's role will immediately remove that permission on the database and is useful when you want to disable a user's access without actually deleting the user.

## Change User Password

If you need to change a user's password, click on the "Change Password" button next to that user. A page will open where you can insert the new password, then click on the "Change Password" button to save it.

![Change MongoDB user password form with the new password field](/img/openpanel-screenshots/mongodb/password-form.png#gh-light-mode-only)
![Change MongoDB user password form with the new password field](/img/openpanel-screenshots/mongodb/password-form_dark.png#gh-dark-mode-only)

## Delete User

To delete a MongoDB user, click the **Delete** button next to the user in the Users table. The button turns into **Confirm** with a 5-second countdown; click it again before the countdown ends to delete the user.

![Delete button of a MongoDB user turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/mongodb/users-delete.png#gh-light-mode-only)
![Delete button of a MongoDB user turned into a Confirm button with a countdown after the first click](/img/openpanel-screenshots/mongodb/users-delete_dark.png#gh-dark-mode-only)

:::danger
⚠️ Deleting a MongoDB user will immediately remove that user and revoke all privileges to databases.
:::

:::info
Built-in MongoDB users (such as `admin` and `root`) are marked as **System User** and cannot be edited or deleted.
:::
