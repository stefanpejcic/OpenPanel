---
sidebar_position: 1
---

# Databases

![databases.png](/img/panel/v1/databases/databases.png)

MySQL databases are used to store and manage your website's data, such as content, user information, and product details, making it accessible and organized for your web applications.

On the Databases page, you can view and manage your MySQL databases and users.

The page provides two tables: Databases and Users, along with options related to them:

- **Databases:** Here, you can view database sizes and all users assigned to each database.
- **Users:** This section lists all your MySQL users and offers options to reset a user's password or delete a user.

Available options on the Databases page are:

- **View Databases and Users**
- **Create a new database**
- **Delete a MySQL Database**
- **Create a new database user**
- **Assign a user to a database**
- **Remove a user from a database**
- **Reset a user's password**
- **Delete a user**

## Create a MySQL Database

To create a new MySQL database, click on the "New Database" button and fill in the name of the new database:

![databases_new.png](/img/panel/v1/databases/databases_new.png)

## Delete a MySQL Database

To delete an existing MySQL database, click on the "Delete" button next to the database name in the table.

Then click on the same 'Confirm' button.

![databases_delete_db.png](/img/panel/v1/databases/databases_delete_db.png)

:::danger
⚠️ Deleting a MySQL database will permanently delete all tables and data for that database.
:::

## Create a New Database User

MySQL users are essential for controlling who can access and interact with your databases, ensuring data security and controlled access to your website's information. 

To create a new database user, click on the "New User" button and fill in the name and password for the new user.

![databases_new_user.png](/img/panel/v1/databases/databases_new_user.png)

## Assign a User to a Database

For a MySQL user to be allowed to connect to a database, they need to be added (assigned) to that database. Assigning a user will grant them all privileges over the database. To assign a user to a specific database, click on the "Assign to Database" button and select a username and database.

![databases_assign.png](/img/panel/v1/databases/databases_assign.png)

## Remove a User from a Database

To remove a user from a database, simply click on the "Remove from Database" button, and in the new modal, select a username to be removed from a database.

![databases_remove_user.png](/img/panel/v1/databases/databases_remove_user.png)

Removing a user will immediately remove all permissions for that user to the database and is useful when you want to temporarily disable a user's access to a database without actually deleting the user.

## Change User Password

If you need to change a user's password, simply click on the "Change Password" button next to that user. A modal will appear, and you can insert the new password in the field, then click on the "Change Password" button to save the new password.

![databases_reset_password.png](/img/panel/v1/databases/databases_reset_password.png)

## Delete User

To delete a MySQL user, click on the delete button next to the user in the Users table:

![databases_delete_user.png](/img/panel/v1/databases/databases_delete_user.png)

:::danger
⚠️ Deleting a MySQL user will immediately remove that user and revoke all privileges to databases.
:::
