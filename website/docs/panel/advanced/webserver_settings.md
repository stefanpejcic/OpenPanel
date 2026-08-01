---
sidebar_position: 3
---

# WebServer Settings

The **WebServer Settings** interface enables OpenPanel users to edit the primary webserver configuration file directly using a text editor.

These settings are **global** and apply to **all domains** (i.e., affect the virtual hosts configuration).

> ⚠️ **Important:** Always create a backup of the configuration file before making any changes.

- For **Apache**, you can edit the `httpd.conf` file.  
- For **Nginx** and **OpenResty**, you can edit the `nginx.conf` file.
- For **OpenLitespeed** and **Litespeed**, you can edit the `openlitespeed.conf` file.

After making your changes, click **Save Changes** to write the file. If the webserver container is running, the new configuration is tested before it is applied - if it fails the syntax check, the previous content is restored and the save is rejected. When the configuration is valid (or the container isn't running), the file is saved and the webserver container is restarted automatically.

If a default configuration template is available for your webserver, a **Restore Default** button is also shown, letting you overwrite the current file with the stock configuration and restart the service.
