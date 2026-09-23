# Delete Multiple Accounts

The `opencli user-delete` command lets you delete OpenPanel user accounts in bulk. **Use caution**-it's easy to accidentally delete the wrong users.

### Delete a Single User

```bash
opencli user-delete USERNAME
```

You’ll be prompted for confirmation. To skip confirmation, use the `-y` flag:

```bash
opencli user-delete USERNAME -y
```

### Delete All Users

To delete **all** users on the server:

```bash
opencli user-delete -all
```

By default, you'll be prompted to confirm each deletion. To skip all confirmations, use the `-y` flag:

```bash
opencli user-delete -all -y
```
