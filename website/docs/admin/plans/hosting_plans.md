---
sidebar_position: 1
---

# User Packages

Hosting plans set limits for users. 

## List hosting plans

<Tabs>
  <TabItem value="openadmin-plan-list" label="With OpenAdmin" default>


To list existing plans navigate to **OpenAdmin > Hosting Plans > User Packages**:

![openadmin plans](/img/admin/tremor/plans_list.png)


| Field              | Description                                                               |
| ------------------ | ------------------------------------------------------------------------- |
| **Name**      | Display name of the plan.            |
| **Description**   | Optional plan description, shown to users on their Server Information page. |
| **Memory**            | Physical Memory (RAM) in GB allocated to the user on this hosting plan.     |
| **CPU**            | Number of CPU cores dedicated to the user on this hosting plan.             |
| **Disk**     | Disk space in GB allocated for all user files.           |
| **Inodes**   | Limits the total number of files allowed for the user.   |
| **Port Speed**            | Maximum bandwidth (port speed) for the user, in mbit/s.     |
| **Domains**  | Total number of domain names allowed per user on the plan.                  |
| **Websites** | Total number of websites (WordPress, Website Builder, NodeJS, Python) per user on the plan.   |
| **Databases** | Total number of MySQL and PostgreSQL databases allowed per user on the plan. The limit applies separately to each database engine.              |
| **Email accounts** | Total number of email accounts that user can create on the plan.              |
| **Mailbox quota** | Max mailbox size for email accounts that user can set on this plan.              |
| **Max hourly emails** | Max number of emails that all addresses under this account can send within one hour.              |
| **FTP accounts** | Total number of ftp accounts that user can create on the plan.             |
| **Feature Set** | [Feature Sets](/docs/admin/plans/feature-manager) determine which pages users can access from the OpenPanel interface.               |
| **Used by** | Number of users currently on this plan. Click the number to view those users. |



  </TabItem>
  <TabItem value="CLI-plan-list" label="With OpenCLI">

To list all current hosting packages (plans) run:

```bash
opencli plan-list
```

Example output:
```bash
[root@fajlovi ~]# opencli plan-list
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+-----------------+------------------+
| id | name           | description            | domains_limit | websites_limit | email_limit | ftp_limit | disk_limit | inodes_limit | db_limit | cpu  | ram  | bandwidth | feature_set | max_email_quota | max_hourly_email |
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+-----------------+------------------+
|  1 | Standard plan  | Small plan for testing |             0 |             10 |           0 |         0 | 5 GB       |      1000000 |        0 | 2    | 2g   |        10 | basic       | 0               | 0                |
|  2 | Developer Plus | 4 cores, 6G ram        |             0 |             10 |           0 |         0 | 20 GB      |      2500000 |        0 | 4    | 6g   |       100 | default     | 0               | 1000             |
+----+----------------+------------------------+---------------+----------------+-------------+-----------+------------+--------------+----------+------+------+-----------+-------------+-----------------+------------------+

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

![openadmin plans create](/img/admin/tremor/plans_create.png)


* **Name** – Plan name.
* **Description** – Optional plan description. It is visible to users on their Server Information page, not just to admins.
* **Disk** – Disk space in GB to allocate for user services, files and logs. Use `0` for unlimited.
* **Inodes** – Total number of files to allow the user on this plan. Use `0` for unlimited.
* **CPU cores** – Limits the total CPU (cores) used by all of the user's running services combined. Set to `0` for unlimited.
* **Memory** – Limits the total memory (in GB) used by all of the user's running services combined. Set to `0` for unlimited.
* **Port Speed** – Bandwidth (port speed) limit, in Mbit/s, for the user's services.
* **Domains** – Total number of domains: primary, addons, aliases. Subdomains are excluded.
* **Websites** – Max number of websites the user can install/manage on this plan. Includes WordPress, NodeJS, Python and Website Builder. Use `0` for unlimited.
* **Databases** – Max number of MySQL and PostgreSQL databases the user can have on this plan. The limit applies separately to each engine (e.g. a value of `10` allows up to 10 MySQL databases *and* up to 10 PostgreSQL databases). Use `0` for unlimited.
* **Email Accounts** – Max number of email accounts the user can create and manage on this plan. Use `0` for unlimited.
* **Mailbox quota** – Max mailbox size for email accounts that user can set on this plan. Examples: `100k`, `250M`, `3G`, `1T`, or `0` for unlimited.
* **Max hourly emails** – Max number of emails that all addresses under this account can send within one hour. Use `0` for unlimited.
* **FTP Accounts** – Max number of FTP accounts the user can create and manage on this plan. Use `0` for unlimited.
* **Feature set** – Feature set that determines which features users on this plan have access to.

</TabItem>
<TabItem value="CLI-plan-new" label="With OpenCLI">
    
To create a new plan run the following command:

```bash
pencli plan-create name"<TEXT>" description="<TEXT>" emails=<COUNT> ftp=<COUNT> domains=<COUNT> websites=<COUNT> disk=<COUNT> inodes=<COUNT> databases=<COUNT> cpu=<COUNT> ram=<COUNT> bandwidth=<COUNT> feature_set=<NAME> max_email_quota=<COUNT> max_hourly_email=<COUNT>
```

Example:
```bash
opencli plan-create name="New Plan" description="This is a new plan" emails=100 ftp=50 domains=20 websites=30 disk=100 inodes=100000 databases=10 cpu=4 ram=8 bandwidth=100 feature_set=default max_email_quota=2G max_hourly_email=1000
```

  </TabItem>
</Tabs>


## Modify plan

To change plan limits, go to **OpenAdmin > Hosting Plans > User Packages**, click the **...** (kebab) menu at the end of the plan's row, choose **Edit**, and set the new limits.

![openadmin plans edit](/img/admin/tremor/plans_edit_1.png)

![openadmin plans edit limits](/img/admin/tremor/plans_edit_2.png)


The new limits will be applied immediately to all accounts using the package.

## List Users on Plan

<Tabs>
  <TabItem value="openadmin-plan-usage" label="With OpenAdmin" default>

To view all users that are currently using a hosting package, click the number shown in the **Used by** column for that plan on the User Packages page - this opens the Users list already filtered to that package. Alternatively, on the Users page, sort the table by the **Package** column, or type the package name in the search field.

![openadmin plans usage](/img/admin/tremor/plans_usage_1.png)

![openadmin plans usage](/img/admin/tremor/plans_usage_2.png)

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
    
To delete a hosting package, click the **...** (kebab) menu at the end of the package's row and choose **Delete**. The **Delete** option only appears in the menu when the package has no users assigned to it.

![openadmin plans delete](/img/admin/tremor/plans_delete.png)


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
