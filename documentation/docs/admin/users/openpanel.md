---
sidebar_position: 1
---

# OpenPanel Users

OpenPanel currently has only a single user role named **User** that can only manage their docker container and inherits settings specified by the Admin user.


## List Users


<Tabs>
  <TabItem value="openadmin-users" label="With OpenAdmin" default>

To access all OpenPanel users, navigate to OpenAdmin > Users.

The Users page displays a table showcasing each user's Gravatar linked to their email address, username, assigned IP Address, hosting plan name, creation date of the account, a login link enabling user impersonation, and an actions dropdown. In the actions dropdown, you can perform actions such as editing, suspending, or deleting the user.

![openadmin users page](/img/admin/openadmin_users_page.png)


Suspended users are highlighted in red, and no actions can be performed on a suspended user.

  </TabItem>
  <TabItem value="CLI-users" label="With OpenCLI">

To list all users, use the following command:

```bash
opencli user-list
```

Example output:
```bash
opencli user-list
+----+----------+-----------------+-----------------+---------------------+
| id | username | email           | plan_name       | registered_date     |
+----+----------+-----------------+-----------------+---------------------+
| 52 | stefan   | stefan          | cloud_4_nginx_3 | 2023-11-16 19:11:20 |
| 53 | petar    | petarc@petar.rs | cloud_8_nginx   | 2023-11-17 12:25:44 |
| 54 | rasa     | rasa@rasa.rs    | cloud_12_nginx  | 2023-11-17 15:09:28 |
+----+----------+-----------------+-----------------+---------------------+
```

You can also format the data as JSON:

```bash
opencli user-list --json
```

  </TabItem>
</Tabs>


## Create Users


<Tabs>
  <TabItem value="openadmin-users-new" label="With OpenAdmin" default>

To create a new user, click on the 'New User' button on the Users page. A new section will be displayed with a form where you can set the email address, username, generate a strong password, and assign a hosting plan for the user.

![openadmin users add new](/img/admin/openadmin_add_new_user.png)

  </TabItem>
  <TabItem value="CLI-users-new" label="With OpenCLI">

From the terminal: [opencli user-add](/docs/admin/scripts/users#add-user)

  </TabItem>
</Tabs>



## Reset User Password

<Tabs>
  <TabItem value="openadmin-users-reset" label="With OpenAdmin" default>

To reset password for a user click on the Edit dropdown in table for that user in OpenAdmin > Users or from the individual User page click on "Edit information" and set the new password in the Password field then save.

Step 1.             |  Step 2.
:-------------------------:|:-------------------------:
![openadmin users reset password step 1](/img/admin/openadmin_users_edit_information.png)  |  ![openadmin users reset password step 2](/img/admin/openadmin_users_edit_information_change_password.png)


  </TabItem>
  <TabItem value="CLI-users-reset" label="With OpenCLI">

From the terminal: [opencli user-password](/docs/admin/scripts/users#change-password)

  </TabItem>
</Tabs>


## Detailed User Information

To view detailed information about the account click on the Gravatar image or the username in the users table.

This page shows detailed information about the account and provides tools to manage it.

![openadmin users single user view](/img/admin/openadmin_users_single_user_view.png)

The username is displayed at the top, along with the status of the Docker container for the user. Colors indicate whether the user is suspended or if the Docker container has encountered an error. Next to the username, there are buttons that allow you to suspend/unsuspend the user, delete the user, a configure button to edit user settings inside their Docker container, and a 'Login as user' button that automatically logs you into their OpenPanel interface.

![openadmin users single user view 1](/img/admin/user_1.png)

There are 4 widgets on top of the page:

![openadmin users single user view 2](/img/admin/user_2.png)

- **CPU Usage:** Shows the current CPU usage of all processes by the user, represented in percentage with color indicators.
- **Memory Usage:** Displays the current memory usage of the Docker container, represented in percentage with color indicators.
- **Disk Usage:** Shows the current disk usage of the user's Docker container, represented in percentage with color indicators.
- **IP Address:** Displays the public IPv4 address for the user, indicating whether the user has a Dedicated IP address assigned.

----

The next section is divided into two parts: tabs and widgets.

There are 6 tabs that allow you to view relevant information about the user's Docker container:

### Docker

![openadmin users single user view docker_tab](/img/admin/docker_tab.png)

The Docker tab displays information about the Docker container for the user, including:

- **Server:** Server on which the container is running.
- **Private IP:** Private IP address within the plan network that the container is using.
- **ID:** Unique container ID for that user container.
- **Memory Allocated:** RAM allocated to the user container (can be manually extended even beyond the plan limit).
- **CPU:** Number of CPU cores allocated to the container (can also be increased outside the plan limits).
- **Docker Image:** The name of the image (Nginx or Apache) that the container is using and the OS in the image (Debian or Ubuntu).
- **Hostname:** Hostname that the user sees via SSH in their Docker container (same as your server name).
- **Exposed Ports:** Ports inside the Docker container that accept incoming connections (e.g., ports for SSH, MySQL, REDIS, Apache).
- **Container Created:** Timestamp when the container was started (may be different from the account creation date).

----

### Websites

The Websites tab will display all domains and websites that the user has inside their Docker container.

- **Domains table** showcases information such as domain name, root directory, and links to view the access log for the domain, edit DNS records, and edit the VirtualHost file for Nginx associated with the domain.
- **Websites table** displays the website URL, type (WordPress, Node.js, or Python), CMS version, and the time when the user installed or added it to the WP ;anager/PM2 interfaces.

![openadmin users single user view websites tab](/img/admin/websites_tab.png)

----

### Services

The Services tab displays a list of all services installed inside the user's Docker container, along with their current status. You have options to start, stop, or restart a service.

![openadmin users single user view services tab](/img/admin/services_tab.png)

----

### Backups

The Backups tab displays a list of all available backups for the user account, showcasing backup content and sizes.

![openadmin users single user view backups tab](/img/admin/backups_tab.png)

----

### Usage

The Usage tab will display Docker container stats for the user, including CPU usage, memory percentage used at that moment, network I/O, and total block I/O. This information is the same to what users can view from OpenPanel > Resource Usage.

![openadmin users single user view usage tab](/img/admin/usage_tab.png)

----

### Activity

The Activity tab shows the user's account activity log, providing the same information users can view from OpenPanel > Account Activity page.

![openadmin users single user view activity tab](/img/admin/activity_tab.png)

----

### General information

General information widget displays the general information about the user and their container:

![openadmin users single user view general info](/img/admin/general_info_tab.png)

- **User ID:** Unique ID for the user account.
- **Email Address:** Current email address for the user.
- **Two Factor:** Indicates whether 2FA is enabled by the user.
- **Hosting Plan:** Name of the hosting plan assigned to the user.
- **IP Address:** Public IPv4 address for the user.
- **Server Location:** Flag and country name indicating the geolocation of the IP address.
- **Private IP:** Private IP address for the Docker container used in internal networking.
- **Setup Date:** Date when the user account was created.
- **Domains:** Number of domains that the user has added.
- **Websites:** Number of websites that the user has added.

To edit any information for the user, click on the 'Edit Information' link, and a new modal will be displayed where you can change the username, email, plan, IP, or password.

![openadmin users single user edit_info_tab](/img/admin/edit_info_tab.png)

----

### Ports

The Ports widget displays all ports published in the user's Docker container and corresponding randomly generated ports for the user on the host server machine.

![openadmin users single user ports_tab](/img/admin/ports_tab.png)


## Suspend User

<Tabs>
  <TabItem value="openadmin-user-suspend" label="With OpenAdmin" default>

Suspending an account will immediately disable the user's access to the OpenPanel. This action involves pausing the user's Docker container and revoking access to their email, website, and other associated services. Please be aware of the immediate impact before proceeding.

To suspend a user click on the Suspend button on that user page and click on 'Suspend' on the confirmation modal.

Step 1.             |  Step 2.
:-------------------------:|:-------------------------:
![openadmin users suspend step 1](/img/admin/openadmin_users_suspend_1.png)  |  ![openadmin users suspend step 2](/img/admin/openadmin_users_suspend_2.png)

  </TabItem>
  <TabItem value="CLI-user-suspend" label="With OpenCLI">

From the terminal: [opencli user-suspend](/docs/admin/scripts/users#suspend-user)

  </TabItem>
</Tabs>

## Unsuspend User

<Tabs>
  <TabItem value="openadmin-user-unsuspend" label="With OpenAdmin" default>

To unsuspend a user click on the Unsuspend button for that user.

![openadmin users add new](/img/admin/openadmin_users_unsuspend.png)

  </TabItem>
  <TabItem value="CLI-user-unsuspend" label="With OpenCLI">
    
From the terminal: [opencli user-unsuspend](/docs/admin/scripts/users#unsuspend-user)

  </TabItem>
</Tabs>

## Change IP address for User

<Tabs>
  <TabItem value="openadmin-user-ip" label="With OpenAdmin" default>

To change IP address for a user, click on the 'Edit Information' link for the user, then elect the new IP address and click on 'Save changes'.

Step 1.             |  Step 2.
:-------------------------:|:-------------------------:
![openadmin users change ip address step 1](/img/admin/openadmin_users_edit_information.png)  |  ![openadmin users change ip address step 2](/img/admin/openadmin_users_edit_information_change_ip_address.png)

  </TabItem>
  <TabItem value="CLI-user-ip" label="With OpenCLI">
    
From the terminal: [opencli user-ip](/docs/admin/scripts/users#assign--remove-ip-to-user)

  </TabItem>
</Tabs>





## Delete User


<Tabs>
  <TabItem value="openadmin-user-delete" label="With OpenAdmin" default>

To delete a user click on the delete button for that user, then type 'delete' in the confirmation modal and finally click on the 'Terminate' button.

Step 1.             |  Step 2.
:-------------------------:|:-------------------------:
![openadmin users delete step 1](/img/admin/openadmin_users_delete_1.png)  |  ![openadmin users deletes step 2](/img/admin/openadmin_users_delete_2.png)

  </TabItem>
  <TabItem value="CLI-user-delete" label="With OpenCLI">
    
From the terminal: [opencli user-delete](/docs/admin/scripts/users#delete-user)

  </TabItem>
</Tabs>


:::danger
This action is irreversible and will permanently delete all user data.
:::
