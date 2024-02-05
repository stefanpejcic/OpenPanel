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

From the terminal:

[opencli admin new](/docs/admin/scripts/admin#create-new-admin)
[opencli admin password](/docs/admin/scripts/admin#reset-admin-password)
[opencli admin delete](/docs/admin/scripts/admin#delete-admin-user)

  </TabItem>
</Tabs>

## Reset Admin Password




<Tabs>
  <TabItem value="openadmin-admin-reset" label="With OpenAdmin" default>

To reset admin password click on the user in Settings > OpenAdmin page, then click on Edit button and set the password.

  </TabItem>
  <TabItem value="cli-reset" label="With OpenCLI">

From the terminal: [opencli admin password](/docs/admin/scripts/admin#reset-admin-password)

  </TabItem>
</Tabs>


## Create new Admin

<Tabs>
  <TabItem value="openadmin-admin-new" label="With OpenAdmin" default>

To create new admin user click on the 'New' button in Settings > OpenAdmin page, set the username and password and click on save.

  </TabItem>
  <TabItem value="cli-new" label="With OpenCLI">

From the terminal: [opencli admin new](/docs/admin/scripts/admin#create-new-admin)

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
