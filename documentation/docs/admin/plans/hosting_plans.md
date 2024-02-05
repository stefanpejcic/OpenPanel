---
sidebar_position: 1
---

# Hosting Plans

Hosting plans are used to set services and limits for users.

## List hosting plans

To list existing plans navigate to Plans page:

![openadmin list plans](/img/admin/adminpanel_plans.png)

| Field              | Description                                                               |
| ------------------ | ------------------------------------------------------------------------- |
| **ID**             | Unique ID for the plan.                                                    |
| **Name**           | Display name that users will see in their OpenPanel dashboards.            |
| **Description**    | Visible only to administrators.                                           |
| **Domains Limit**  | Total number of domain names allowed per user on the plan.                  |
| **Websites Limit** | Total number of websites (WordPress, NodeJS, Python) per user on the plan.   |
| **Disk Limit**     | Disk space allocated for the user's container           |
| **Inodes Limit**   | *(DEPRECATED)* Limits the total number of files allowed in the container.   |
| **Database Limit** | Total number of MySQL databases allowed per user on the plan.              |
| **CPU**            | Number of CPU cores dedicated to the user on this hosting plan.             |
| **RAM**            | Physical Memory (RAM) in GB allocated to the user on this hosting plan.     |
| **Docker Image**   | Name of the Docker image used when creating new accounts on the plan.        |


## Create a plan

To create a new hosting plan click on the 'Create new plan' nutton and set the desired limits for the plan.

![openadmin create new plan](/img/admin/adminpanel_plans_create_new.png)

## Change user plan

Please visit [this page](/docs/admin/plans/change-plan-for-user)

## Modify plan

To change plan limits click on the edit button for the plan and set the new limits.

The new limits will be applied to new accounts only. To apply the limits to all users on the same plan automatically you can use the `apply limits` script.

When changing the limits for a plan, to apply them to existing users:
- Domains, Websites, and Databases limits will automatically increase.
- CPU and RAM limits require the docker container to be rebuilt.
- Disk space **CAN NOT BE CHANGED**, inodes **CAN NOT LONGER BE SET**.
- Bandwidth (port speed) limits require a manual intervention.

