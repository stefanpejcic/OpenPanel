---
sidebar_position: 3
---

import Tabs from '@theme/Tabs';
import TabItem from '@theme/TabItem';


# Administrators

The admin panel has three user roles:


| Role              | Description                                                               |
| ------------------ | ------------------------------------------------------------------------- |
| **Super Admin**    | Has unrestricted privileges, created on OpenPanel installation. |
| **Admin**          | Has restricted privileges, can not access all OpenAdmin UI pages and can not edit the SuperAdmin user. |
| **Reseller**       | Has restricted privileges, manages only the users assigned to them. Reseller users are managed from the separate [Resellers](/docs/admin/accounts/resellers) page and are not listed on the Administrators page. |


## Manage Admin users


<Tabs>
  <TabItem value="openadmin-admin-users" label="With OpenAdmin" default>
  
  Manage administrative users with access to the OpenAdmin interface via **Accounts > Administrators**.
  
  For each admin user, the table shows: Username, Status, Role, 2FA (whether two-factor authentication is enabled), Passkeys (whether a passkey is registered), Last Login IP, Last Login Time, and an Edit menu with the available actions for that user.

  Reseller users are not listed on this page - they are managed separately under **Accounts > Resellers**.

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

To reset an admin's password, open the Edit menu for that user on the **Accounts > Administrators** page, select **Change Password**, then set the new password and click **Change Password**.

  </TabItem>
  <TabItem value="cli-reset" label="With OpenCLI">

To reset the password for an admin user:

```bash
opencli admin password <username> <new_password>
```

Example, reset password for and Admin user:
```bash
opencli admin password admin Pyl7_L2M1
```

  </TabItem>
</Tabs>


## Create new Admin

<Tabs>
  <TabItem value="openadmin-admin-new" label="With OpenAdmin" default>

To create a new admin user, click on the **Create New** button on the **Accounts > Administrators** page, set the username and password and click on **Create**.

:::info
Creating additional Administrator accounts requires an Enterprise license - Community edition supports only a single Administrator (the Super Admin) and does not display the **Create New** button. New admin users created this way are always assigned the **Admin** role; the **Super Admin** role can only be assigned during OpenPanel installation.
:::

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

To rename an Admin user, open the Edit menu for that user on the **Accounts > Administrators** page, select **Rename**, set the new username and click **Change Username**.

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

To suspend an Admin user, open the Edit menu for that user on the **Accounts > Administrators** page and click **Suspend**. To unsuspend the user, open the Edit menu again and click **Unsuspend**.

:::info
Only users with the **Admin** role can be suspended from this page. The **Super Admin** user cannot be suspended, and an admin cannot suspend their own account.
:::

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

Open the Edit menu for the user on the **Accounts > Administrators** page and click **Delete**, then confirm.

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


## Disable Two-Factor Authentication or Passkeys

<Tabs>
  <TabItem value="openadmin-admin-2fa" label="With OpenAdmin" default>

If an admin user has Two-Factor Authentication or a Passkey enabled (shown in the **2FA** and **Passkeys** columns on the **Accounts > Administrators** page), the Super Admin can open that user's Edit menu and select **Disable 2FA** or **Disable Passkeys** to remove it, for example if the user is locked out of their authenticator or security key. Only the Super Admin can perform this action, and it cannot be used on the Super Admin's own account.

  </TabItem>
</Tabs>


