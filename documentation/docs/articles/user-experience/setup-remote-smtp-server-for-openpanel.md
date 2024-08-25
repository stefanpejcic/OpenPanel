# Setup Remote SMTP Server for OpenPanel

To set up a remote SMTP server for all emails sent from OpenPanel user websites, follow these steps:

## Whitelist the Remote SMTP Server

On the server with OpenPanel installed, ensure that the IP address of the remote SMTP server is whitelisted on [firewall](/docs/admin/security/firewall/).

Similarly, on the remote SMTP server, make sure to whitelist the IP address of the OpenPanel server.


## Configure SMTP per User

For each user, SMTP settings are configured via the `/etc/msmtprc` file. An example file content is shown below:

```
# /etc/msmtprc
defaults
auth       off
tls        off
logfile    /var/log/msmtp.log

account    default
host       openadmin_mailserver
port       25
from       USERNAME@HOSTNAME
```

[Login as root for that user](https://dev.openpanel.com/cli/commands.html#Login-as-User) and edit the file. In this file, set the email address and server to be used for all outgoing emails. 

For example, you can configure Gmail SMTP settings as shown in the [Arch Linux Wiki](https://wiki.archlinux.org/title/Msmtp).

After saving the configuration, test sending emails from a PHP website like WordPress or directly from the terminal.


## Configure SMTP for all existing users

Create a new file on the server, and in it copy the [default /etc/msmtprc file](https://github.com/stefanpejcic/OpenPanel/blob/main/docker/apache/email/msmtprc).

In the file set your remote SMTP server address and optionally logins.

When finished, run the bellow command to copy the file to each OpenPanel user:

```bash
  for user in $(opencli user-list --json | grep -v 'SUSPENDED' | awk -F'"' '/username/ {print $4}'); do
    echo "Processing user: $user"
    docker cp filename $user:/etc/msmtp
  done
```


## Configure SMTP automatically for new users

To automatically have all users pre-configured to use your remote SMTP server, [build a docker image](/docs/articles/docker/building_a_docker_image_example_include_php_ioncubeloader/) that has your custom [/etc/msmtprc](https://github.com/stefanpejcic/OpenPanel/blob/main/docker/apache/email/msmtprc) file.
