---
sidebar_label: "Delete Multiple Accounts"
---

# How to Delete Multiple Hosting Accounts at Once

The `opencli user-delete` command deletes OpenPanel user accounts, and can be combined with `opencli user-list` to delete them in bulk. **Use caution**-it's easy to accidentally delete the wrong users.

### Delete a Single User

```bash
opencli user-delete USERNAME
```

You’ll be prompted for confirmation. To skip confirmation, use the `-y` flag:

```bash
opencli user-delete USERNAME -y
```

### Delete Multiple Users

`opencli user-delete` takes one username at a time. To delete **all** users on the server, loop over the output of `opencli user-list --json`:

```bash
opencli user-list --json | jq -r '.data[].username' | while read -r u; do opencli user-delete "$u" -y; done
```

The `-y` flag skips the confirmation for each user, so check the list first with `opencli user-list`.
