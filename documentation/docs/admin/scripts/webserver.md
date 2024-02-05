---
sidebar_position: 7
---

# Apache / Nginx

Scripts for managing users webserver: Nginx or Apache


## Get users webserver

To list users current webserver run the following command:

```bash
opencli webserver-get_webserver_for_user <USERNAME>
```

Example:
```bash
# opencli webserver-get_webserver_for_user stefan
Web Server for user stefan: apache
```

The script will by default display cached information from the users `server_config.yml` file, optionally you can add `--update` flag to check the current webserver in the container and update the file.

Example:
```bash
# opencli webserver-get_webserver_for_user stefan --update
Web Server for user stefan updated to: apache
```

## Fix Permissions

The `fixperms` script can be used to fix permissions on user files.

It performs:
- sets the owner of all files inside /home/$username to the user.
- sets the permissions of .php files to 755.
- sets the permissions of .cgi and .pl files to 755.
- sets the permissions of .log files to 640.
- changes the ownership of all directories to match the user.
- sets the permissions of all directories to 755.

```bash
opencli webserver-fixperms <USERNAME>
```

You can pass the `--all` flag to change permissions for all users:

```bash
opencli webserver-fixperms --all
```

## Install ModSecurity

You can install modsecurity by using:

```bash
opencli nginx-install_modsec
```
