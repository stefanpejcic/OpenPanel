---
sidebar_position: 1
---

# User Packages

Hosting plans set limits for users. 

## List hosting plans

<Tabs>
  <TabItem value="openadmin-plan-list" label="With OpenAdmin" default>


To list existing plans navigate to Plans page:

| Field              | Description                                                               |
| ------------------ | ------------------------------------------------------------------------- |
| **Plan Name**      | Display name that users will see in their OpenPanel dashboards.            |
| **Memory**            | Physical Memory (RAM) in GB allocated to the user on this hosting plan.     |
| **CPU**            | Number of CPU cores dedicated to the user on this hosting plan.             |
| **Disk**     | Disk space in GB allocated for all user files.           |
| **Inodes**   | Limits the total number of files allowed for the user.   |
| **Port Speed**            | Maximum post speed for users in mbit/s.     |
| **Domains**  | Total number of domain names allowed per user on the plan.                  |
| **Websites** | Total number of websites (WordPress, NodeJS, Python) per user on the plan.   |
| **Databases** | Total number of MySQL/MariaDB databases allowed per user on the plan.              |
| **Email accounts** | Total number of email accounts that user can create on the plan.              |
| **FTP accounts** | Total number of ftp accounts that user can create on the plan.             |
| **Feature Set** | [Feature Sets](/docs/admin/settings/openpanel/#enable-features) determine which pages users can access from the OpenPanel interface.               |



  </TabItem>
  <TabItem value="CLI-plan-list" label="With OpenCLI">

To list all current hosting packages (plans) run:

```bash
opencli plan-list
```

Example output:
```bash
[root@fajlovi ~]# opencli plan-list
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+
| id | name           | description            | domains_limit | websites_limit | email_limit | ftp_limit | disk_limit | inodes_limit | db_limit | cpu  | ram  | bandwidth | feature_set |
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+
|  1 | Standard plan  | Small plan for testing |             0 |             10 |           0 |         0 | 5 GB       |      1000000 |        0 | 2    | 2g   |        10 | default     |
|  2 | Developer Plus | 4 cores, 6G ram        |             0 |             10 |           0 |         0 | 10 GB      |      1000000 |        0 | 2    | 3g   |       100 | proba       |
|  3 | unlimited      | unlimited              |             0 |              0 |           0 |         0 | 100 GB     |       250000 |        0 | 2    | 3g   |         0 | default     |
|  4 | unlimited2     | default                |             0 |              0 |           0 |         0 | 0 GB       |            0 |        0 | 0    | 0g   |         0 | proba       |
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+

```

You can also format the data as JSON:

```bash
opencli plan-list --json
```

  </TabItem>
</Tabs>

## Create a plan

<Tabs>
  <TabItem value="openadmin-plan-new" label="With OpenAdmin" default>

To create a new hosting package, click the **'Create New'** button and configure the desired limits:

* **Name** – Can include any characters.
* **Description** – Internal note for admins, visible only in OpenAdmin.
* **Disk** – Storage in GB. Use `0` for unlimited.
* **Inodes** – Number of inodes. Use `0` for unlimited.
* **CPU** – Number of CPU cores allocated across all user services. Set to `0` for unlimited. Cannot exceed total server cores.
* **Memory** – Amount of physical memory in GB allocated across all user services. Set to `0` for unlimited. Cannot exceed total server memory.
* **Port Speed** – Maximum speed in Mbit/s for user services *(Deprecated and not enforced)*.
* **Databases** – Max number of databases (MySQL, MariaDB, PostgreSQL). Use `0` for unlimited.
* **Websites** – Max number of websites in Site Manager (WordPress, WebsiteBuilder, NodeJS/Python). Use `0` for unlimited.
* **FTP accounts** – Max number of FTP sub-accounts. Use `0` for unlimited.
* **Email accounts** – Max number of email sub-accounts. Use `0` for unlimited.
* **Feature Set** – Name of the feature set that defines available services in the OpenPanel UI.


  </TabItem>
  <TabItem value="CLI-plan-new" label="With OpenCLI">
    
To create a new plan run the following command:

```bash
opencli plan-create 'name' 'description' email_limit ftp_limit domains_limit websites_limit disk_limit inodes_limit db_limit cpu ram bandwidth feature_set
```

Example:
```bash
opencli plan-create Another Plan' 'this plan is used for X' 0 0 10 10 50 1000000 25 2 4 100 'default'
```

  </TabItem>
</Tabs>


## Modify plan

To change plan limits click on the **Edit** button for the plan in **OpenAdmin > User Packages** and set the new limits.

The new limits will be applied immediately to all accounts using the package.

## List Users on Plan

<Tabs>
  <TabItem value="openadmin-plan-usage" label="With OpenAdmin" default>

To view all users that are currently using a hosting package, simply sort the users table by that package name, or in the search field type the package name.

  </TabItem>
  <TabItem value="CLI-plan-usage" label="With OpenCLI">
    
List all users that are currently using a plan:

```bash
opencli plan-usage
```

Example:
```bash
[root@fajlovi ~]# opencli plan-usage 'Standard plan'
+----+----------------------------------+----------------------+---------------+---------------------+
| id | username                         | email                | plan_name     | registered_date     |
+----+----------------------------------+----------------------+---------------+---------------------+
|  3 | forums                           | stefan@openpanel.com | Standard plan | 2025-05-08 19:25:47 |
| 19 | radovan                          | radovan@jecmenica.rs | Standard plan | 2025-05-29 07:47:15 |
+----+----------------------------------+----------------------+---------------+---------------------+
```

You can also format the data as JSON:

```bash
opencli plan-usage --json
```
  </TabItem>
</Tabs>

## Delete Plan

<Tabs>
  <TabItem value="openadmin-plan-delete" label="With OpenAdmin" default>
    
To delete a hosting package click on the **Delete** button next to the package name.

  </TabItem>
  <TabItem value="CLI-plan-delete" label="With OpenCLI">

To delete a hosting plan: 

```bash
opencli plan-delete <PLAN_NAME> 
```

Example:
```bash
opencli plan-delete 'Standard plan'
```
  </TabItem>
</Tabs>

Note: A package cannot be deleted if it has users assigned to it.
