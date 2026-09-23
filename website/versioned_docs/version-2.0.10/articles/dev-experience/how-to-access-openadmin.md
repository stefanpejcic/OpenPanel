# Access OpenAdmin

Run `opencli admin` command to find the address on which admin panel is accessible. Example output:

```bash
root@server:/home# opencli admin
● OpenAdmin is running and is available on: https://server.openpanel.org:2087/
```

To login to admin panel you need a username and password.

![openadmin login page](/img/admin/openadmin_login_page.png)

Both username and password are random generated on installation.

To view admin accounts:

```bash
opencli admin list
```

To set a new password for the admin account run command: `opencli admin password USER_HERE NEW_PASSWORD_HERE`

Example:
```bash
root@server:/home# opencli admin password stefan ba63vfav7fq36vas
Password for user 'stefan' changed.

===============================================================
● OpenAdmin is running and is available on: https://server.openpanel.co:2087/

- username: stefan
- password: ba63vfav7fq36vas

===============================================================
```
