---
sidebar_position: 2
---

# Reseller Packages

Reseller packages are limit templates for [reseller accounts](/docs/admin/accounts/resellers), the same way [User Packages](/docs/admin/plans/hosting_plans) are for users. Pick a package when creating a reseller, or let your billing system create resellers on a package through the API, and its limits are applied automatically.

:::info
Resellers is an Enterprise-only feature.
:::

Each package sets:

| Field | Description |
|---|---|
| **Name** | Used to pick the package in OpenAdmin, OpenCLI, the API and billing systems. Letters, numbers, `-` and `_` only, up to 64 characters. |
| **Description** | Optional note, for example `10 accounts, 50 GB`. |
| **Max Accounts** | How many user accounts resellers on this package can have. `0` means unlimited. |
| **Max Disk Usage** | Total disk space for all accounts of the reseller, in GB. `0` means unlimited. |
| **User Packages** | Which [User Packages](/docs/admin/plans/hosting_plans) resellers on this package can assign to their users. |

Resellers created without a package get the default limits from `/etc/openpanel/openadmin/config/reseller_template.json`.

## List reseller packages

<Tabs>
  <TabItem value="openadmin-reseller-package-list" label="With OpenAdmin" default>

Navigate to **OpenAdmin > Hosting Plans > Reseller Packages**. The table shows each package's name and description, Max Accounts, Max Disk, the number of User Packages it allows and how many resellers are on it (hover the number to see who). Click the arrows next to a column name to sort the table by that column, unlimited (`0`) sorts as the largest value.

![Reseller Packages page listing packages with their max accounts, max disk, number of user packages and resellers on each](/img/openadmin-screenshots/plans/reseller_packages-list.png#gh-light-mode-only)
![Reseller Packages page listing packages with their max accounts, max disk, number of user packages and resellers on each](/img/openadmin-screenshots/plans/reseller_packages-list_dark.png#gh-dark-mode-only)

  </TabItem>
  <TabItem value="API-reseller-package-list" label="With API">

```bash
curl -X GET https://SERVER:2087/api/plans/resellers -H "Authorization: Bearer TOKEN_HERE"
```

Example output:

```json
[
  {
    "name": "reseller_basic",
    "description": "10 accounts, 50 GB",
    "max_accounts": 10,
    "max_disk_blocks": 51200000,
    "allowed_plans": [1, 2],
    "resellers": ["reseller1"]
  }
]
```

  </TabItem>
</Tabs>

## Create a reseller package

<Tabs>
  <TabItem value="openadmin-reseller-package-new" label="With OpenAdmin" default>

On **OpenAdmin > Hosting Plans > Reseller Packages** click **Create New**, fill in the name, description, Max Accounts and Max Disk Usage, tick the User Packages resellers on it can assign, and click **Create**.

![Create New reseller package form with name, description, max accounts, max disk usage and the user packages resellers can assign](/img/openadmin-screenshots/plans/reseller_packages-new.png#gh-light-mode-only)
![Create New reseller package form with name, description, max accounts, max disk usage and the user packages resellers can assign](/img/openadmin-screenshots/plans/reseller_packages-new_dark.png#gh-dark-mode-only)

  </TabItem>
  <TabItem value="API-reseller-package-new" label="With API">

```bash
curl -X POST https://SERVER:2087/api/plans/resellers \
  -H "Authorization: Bearer TOKEN_HERE" -H "Content-Type: application/json" \
  -d '{"action":"create","name":"reseller_basic","description":"10 accounts, 50 GB","max_accounts":10,"max_disk_blocks":51200000,"allowed_plans":[1,2]}'
```

Over the API the disk limit is set in blocks, 1 GB is `1024000` blocks. `allowed_plans` takes plan IDs, as a list (`[1,2]`) or a string (`"1,2"`).

  </TabItem>
</Tabs>

## Edit a reseller package

<Tabs>
  <TabItem value="openadmin-reseller-package-edit" label="With OpenAdmin" default>

Click the package name, or open the package's **Edit** menu and click **Edit**, change the limits and click **Save**. The page header lists the resellers on the package.

![Edit page of a reseller package listing the resellers on it, with its limits and allowed user packages](/img/openadmin-screenshots/plans/reseller_packages-edit.png#gh-light-mode-only)
![Edit page of a reseller package listing the resellers on it, with its limits and allowed user packages](/img/openadmin-screenshots/plans/reseller_packages-edit_dark.png#gh-dark-mode-only)

  </TabItem>
  <TabItem value="API-reseller-package-edit" label="With API">

```bash
curl -X POST https://SERVER:2087/api/plans/resellers \
  -H "Authorization: Bearer TOKEN_HERE" -H "Content-Type: application/json" \
  -d '{"action":"update","name":"reseller_basic","max_accounts":20,"max_disk_blocks":102400000,"allowed_plans":[1,2,3]}'
```

`update` replaces all limits of the package, send every field.

  </TabItem>
</Tabs>

Saving a package applies the new limits to every reseller on it. If a reseller already has more accounts or uses more disk than the new limits allow, they keep their old limits and the notice lists them.

## Delete a reseller package

<Tabs>
  <TabItem value="openadmin-reseller-package-delete" label="With OpenAdmin" default>

Open the package's **Edit** menu, click **Delete** and confirm. **Delete** isn't shown while resellers are on the package, move them to another package first.

![Edit menu of a reseller package with the Edit and Delete options](/img/openadmin-screenshots/plans/reseller_packages-menu.png#gh-light-mode-only)
![Edit menu of a reseller package with the Edit and Delete options](/img/openadmin-screenshots/plans/reseller_packages-menu_dark.png#gh-dark-mode-only)

  </TabItem>
  <TabItem value="API-reseller-package-delete" label="With API">

```bash
curl -X POST https://SERVER:2087/api/plans/resellers \
  -H "Authorization: Bearer TOKEN_HERE" -H "Content-Type: application/json" \
  -d '{"action":"delete","name":"reseller_basic"}'
```

  </TabItem>
</Tabs>

## Create a reseller on a package

<Tabs>
  <TabItem value="openadmin-reseller-on-package" label="With OpenAdmin" default>

On **Accounts > Resellers** click **Create New**, set the username and password, pick the package from **Reseller Package** and click **Create**.

  </TabItem>
  <TabItem value="CLI-reseller-on-package" label="With OpenCLI">

```bash
opencli admin new <username> <password> --reseller --plan=<package>
```

  </TabItem>
  <TabItem value="API-reseller-on-package" label="With API">

```bash
curl -X POST https://SERVER:2087/api/settings/resellers \
  -H "Authorization: Bearer TOKEN_HERE" -H "Content-Type: application/json" \
  -d '{"action":"create","username":"reseller1","password":"secret123","plan":"reseller_basic"}'
```

  </TabItem>
</Tabs>

## Move a reseller to another package

Use this to upgrade or downgrade a reseller. Moving to a package with fewer accounts or less disk than the reseller already uses is refused with a message saying what to free up first.

<Tabs>
  <TabItem value="openadmin-reseller-change-package" label="With OpenAdmin" default>

On **Accounts > Resellers** open **Edit Plans & Limits** for the reseller, pick the package under **Reseller Package** and click **Apply Package**.

![Edit Plans & Limits page of a reseller with the Reseller Package picker and the Apply Package button above the custom limits](/img/openadmin-screenshots/accounts/resellers-package.png#gh-light-mode-only)
![Edit Plans & Limits page of a reseller with the Reseller Package picker and the Apply Package button above the custom limits](/img/openadmin-screenshots/accounts/resellers-package_dark.png#gh-dark-mode-only)

  </TabItem>
  <TabItem value="CLI-reseller-change-package" label="With OpenCLI">

```bash
opencli admin update <username> --plan=<package>
```

  </TabItem>
  <TabItem value="API-reseller-change-package" label="With API">

```bash
curl -X POST https://SERVER:2087/api/settings/resellers \
  -H "Authorization: Bearer TOKEN_HERE" -H "Content-Type: application/json" \
  -d '{"action":"change_plan","username":"reseller1","plan":"reseller_pro"}'
```

  </TabItem>
</Tabs>

Limits changed by hand on **Edit Plans & Limits** stay until the reseller's package is edited or another package is applied, then they're replaced by the package's limits.

## How limits are enforced

- **Max Accounts** - the reseller can't create a new user, and no user can be moved to them, once they have this many accounts.
- **User Packages** - users of the reseller can only be created on, or changed to, the allowed packages. This also applies when an administrator does it for the reseller.
- **Max Disk Usage** - a soft limit. Once the reseller's users together reach it, the reseller can't create new users, users can't be moved to them, and their users can't be changed to a package with more disk space. Existing accounts keep working. Disk usage is recalculated by the [files-calculate_resellers_storage](/docs/articles/server/openadmin-system-crons/#files-calculate_resellers_storage) cron job, so it can take a while to catch up.
