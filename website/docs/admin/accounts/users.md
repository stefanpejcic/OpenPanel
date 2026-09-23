---
sidebar_position: 1
---

# Users

OpenPanel has a single user role named **User** that can only manage their docker container and inherits settings specified by the Admin user.


## List Users


<Tabs>
  <TabItem value="openadmin-users" label="OpenAdmin" default>
  
  To access all OpenPanel users, navigate to **Accounts > Users**.
  
  The Users page displays a table with user information and buttons to manage it.
  
  ![Users page listing OpenPanel accounts with their status, plan, limits, usage and the Impersonate button](/img/openadmin-screenshots/accounts/users-list.png#gh-light-mode-only)
  ![Users page listing OpenPanel accounts with their status, plan, limits, usage and the Impersonate button](/img/openadmin-screenshots/accounts/users-list_dark.png#gh-dark-mode-only)
  
  Additional columns can be displayed using the 'Show Columns' button.

  ![Show Columns menu of the Users table with a toggle for each column](/img/openadmin-screenshots/accounts/users-columns.png#gh-light-mode-only)
  ![Show Columns menu of the Users table with a toggle for each column](/img/openadmin-screenshots/accounts/users-columns_dark.png#gh-dark-mode-only)

  Suspended users are highlighted in red.

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

To create a new user, click on the **Create New** button on the Users page. A form is displayed where you can set the username, email address (optionally sending the user a welcome email with their login credentials), and generate a strong password.

You can also choose the webserver for the account (and optionally enable Varnish Cache), the database type (MySQL or MariaDB), assign a reseller as the account's owner (Enterprise license only, and only when creating the user as a Super Admin/Admin), and select a hosting plan to assign to the user.

![Create New user form with username, email, password, webserver, database type and hosting plan](/img/openadmin-screenshots/accounts/users-new.png#gh-light-mode-only)
![Create New user form with username, email, password, webserver, database type and hosting plan](/img/openadmin-screenshots/accounts/users-new_dark.png#gh-dark-mode-only)

  </TabItem>
  <TabItem value="CLI-users-new" label="OpenCLI">

To create a new user run the following command:

```bash
opencli user-add <USERNAME> <PASSWORD> <EMAIL> <PLAN_NAME>
```
Example: 
```bash
opencli user-add filip password1234 filip@openadmin.com default_plan_apache
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
curl -X POST "http://PANEL:2087/api/users" -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGcBns" -H "Content-Type: application/json" -d '{"username":"stefan","password":"strongishpassword1234","email":"stefan@pejcic.rs","plan_name":"default_plan_nginx"}'
```

Example response:

```json
{
  "response": {
    "message": "Successfully added user stefan password: strongishpassword1234"
  },
  "success": true
}
```
  
  </TabItem>
</Tabs>

- The OpenPanel username must be 3 to 20 characters long and can only contain letters and numbers.
- The OpenPanel password must be 6 to 30 characters long and cannot contain an apostrophe (`'`).


## Single User

To view detailed information about a user, and edit their settings, click on their username in the users table. The user page has a menu on the left with these tabs: Overview, Services, Storage, Edit, Permissions, Export, Suspend (replaced by a single Unsuspend action if the account is already suspended), Delete, Activity Log and Login Log.


### Statistics 

Overview is the default tab. Its top part displays current usage statistics:

- Storage used
- Inodes used
- CPU usage
- Memory usage
- Number of running containers
- Disk I/O
- Network I/O
- Number of PIDs
- Time statistics usage was last update
- Historical usage

![Overview tab of a user with gauges for storage, inodes, CPU and memory usage and the View Past Usage button](/img/openadmin-screenshots/accounts/users-stats.png#gh-light-mode-only)
![Overview tab of a user with gauges for storage, inodes, CPU and memory usage and the View Past Usage button](/img/openadmin-screenshots/accounts/users-stats_dark.png#gh-dark-mode-only)

Clicking on 'View Past Usage' will display a table with past resource usage for the user: Date, CPU %, CPU usage, Memory %, Memory usage and Tasks.

![Past resource usage table of a user with the date, CPU and memory usage and the number of tasks](/img/openadmin-screenshots/accounts/users-history.png#gh-light-mode-only)
![Past resource usage table of a user with the date, CPU and memory usage and the number of tasks](/img/openadmin-screenshots/accounts/users-history_dark.png#gh-dark-mode-only)


### Services

Services tab displays all user services (docker containers). Columns can be toggled with the 'Show Columns' button and include:

- Service name
- Docker Image name and tag
- Published ports
- Environment variables (sensitive values such as passwords are masked and can be revealed on click)
- Current CPU usage
- Current Memory usage
- Actions, including a terminal link to run docker exec commands in that service.

An 'Edit Services' button also lets Administrators edit the raw service configuration.

![Services tab listing the user containers with their CPU and memory usage, PIDs and actions](/img/openadmin-screenshots/accounts/users-services.png#gh-light-mode-only)
![Services tab listing the user containers with their CPU and memory usage, PIDs and actions](/img/openadmin-screenshots/accounts/users-services_dark.png#gh-dark-mode-only)

### Storage

Storage tab displays data from the [docker system df](https://docs.docker.com/reference/cli/docker/system/df/) command.

- Volumes
- Containers
- Images

![Storage tab with the user volumes, containers and images from docker system df](/img/openadmin-screenshots/accounts/users-storage.png#gh-light-mode-only)
![Storage tab with the user volumes, containers and images from docker system df](/img/openadmin-screenshots/accounts/users-storage_dark.png#gh-dark-mode-only)

### Overview

Below the usage statistics, the Overview tab displays detailed user information.

![User details on the Overview tab: username, email, plan, locale, 2FA status, IP address, location, server, docker context and setup time](/img/openadmin-screenshots/accounts/users-info.png#gh-light-mode-only)
![User details on the Overview tab: username, email, plan, locale, 2FA status, IP address, location, server, docker context and setup time](/img/openadmin-screenshots/accounts/users-info_dark.png#gh-dark-mode-only)

Displayed information:

- Username
- Email Address
- User ID
- IP Address
- Geo Location for the IP
- Server Name
- Docker Context
- 2FA status
- Reseller (if the user is owned by a reseller)
- Setup Time


### Permissions

The Permissions tab lets Administrators view and, for individual users, override which OpenPanel features/pages are enabled. By default a user's permissions follow their hosting plan's defaults; switching to **Custom** mode allows enabling or disabling individual features for that user only, independent of the plan. Plan-wide feature defaults are managed separately in Feature Manager.

![Permissions tab where the enabled OpenPanel features follow the plan or are set per user](/img/openadmin-screenshots/accounts/users-permissions.png#gh-light-mode-only)
![Permissions tab where the enabled OpenPanel features follow the plan or are set per user](/img/openadmin-screenshots/accounts/users-permissions_dark.png#gh-dark-mode-only)

### Activity

Displays [users activity log](/docs/panel/security/account_activity/).

- Date
- Action performed
- IP Address

![Activity Log tab with the date and the action performed](/img/openadmin-screenshots/accounts/users-activity.png#gh-light-mode-only)
![Activity Log tab with the date and the action performed](/img/openadmin-screenshots/accounts/users-activity_dark.png#gh-dark-mode-only)

### Login Log

Displays a log of successful logins for the user, separate from the general Activity log:

- Date
- Country
- IP Address

![Login Log tab with the date, country and IP address of each login](/img/openadmin-screenshots/accounts/users-logins.png#gh-light-mode-only)
![Login Log tab with the date, country and IP address of each login](/img/openadmin-screenshots/accounts/users-logins_dark.png#gh-dark-mode-only)

### Edit
From the Edit tab, Administrators can edit user information:

- Username
- Email address
- Password (leave empty to keep the current password)
- IP address
- Reseller (change the account's owner/reseller)
- Hosting Package

Click **Save** to apply the changes.

![Edit tab with the username, email, password, IP address, reseller and hosting package fields](/img/openadmin-screenshots/accounts/users-edit.png#gh-light-mode-only)
![Edit tab with the username, email, password, IP address, reseller and hosting package fields](/img/openadmin-screenshots/accounts/users-edit_dark.png#gh-dark-mode-only)

Below the form, the Edit tab also lets Administrators set a custom message that is shown to this user in OpenPanel.

### Export

The Export tab has two options:

- **Generate full account backup** – Creates a compressed archive of the account's home directory, databases, domains, websites, email, FTP, DNS zones, SSL certificates, cronjobs and containers/images. Previously generated backups are listed below, where they can be downloaded or deleted.

![Export tab with the Generate full account backup option and the list of existing backups](/img/openadmin-screenshots/accounts/users-export.png#gh-light-mode-only)
![Export tab with the Generate full account backup option and the list of existing backups](/img/openadmin-screenshots/accounts/users-export_dark.png#gh-dark-mode-only)

- **Transfer to another server** (Enterprise license only) – Migrates the user account, along with all its containers and data, to another OpenPanel server over SSH. You provide the remote server's address/port and root SSH credentials, and can optionally enable "Live Transfer" so that once the migration completes, the account is automatically suspended on the current server and its domains' DNS is updated to point to the new server.

![Transfer to another server form on the Export tab with the remote server, SSH credentials and the Live Transfer option](/img/openadmin-screenshots/accounts/users-transfer.png#gh-light-mode-only)
![Transfer to another server form on the Export tab with the remote server, SSH credentials and the Live Transfer option](/img/openadmin-screenshots/accounts/users-transfer_dark.png#gh-dark-mode-only)

### Suspend

<Tabs>
  <TabItem value="openadmin-user-suspend" label="With OpenAdmin" default>

Suspending an account will immediately disable the user's access to the OpenPanel. This action involves pausing the user's Docker container and revoking access to their email, website, and other associated services. Please be aware of the immediate impact before proceeding.

To suspend a user, open the "Suspend" tab on that user's page and type the username to confirm, then click the **Suspend account** button.

![Suspend tab asking to type the username before suspending the account](/img/openadmin-screenshots/accounts/users-suspend.png#gh-light-mode-only)
![Suspend tab asking to type the username before suspending the account](/img/openadmin-screenshots/accounts/users-suspend_dark.png#gh-dark-mode-only)

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

### Unsuspend

<Tabs>
  <TabItem value="openadmin-user-unsuspend" label="With OpenAdmin" default>

For a suspended user, the row of tabs on the user's page is replaced with a single **Unsuspend** button. Click it to restore access for the user.

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


### Reset Password

<Tabs>
  <TabItem value="openadmin-users-reset" label="OpenAdmin" default>

To reset password for a user, click on the "Edit" tab and set the new password in the Password field (leave it empty to keep the current password) then click **Save**.

![Edit tab with the password field and the Generate button](/img/openadmin-screenshots/accounts/users-edit.png#gh-light-mode-only)
![Edit tab with the password field and the Generate button](/img/openadmin-screenshots/accounts/users-edit_dark.png#gh-dark-mode-only)


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



### Rename

<Tabs>
  <TabItem value="openadmin-user-username" label="With OpenAdmin" default>

To rename a user, click on the "Edit" tab for the user, then change the 'Username' field and click **Save**.


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


### Change Package

<Tabs>
  <TabItem value="openadmin-user-plan" label="With OpenAdmin" default>

To change a package for a user, click on the "Edit" tab for the user, then select the new hosting plan and click **Save**.

  </TabItem>
  <TabItem value="CLI-user-plan" label="With OpenCLI">

To change a package for a user run the following command:

```bash
opencli user-change_plan <USERNAME> '<NEW_PLAN_NAME>'
```
  </TabItem>
</Tabs>


### Change Email

<Tabs>
  <TabItem value="openadmin-user-email" label="With OpenAdmin" default>

To change email address for a user, click on the "Edit" tab for the user, then change the 'Email' field and click **Save**.

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



### Login to OpenPanel

To auto-login to a user's OpenPanel account, click on the **Impersonate** button in the top-right corner of the user's page.
 

### Delete User


<Tabs>
  <TabItem value="openadmin-user-delete" label="With OpenAdmin" default>

To delete a user, open the "Delete" tab for that user, type the username to confirm, then click **Delete account permanently**.

![Delete tab asking to type the username before deleting the account permanently](/img/openadmin-screenshots/accounts/users-delete.png#gh-light-mode-only)
![Delete tab asking to type the username before deleting the account permanently](/img/openadmin-screenshots/accounts/users-delete_dark.png#gh-dark-mode-only)


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

