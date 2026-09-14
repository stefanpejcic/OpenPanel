# ClientExec

OpenPanel Enterprise edition has billing integrations with [WHMCS](/docs/articles/extensions/openpanel-and-whmcs/), [FOSSBilling](/docs/articles/extensions/openpanel-and-fossbilling/), [Blesta](/docs/articles/extensions/openpanel-and-blesta/) and ClientExec.

OpenPanel ClientExec plugin allows users to integrate billing automations with their OpenPanel server.

# OpenPanel

To setup ClientExec to use your OpenPanel server follow these steps:

## Enable API

First make sure that API access is enabled by going to `OpenAdmin > API` or by running `opencli config get api` from the terminal:

If API is not enabled, click on the "Enable API access" button or from terminal run
```bash
opencli config update api on
```

We recommend creating new Administrator user for the plugin, to create a new user navigate to *OpenAdmin > OpenAdmin Settings* and create new admin user, or from terminal run:
```bash
opencli admin new USERNAME_HERE PASSWORD_HERE
```

## Whitelist on OpenPanel

On OpenPanel server make sure that the OpenAdmin port 2087 is open on `OpenAdmin > Firewall` or whitelist the IP address of your ClientExec server.
to whitelist ip address from terminal run:

```bash
csf -a CLIENTEXEC_IP_HERE
```
or if using UFW:
```bash
ufw allow from CLIENTEXEC_IP_HERE
```

## Create hosting package

Hosting packages need to be created on both OpenPanel and ClientExec servers.
On OpenPanel server login to admin panel and on `OpenAdmin > Plans` create hosting packages that you will be assigning to users on ClientExec.

# ClientExec

## Requirements

- Server with OpenPanel Enterprise license
- ClientExec

## Install OpenPanel ClientExec Plugin

Login to SSH for the ClientExec server, navigate to `path_to_clientexec/plugins/server` and run this command to create a new folder and in it download the plugin:
```bash
git clone https://github.com/stefanpejcic/openpanel-clientexec.git openpanel
```

## Plugin Setup

In ClientExec admin, add a new server and select the **OpenPanel** plugin, then fill in the OpenAdmin admin username/password, hostname, and port (default `2087`).

## Supported Operations

From the ClientExec admin area, the plugin supports:

- Create a new user account + add domain
- Suspend account
- Unsuspend account
- Change package
- Change password
- Terminate account
- Test connection

## Create hosting package

On the ClientExec server create hosting plans that match the package names created on the OpenPanel server, then assign the **OpenPanel** server to the product.

## Test

Create a new order to test the connection and confirm the OpenPanel account is provisioned.

## Updating the plugin

Login to SSH for the ClientExec server, navigate to `path_to_clientexec/plugins/server/openpanel` and run:
```bash
git pull
```

## Troubleshooting

On ClientExec:
1. Enable debug logging for the server module and reproduce the action
2. Check the module log in the ClientExec admin area

On OpenPanel server:
1. [Enable `DEV_MODE` on OpenAdmin](https://dev.openpanel.com/cli/config.html#dev-mode)
2. Send requests from the ClientExec server
3. View the logs in: `/var/log/openpanel/admin/api.log`

## Bug Reports

Report a [new issue on GitHub](https://github.com/stefanpejcic/openpanel-clientexec/issues/new/choose).
