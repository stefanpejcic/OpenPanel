---
sidebar_position: 1
---

# Users

OpenPanel has a single user role named **User** that can only manage their docker container and inherits settings specified by the Admin user.


## List Users


<Tabs>
  <TabItem value="openadmin-users" label="OpenAdmin" default>

To access all OpenPanel users, navigate to OpenAdmin > Users.

The Users page displays a table showcasing each user's Gravatar linked to their email address, username, assigned IP Address, hosting plan name, creation date of the account, a login link enabling user impersonation, and *manage* button to get detailed user overview.

![openadmin users page](/img/admin/openadmin_users_list.gif)

Suspended users are highlighted in red, and no actions can be performed on a suspended user.

  </TabItem>
  <TabItem value="CLI-users" label="OpenCLI">

To list all users, use the following command:

```bash
opencli user-list
```

Example output:
```bash
opencli user-list
+----+----------------------------------+----------------------+----------------+------------------+-------+---------------------+
| id | username                         | email                | plan_name      | server           | owner | registered_date     |
+----+----------------------------------+----------------------+----------------+------------------+-------+---------------------+
|  3 | forums                           | stefan@openpanel.com | Standard plan  | forums           | NULL  | 2025-05-08 19:25:47 |
|  7 | pcx3                             | stefan@pejcic.rs     | Developer Plus | pcx3             | NULL  | 2025-05-09 12:26:20 |
|  9 | openpanelwebsite                 | info@openpanel.com   | Standard plan  | openpanelwebsite | NULL  | 2025-05-09 14:47:27 |
| 19 | SUSPENDED_20250529173435_radovan | radovan@jecmenica.rs | Standard plan  | radovan          | NULL  | 2025-05-29 07:47:15 |
+----+----------------------------------+----------------------+----------------+------------------+-------+---------------------+
```

You can also format the data as JSON:

```bash
opencli user-list --json
```
  </TabItem>
  <TabItem value="API-users" label="API">

To list all users, use the following api endpoint:

```bash
curl -X GET http://PANEL:2087/api/users -H "Authorization: Bearer JWT_TOKEN_HERE"
```

  </TabItem>
</Tabs>


## Create Users


<Tabs>
  <TabItem value="openadmin-users-new" label="OpenAdmin" default>

To create a new user, click on the 'New User' button on the Users page. A new section will be displayed with a form where you can set the email address, username, generate a strong password, and assign a hosting plan for the user.

![add new user openadmin](/img/admin/2025-06-09_08-20.png)

  </TabItem>
  <TabItem value="CLI-users-new" label="OpenCLI">

To create a new user run the following command:

```bash
opencli user-add <USERNAME> <PASSWORD> <EMAIL> <PLAN_NAME>
```
Example: 
```bash
opencli user-add filip masdhjkb213g filip@openadmin.co default_plan_apache
```

:::tip
Provide `generate` as password to generate a strong random password.
:::
  </TabItem>
  <TabItem value="API-users-new" label="API">

To create a new user use the following api call:

```bash
curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer JWT_TOKEN_HERE" -d '{"email": "EMAIL_HERE", "username": "USERNAME_HERE", "password": "PASSWORD_HERE", "plan_name": "PLAN_NAME_HERE"}' http://PANEL:2087/api/users
```

Example:
```bash
curl -X POST "http://PANEL:2087/api/users" -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGcBns" -H "Content-Type: application/json" -d '{"username":"stefan","password":"s32dsasaq","email":"stefan@pejcic.rs","plan_name":"default_plan_nginx"}'
```

Example response:

```json
{
  "response": {
    "message": "Successfully added user stefan password: s32dsasaq"
  },
  "success": true
}
```
  
  </TabItem>
</Tabs>

- The OpenPanel username must be 3 to 16 characters long and can only contain letters and numbers.
- The OpenPanel password must be 6 to 30 characters long and can include any characters except for single quotes (`'`) and double quotes (`"`).


## Reset User Password

<Tabs>
  <TabItem value="openadmin-users-reset" label="OpenAdmin" default>

To reset password for a user click on the user in *OpenAdmin > Users *or from the individual User page click on "Edit" tab and set the new password in the Password field then click Save.

![add new user openadmin](/img/admin/reset_password.png)


  </TabItem>
  <TabItem value="CLI-users-reset" label="OpenCLI">

To reset the password for a OpenPanel user, you can use the `user-password` command:

```bash
opencli user-password <USERNAME> <NEW_PASSWORD>
```

Use the `--ssh` flag to also change the password for the SSH user in the container.

Example:

```bash
opencli user-password filip Ty7_K8_M2 --ssh
```

  </TabItem>
  <TabItem value="API-users-reset" label="API">

To reset password for an OpenPanel user, use the following api call:

```bash
curl -X PATCH http://PANEL:2087/api/users/USERNAME_HERE -H "Content-Type: application/json" -H "Authorization: Bearer JWT_TOKEN_HERE" -d '{"password": "NEW_PASSWORD_HERE"}'
```
  </TabItem>
</Tabs>


## Single User

To view detailed information about a user, and edit their settings, click on their username in the users table.


### Statistics 

Statistics is the default tab, displays current usage statistics:

- Storage used
- Inodes used
- CPU usage
- Memoru usage
- Number of running containers
- Disk I/O
- Network I/O
- Number of PIDs
- Time statistics usage was last update
- Historical usage

![user statistics](/img/admin/user_usage.png)


### Services

Services page displays all user services (docker containers):

- Service name
- Docker Image name and tag
- Current CPU usage
- Allocated CPU for the service
- Current Memory usage
- Allocated Memory for the service
- Current status: Enabled or Disabled
- Terminal link to run docker exec commands in that service.

![docker services](/img/admin/docker_services.png)

### Overview

Overview page displays detailed user inforamtion and allows Administrator to set a custom message specifically for this user.

![user overview](/img/admin/2025-06-09_08-34.png)

Displayed information:

- User ID
- Email Address
- IP Address
- Geo Location for the IP
- Server Name
- Docker Context
- 2FA status
- Setup Time
- Custom Message for user


### Activity

Displays [users activity log](/docs/panel/account/account_activity/).

- Date
- Action performed
- IP Address

![user activity](/img/admin/login_log.png)

### Edit
From the Edit tab, Administrators can edit user information:

- Username
- Email address
- Password
- IP address
- Hosting Package

![user edit](/img/admin/edit_user.png)

## Suspend User

<Tabs>
  <TabItem value="openadmin-user-suspend" label="With OpenAdmin" default>

Suspending an account will immediately disable the user's access to the OpenPanel. This action involves pausing the user's Docker container and revoking access to their email, website, and other associated services. Please be aware of the immediate impact before proceeding.

To suspend a user click on the Suspend link on that user page and type the username to confirm, then click on 'Suspend account' button.

![suspend user](/img/admin/openadmin_suspend_user.gif)

  </TabItem>
  <TabItem value="CLI-user-suspend" label="With OpenCLI">

To suspend (temporary disable access) to user, run the following command:

```bash
opencli user-suspend <USERNAME>
```
Example:

```bash
opencli user-suspend filip
```


  </TabItem>
</Tabs>

## Unsuspend User

<Tabs>
  <TabItem value="openadmin-user-unsuspend" label="With OpenAdmin" default>

To unsuspend a user click on the Unsuspend button for that user.

  </TabItem>
  <TabItem value="CLI-user-unsuspend" label="With OpenCLI">
    
To unsuspend (enable access) to user, run the following command:

```bash
opencli user-unsuspend <USERNAME>
```

Example:
```bash
opencli user-unsuspend filip
```

  </TabItem>
</Tabs>



## Rename Username

<Tabs>
  <TabItem value="openadmin-user-username" label="With OpenAdmin" default>

To Rename a user, click on the 'Edit Information' link for the user, then change the address in 'Username' field and click on 'Save changes'.


  </TabItem>
  <TabItem value="CLI-user-email" label="With OpenCLI">

To change username for a user run the following command:

```bash
opencli user-rename <old_username> <new_username>
```

Example:

```bash
#opencli user-rename stefan pejcic
User 'stefan' successfully renamed to 'pejcic'.
```
  </TabItem>
</Tabs>



## Change Email address

<Tabs>
  <TabItem value="openadmin-user-email" label="With OpenAdmin" default>

To change email address for a user, click on the 'Edit Information' link for the user, then change the address in 'Email address' field and click on 'Save changes'.

  </TabItem>
  <TabItem value="CLI-user-email" label="With OpenCLI">

To change email address for a user run the following command:

```bash
opencli user-email <USERNAME> <NEW_EMAIL>
```

Example:

```bash
#opencli user-email stefan stefan@pejcic.rs
Email for user stefan updated to stefan@pejcic.rs.
```
  </TabItem>
</Tabs>




## Delete User


<Tabs>
  <TabItem value="openadmin-user-delete" label="With OpenAdmin" default>

To delete a user click on the delete button for that user, then type 'delete' in the confirmation modal and finally click on the 'Terminate' button.


  </TabItem>
  <TabItem value="CLI-user-delete" label="With OpenCLI">
    
To delete a user and all his data run the following command:

```bash
opencli user-delete <USERNAME>
```

add `-y` flag to disable prompt.

Example:
```bash
opencli user-delete filip -y
```

  </TabItem>
</Tabs>


:::danger
This action is irreversible and will permanently delete all user data.
:::
