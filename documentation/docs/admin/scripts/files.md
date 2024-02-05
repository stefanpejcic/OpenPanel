---
sidebar_position: 10
---

# Files

Scripts for managing users files.

## Fix Permissions

The `fix_permissions` script is usd to fix file permissions and owner for users files inside their home directory.

To fix permissions for all active users on the server:

```bash
opencli files-fix_permissions --all
```
For a single user:

```bash
opencli files-fix_permissions [USERNAME] [PATH]
```
