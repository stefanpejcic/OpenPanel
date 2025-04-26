---
sidebar_position: 3
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


# OpenAdmin Users

The admin panel has three user roles:


| Role              | Description                                                               |
| ------------------ | ------------------------------------------------------------------------- |
| **Super Admin**    | Has unrestricted privileges, created on OpenPanel installation. |
| **Admin**          | Has unrestricted privileges, but can not edit the SuperAdmin user. |
| **Reseller**       | Has restrited privileges. |


## Manage Admin users


<Tabs>
  <TabItem value="openadmin-admin-users" label="With OpenAdmin" default>

Use *Settings > OpenAdmin* page to manage admin users that can access OpenAdmin interface:

![openadmin admin users](/img/admin/openadmin_admin_page.png)

  </TabItem>
  <TabItem value="CLI" label="With OpenCLI">

To list admin users use command:

```bash
opencli admin list
```

  </TabItem>
</Tabs>

## Reset Admin Password


<Tabs>
  <TabItem value="openadmin-admin-reset" label="With OpenAdmin" default>

To reset admin password click on the Edit button for that user from *Settings > OpenAdmin* page, then set the new password.

![openadmin admin password](/img/admin/openadmin_admin_password.png)

  </TabItem>
  <TabItem value="cli-reset" label="With OpenCLI">

To reset the password for an admin user:

```bash
opencli admin password <username> <new_password>
```

Example, reset password for Super Admin user:
```bash
opencli admin password admin Pyl7_L2M1
```

Example, reset password for regular Admin user:
```bash
opencli admin password filip Pyl7_L2M1
```

  </TabItem>
</Tabs>


## Create new Admin

<Tabs>
  <TabItem value="openadmin-admin-new" label="With OpenAdmin" default>

To create new admin user click on the 'New' button in *Settings > OpenAdmin* page, set the username and password and click on *Save*.

![openadmin admin new](/img/admin/openadmin_admin_new.png)


  </TabItem>
  <TabItem value="cli-new" label="With OpenCLI">

To create new admin accounts:

```bash
opencli admin new <username> <password>
```

Example:
```bash
opencli admin new filip Pyl7_L2M1
```

  </TabItem>
</Tabs>





## Rename Admin user

<Tabs>
  <TabItem value="openadmin-admin-rename" label="With OpenAdmin" default>

To rename an Admin user, select the user on **Settings > OpenAdmin** page and click on the Edit button and set new username.

![openadmin admin rename](/img/admin/openadmin_admin_rename.png)


  </TabItem>
  <TabItem value="cli-rename" label="With OpenCLI">

To rename admin user:

```bash
opencli admin rename <username> <new_username>
```

Example:
```bash
opencli admin rename filip filip2
```
  </TabItem>
</Tabs>


## Suspend Admin user

<Tabs>
  <TabItem value="openadmin-admin-suspend" label="With OpenAdmin" default>

To suspend an Admin user, select the user on **Settings > OpenAdmin** page and click on the Edit button, then uncheck the **Active** status.

![openadmin admin suspend](/img/admin/openadmin_admin_suspend.png)


To unsuspend an Admin user, select the user on **Settings > OpenAdmin** page and click on the Edit button, then **Unsuspend**.
  </TabItem>
  <TabItem value="cli-suspend" label="With OpenCLI">

```bash
opencli admin suspend <username>
```

Example:
```bash
opencli admin suspend filip
```
---

To unsuspend admin user:
```bash
opencli admin unsuspend <username>
```

Example:
```bash
opencli admin unsuspend filip
```

  </TabItem>
</Tabs>


## Delete Admin user

<Tabs>
  <TabItem value="openadmin-admin-delete" label="With OpenAdmin" default>

Select the user on *Settings > OpenAdmin* page and click on the delete button then confirm.

![openadmin admin delete](/img/admin/openadmin_admin_delete.png)


  </TabItem>
  <TabItem value="cli-delete" label="With OpenCLI">

From the terminal:

To delete admin user:
```bash
opencli admin delete <username>
```

Example:
```bash
opencli admin delete filip
```

  </TabItem>
</Tabs>


:::info
The Super Admin user can not be deleted.
:::


