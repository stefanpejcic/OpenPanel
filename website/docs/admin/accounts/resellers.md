---
sidebar_position: 2
---

# Resellers

The **Resellers** feature allows administrators to create and manage Reseller users within OpenPanel. Resellers can act as sub-administrators, managing their own set of user accounts within the limits defined by the root administrator.

This feature is useful for hosting providers who want to delegate control to third-party resellers, while still maintaining overall control and isolation.

:::info
Resellers is an Enterprise-only feature.
:::

Resellers are managed from **Accounts > Resellers**. Reseller logins do not see this page - they instead manage their own account from **Reseller Account** at `/account`.

---

## Reseller Management Interface

The interface displays a table of existing reseller users with the following columns:

- **Username**  
  The unique identifier of the reseller user.

- **Status**  
  Indicates whether the reseller account is active or suspended.

- **2FA**  
  Whether the reseller has Two-Factor Authentication enabled.

- **Passkeys**  
  Whether the reseller has a passkey registered.

- **Last Login IP**  
  The IP address from which the reseller last accessed the panel.

- **Last Login Time**  
  Timestamp of the last successful login by the reseller.

- **Accounts**  
  The number of user accounts currently managed by the reseller, shown as current/maximum (for example `3/10`). The current count links to the reseller's users on the Users page, and the maximum links to the reseller's Edit Plans & Limits page.

- **Storage**  
  A progress bar showing how much of the reseller's disk usage allowance is currently used.

- **Hosting Plans**  
  The number of hosting plans available to the reseller for assigning to their users.

Each row also has an Edit menu with the available actions:

- **Edit Plans & Limits** - set the reseller's Max Accounts, Max Disk Usage (in blocks), and the hosting plans they're allowed to assign to their users.
- **Rename** - change the reseller's username.
- **Change Password** - set a new password for the reseller.
- **Disable 2FA** / **Disable Passkeys** - shown only when enabled for that reseller; lets the Super Admin remove it.
- **Suspend** / **Unsuspend** - temporarily disable the reseller's access.
- **Delete** - permanently remove the reseller account.

To create a new reseller, click **Create New**, set a username and password, and click **Create**.

---

Reseller users have access only to the features and account management tools assigned to them by the root administrator. They cannot exceed the limits defined in their reseller settings (Max Accounts, Max Disk Usage, and the Hosting Plans made available to them).
