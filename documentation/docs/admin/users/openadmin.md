---
sidebar_position: 1
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


# OpenAdmin Users

The admin panel has two user roles:

- **Super Admin** that gets created when OpenPanel is installed and can add other users, with full control over the server.
- **Admin** - sub-admin users that are created by the SuperAdmin and have access to everything that the SuperAdmin has but can not edit the ‘Super Admin’ user.

## Manage Admin users


<Tabs>
  <TabItem value="openadmin-admin-users" label="With OpenAdmin" default>

To manage admin users that can access OpenAdmin interface use Settings > OpenAdmin page

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

To reset admin password click on the user in Settings > OpenAdmin page, then click on Edit button and set the password.

  </TabItem>
  <TabItem value="cli-reset" label="With OpenCLI">

To reset the password for an admin user:

```bash
opencli admin password <new_password> [username | admin]
```

Example, reset password for Super Admin user:
```bash
opencli admin password Pyl7_L2M1 admin
```

Example, reset password for regular Admin user:
```bash
opencli admin password Pyl7_L2M1 filip
```

  </TabItem>
</Tabs>


## Create new Admin

<Tabs>
  <TabItem value="openadmin-admin-new" label="With OpenAdmin" default>

To create new admin user click on the 'New' button in Settings > OpenAdmin page, set the username and password and click on save.

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
  <TabItem value="openadmin-admin-delete" label="With OpenAdmin" default>

To rename an Amdin user, select the user on **Settings > OpenAdmin** page and click on the Edit button and set new username.

  </TabItem>
  <TabItem value="cli-delete" label="With OpenCLI">

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
  <TabItem value="openadmin-admin-delete" label="With OpenAdmin" default>

To suspend an Admin user, select the user on **Settings > OpenAdmin** page and click on the Edit button, then **Suspend**.

To unsuspend an Admin user, select the user on **Settings > OpenAdmin** page and click on the Edit button, then **Unsuspend**.
  </TabItem>
  <TabItem value="cli-delete" label="With OpenCLI">

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

Select the user on Settings > OpenAdmin page and click on the delete button then confirm.

  </TabItem>
  <TabItem value="cli-delete" label="With OpenCLI">

From the terminal: [opencli admin delete](/docs/admin/scripts/admin#delete-admin-user)

  </TabItem>
</Tabs>


:::info
The Super Admin user can not be deleted.
:::


