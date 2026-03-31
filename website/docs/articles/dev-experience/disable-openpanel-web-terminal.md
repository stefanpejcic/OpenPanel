# Disable Terminal within OpenPanel

If for some reason, you want to disable command-line access through **Openpanel > Containers > Terminal** page, you can do that on your server.

There is no setting in OpenAdmin to do so, but you can manually create a file that will disable this feature:

1. Log in to the server as root via SSH.
2. Execute the following command:
   ```bash
   touch /etc/openpanel/disable_openpanel_terminal_ui
   ```



  
