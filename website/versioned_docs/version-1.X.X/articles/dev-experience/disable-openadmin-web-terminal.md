# Disable Terminal within OpenAdmin

If for some reason, you want to disable command-line *root* access through OpenAdmin's Terminal feature, you can do that on your server.

There is no setting in OpenAdmin to do so, but you can manually create a file that will disable this feature:

1. Log in to the server as root via SSH.
2. Execute the following command:
   ```bash
   touch /root/disable_openadmin_terminal_ui
   ```
