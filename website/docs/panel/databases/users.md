

(/img/panel/v2/databases_usrpw.png
(
(
(/img/panel/v2/databases_new_user.png
---
sidebar_position: 2
---

# MySQL Users

This section lists all your MySQL users and offers options to reset a user's password or delete a user.

Available options on the Users page are:

- **Create a new user**
- **Assign a user to a database**
- **Remove a user from a database**
- **Reset a user's password**
- **Delete a user**

## Create a New Database User

MySQL users are essential for controlling who can access and interact with your databases, ensuring data security and controlled access to your website's information. 

To create a new database user, click on the "New User" button and fill in the name and password for the new user.

![databases_new_user.png](/img/panel/v2/databases_new_user.png)

## Assign a User to a Database

For a MySQL user to be allowed to connect to a database, they need to be added (assigned) to that database. Assigning a user will grant them all privileges over the database. To assign a user to a specific database, click on the "Assign to Database" button and select a username and database.

![databases_assign.png](/img/panel/v2/databases_assign.png)

## Remove a User from a Database

To remove a user from a database, simply click on the "Remove User from Database" button, and in the new modal, select a username to be removed from a database.

![databases_remove_user.png](/img/panel/v2/databases_remove.png)

Removing a user will immediately remove all permissions for that user to the database and is useful when you want to temporarily disable a user's access to a database without actually deleting the user.

## Change User Password

If you need to change a user's password, simply click on the "Change Password" button next to that user. A modal will appear, and you can insert the new password in the field, then click on the "Change Password" button to save the new password.

![databases_reset_password.png](/img/panel/v2/databases_usrpw.png)

## Delete User

To delete a MySQL user, click on the delete button next to the user in the Users table and then click confirm on the same button:

![databases_delete_user.png](/img/panel/v2/databases_delusr.png)

:::danger
⚠️ Deleting a MySQL user will immediately remove that user and revoke all privileges to databases.
:::
