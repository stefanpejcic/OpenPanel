---
sidebar_position: 1
---

# Hosting Plans

Hosting plans outline the available services and limitations for users. [Docker images](https://dev.openpanel.co/images/) specify the pre-installed services for a plan. [Enabled Features](/docs/admin/settings/openpanel/#enable-features) determine which pages users can access from the OpenPanel interface. 


## List hosting plans

<Tabs>
  <TabItem value="openadmin-plan-list" label="With OpenAdmin" default>


To list existing plans navigate to Plans page:

![openadmin list plans](/img/admin/adminpanel_plans.png)

| Field              | Description                                                               |
| ------------------ | ------------------------------------------------------------------------- |
| **ID**             | Unique ID for the plan.                                                    |
| **Plan Name**      | Display name that users will see in their OpenPanel dashboards.            |
| **Description**    | Visible only to administrators.                                           |
| **Image**          | Name of the Docker image used when creating new accounts on the plan.        |
| **DOcker Size**     | Disk space allocated for the user's container .          |
| **Storage**     | Disk space allocated for user's websites files *(/home/user/ directory)*.           |
| **Domains**  | Total number of domain names allowed per user on the plan.                  |
| **Websites** | Total number of websites (WordPress, NodeJS, Python) per user on the plan.   |
| **Databases** | Total number of MySQL databases allowed per user on the plan.              |
| **CPU**            | Number of CPU cores dedicated to the user on this hosting plan.             |
| **RAM**            | Physical Memory (RAM) in GB allocated to the user on this hosting plan.     |
| **Port Speed**            | Maximum post speed for users in mbit/s.     |
| **Inodes Limit**   | *(DEPRECATED)* Limits the total number of files allowed in the container.   |




  </TabItem>
  <TabItem value="CLI-plan-list" label="With OpenCLI">

To list all current hosting packages (plans) run:

```bash
opencli plan-list
```

Example output:
```bash
opencli plan-list
+----+-----------------+------------------------+---------------+----------------+------------+--------------+----------+------+------+-----------------+-----------+
| id | name            | description            | domains_limit | websites_limit | disk_limit | inodes_limit | db_limit | cpu  | ram  | docker_image    | bandwidth |
+----+-----------------+------------------------+---------------+----------------+------------+--------------+----------+------+------+-----------------+-----------+
|  1 | cloud_4_nginx   | 20gb space and Nginx   |             0 |             10 | 20 GB      |      1000000 |        0 | 4    | 4g   | dev_plan_nginx  |       100 |
|  2 | cloud_4_apache  | 20gb space and Apache  |             0 |             10 | 20 GB      |      1000000 |        0 | 4    | 4g   | dev_plan_apache |       100 |
|  3 | cloud_8_nginx   | 80gb space and Nginx   |             0 |             50 | 80 GB      |      2000000 |        0 | 8    | 8g   | dev_plan_nginx  |       200 |
|  4 | cloud_8_apache  | 80gb space and Apache  |             0 |             50 | 80 GB      |      2000000 |        0 | 8    | 8g   | dev_plan_apache |       200 |
+----+-----------------+------------------------+---------------+----------------+------------+--------------+----------+------+------+-----------------+-----------+
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

To create a new hosting plan click on the 'Create new plan' nutton and set the desired limits for the plan.

![openadmin create new plan](/img/admin/adminpanel_plans_create_new.gig)

  </TabItem>
  <TabItem value="CLI-plan-new" label="With OpenCLI">
    
To create a new plan run the following command:

```bash
opencli plan-create <NAME> <DESCRIPTION> <DOMAINS_LIMIT> <WEBSITES_LIMIT> <DISK_LIMIT> <INODES_LIMITS> <DATABASES_LIMIT> <CPU_LIMIT> <RAM_LIMIT> <DOCKER_IMAGE> <PORT_SPEED_LIMIT>
```

Example:
```bash
opencli plan-create cloud_8 "Custom plan with 8GB of RAM&CPU" 0 0 15 500000 0 8 8 nginx 200
```

  </TabItem>
</Tabs>


## Change user plan

Please visit [this page](/docs/admin/plans/change-plan-for-user)

## Modify plan

To change plan limits click on the edit button for the plan and set the new limits.


Step 1.             |  Step 2.
:-------------------------:|:-------------------------:
![openadmin edit plan step 1](/img/admin/openadmin_edit_plan_1.png)  |  ![openadmin edit plan step 2](/img/admin/openadmin_edit_plan_2.png)

The new limits will be applied immediately to all accounts.

## List Users on Plan

<Tabs>
  <TabItem value="openadmin-plan-usage" label="With OpenAdmin" default>

To view all users that are currently using a hosting plan, simply sort the users table by that plan name, or in the search fields type the plan name.

  </TabItem>
  <TabItem value="CLI-plan-usage" label="With OpenCLI">
    
List all users that are currently using a plan:

```bash
opencli plan-usage
```

Example:
```bash
opencli plan-usage 2
+----+----------+-------+----------------+---------------------+
| id | username | email | plan_name      | registered_date     |
+----+----------+-------+----------------+---------------------+
|  2 | rasa     | rasa  | cloud_4_apache | 2023-11-30 10:33:52 |
|  3 | aas      | sasa  | cloud_4_apache | 2023-11-30 12:01:49 |
+----+----------+-------+----------------+---------------------+
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
    
To delete a hosting plan click on the delete button next to the plan name.

Step 1.             |  Step 2.
:-------------------------:|:-------------------------:
![openadmin delete plan step 1](/img/admin/admin_delete_plan_1.png)  |  ![openadmin delete plan step 2](/img/admin/admin_delete_plan_2.png)

  </TabItem>
  <TabItem value="CLI-plan-delete" label="With OpenCLI">

To delete a hosting plan: 

```bash
opencli plan-delete <PLAN_NAME> 
```

Example:
```bash
opencli plan-delete 32
```
  </TabItem>
</Tabs>

NOTE: Plan can not be deleted if it has users assigned. 

