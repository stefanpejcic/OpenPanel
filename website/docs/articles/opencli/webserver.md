# Apache / Nginx

Scripts for managing users webserver: Nginx or Apache


### Get users webserver

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
