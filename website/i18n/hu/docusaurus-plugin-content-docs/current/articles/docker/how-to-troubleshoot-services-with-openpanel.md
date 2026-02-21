# Troubleshooting service issues with OpenPanel

If you've run into issues with a service that is managed by OpenPanel and your account has the Docker feature enabled you can use it to restart the service container and check the logs for further troubleshooting.

Simpy go to Openpanel > Docker > Containers and try stopping/starting the service container:

![containersGUI.png](https://i.postimg.cc/650sKYW3/containersgui.png)

To check logs go to Openpanel > Docker > Logs and select a container from the dropdown menu to view it's logs:

![containersLogsGUI.png](https://i.postimg.cc/Fzh4bqF1/containerlogsgui.png)

If the Docker feature is not enabled for your account contact your administrator so they can check further.

## Troubleshooting service issues as an Administrator 

Login to OpenAdmin on port 2087 , go to Users and click on an username start managing that account:

![containersAdmin1.png](https://i.postimg.cc/rpYjNt8p/containers-Admin1.png)

Within the user management page go to the Services tab and Enable/Disable the service:

![containersAdmin2.png](https://i.postimg.cc/hv4G4hdV/containers-Admin2.png)

If restarting the container doesn't solve the issue you'll need to troubleshoot further, access the server via SSH and investigate by running the following commands:

`machinectl shell $username@ /bin/bash -c 'systemctl --user status'` (replace $username with the OpenPanel user)

A status overview of the user’s own systemd service manager will be displayed:

![containersAdminSSH1](https://i.postimg.cc/FKRzQY0S/containers-Admin-SSH1.png)

`machinectl shell $username@ /bin/bash -c 'journalctl --user -u $docker'` (replace $username with the OpenPanel user)

The Docker service logs will be displayed:

![containersAdminSSH2](https://i.postimg.cc/6qjwc0mJ/containers-Admin-SSH2.png)

In case you don't find any errors or potential causes within the output of these commands you should start the service container manually and watch the logs with the command:

`cd /home/$username && docker --context=$username compose up SERVICE_NAME` (replace $username with the OpenPanel user and SERVICE_NAME with the name of the service you are troubleshooting)


