# Installing OpenPanel

Before starting, ensure your server meets [the minimum requirements](https://openpanel.com/docs/admin/intro/#requirements) and runs a supported distribution.

To install OpenPanel on your server:

1. Log in to your new server;
- as root via SSH or
- as a user with sudo privileges and type "sudo -i"
2. Copy and paste openpanel installation command into the terminal
```shell
bash <(curl -sSL https://openpanel.org)
```

The installation script supports [optional flags](/install) that can be used to configure openpanel, skip certain installation steps or simply display debugging information.

If you encountered any errors while running the installation script, please copy & paste the installation log file to [the community forums](https://community.openpanel.org).
