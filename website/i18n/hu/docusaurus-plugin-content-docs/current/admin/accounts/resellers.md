---
sidebar_position: 2
---

# Resellers

The **Resellers** feature allows administrators to create and manage Reseller users within OpenPanel. Resellers can act as sub-administrators, managing their own set of user accounts within the limits defined by the root administrator.

This feature is useful for hosting providers who want to delegate control to third-party resellers, while still maintaining overall control and isolation.

---

## Reseller Management Interface

The interface displays a table of existing reseller users with the following columns:

- **Username**  
  The unique identifier of the reseller user.

- **Status**  
  Indicates whether the reseller account is active or suspended.

- **Last Login IP**  
  The IP address from which the reseller last accessed the panel.

- **Last Login Time**  
  Timestamp of the last successful login by the reseller.

- **Current Accounts**  
  The number of user accounts currently managed by the reseller.

- **Max Accounts**  
  The maximum number of user accounts the reseller is allowed to create.

- **Allowed Plans**  
  Lists the hosting plans available to the reseller for assigning to their users.

---

Reseller users have access only to the features and account management tools assigned to them by the root administrator. They cannot exceed the limits defined in their reseller settings.
