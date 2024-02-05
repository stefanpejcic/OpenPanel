---
sidebar_position: 1
---

# OpenAdmin Users

The admin panel has two user roles:

- **Super Admin** that gets created when OpenPanel is installed and can add other users, with full control over the server.
- **Admin** - sub-admin users that are created by the SuperAdmin and have access to everything that the SuperAdmin has but can not edit the ‘Super Admin’ user.

## Manage Admin users

To manage admin users that can access OpenAdmin interface use Settings > OpenAdmin page

![openadmin admin users](/img/admin/openadmin_admin_page.png)

Or from the terminal:

- [opencli admin new](/docs/admin/scripts/admin#create-new-admin)
- [opencli admin password](/docs/admin/scripts/admin#reset-admin-password)
- [opencli admin delete](/docs/admin/scripts/admin#delete-admin-user)

## Reset Admin Password

To reset admin password click on the user in Settings > OpenAdmin page, then click on Edit button and set the password.

Or from the terminal: [opencli admin password](/docs/admin/scripts/admin#reset-admin-password)

## Create new Admin

To create new admin user click on the 'New' button in Settings > OpenAdmin page, set the username and password and click on save.

Or from the terminal: [opencli admin new](/docs/admin/scripts/admin#create-new-admin)

## Delete Admin user

Select the user on Settings > OpenAdmin page and click on the delete button then confirm.

Or from the terminal: [opencli admin delete](/docs/admin/scripts/admin#delete-admin-user)

:::info
The Super Admin user can not be deleted.
:::
