# WISECP

OpenPanel Enterprise edition has billing integrations with [WHMCS](/docs/articles/extensions/openpanel-and-whmcs/), [FOSSBilling](/docs/articles/extensions/openpanel-and-fossbilling/), [Blesta](/docs/articles/extensions/openpanel-and-blesta/), [ClientExec](/docs/articles/extensions/openpanel-and-clientexec/) and WISECP.

OpenPanel WISECP module allows users to integrate billing automations with their OpenPanel server.

# OpenPanel

To setup WISECP to use your OpenPanel server follow these steps:

## Enable API

First make sure that API access is enabled by going to `OpenAdmin > API` or by running `opencli config get api` from the terminal:

If API is not enabled, click on the "Enable API access" button or from terminal run
```bash
opencli config update api on
```

We recommend creating new Administrator user for the module, to create a new user navigate to *OpenAdmin > OpenAdmin Settings* and create new admin user, or from terminal run:
```bash
opencli admin new USERNAME_HERE PASSWORD_HERE
```

## Whitelist on OpenPanel

On OpenPanel server make sure that the OpenAdmin port 2087 is open on `OpenAdmin > Firewall` or whitelist the IP address of your WISECP server.
to whitelist ip address from terminal run:

```bash
csf -a WISECP_IP_HERE
```
or if using UFW:
```bash
ufw allow from WISECP_IP_HERE
```

## Create hosting package

Hosting packages need to be created on both OpenPanel and WISECP servers.
On OpenPanel server login to admin panel and on `OpenAdmin > Plans` create hosting packages that you will be assigning to users on WISECP.

# WISECP

## Requirements

- Server with OpenPanel Enterprise license
- WISECP

## Install OpenPanel WISECP Module

Login to SSH for the WISECP server, navigate to the WISECP install directory (e.g. `/var/www/wisecp`), then clone this repository and copy its `coremio` folder into place:
```bash
git clone https://github.com/stefanpejcic/openpanel-wisecp-module.git /tmp/openpanel-wisecp-module
cp -r /tmp/openpanel-wisecp-module/coremio/modules/Servers/OpenPanel coremio/modules/Servers/
```

## Module Setup

In the WISECP admin panel, go to **Settings > Servers > Add New**, select **OpenPanel** as the module, then fill in the OpenAdmin admin username/password, hostname, and port (default `2087`).

Assign the server to a product, then open the product's module options and pick which **OpenPanel Plan** it should provision accounts into (the dropdown is populated live from `/api/plans` on the assigned server).

## Supported Operations

From the WISECP admin area, the module supports:

- Create a new account + add domain
- Suspend account
- Unsuspend account
- Change plan (upgrade/downgrade)
- Change password
- Terminate account
- View account information (plan, disk usage, domains) on the order page
- Login to OpenAdmin (root access)

From the WISECP client area, the module supports:

- Auto-login as user
- Disk usage / bandwidth limit display

## Create hosting package

On the WISECP server create hosting plans that match the package names created on the OpenPanel server, then assign the **OpenPanel** server to the product.

## Test

Create a new order to test the connection and confirm the OpenPanel account is provisioned.

## Updating the module

Login to SSH for the WISECP server, pull the latest files from the repository and copy `coremio/modules/Servers/OpenPanel` over the existing installed copy.

## Troubleshooting

On WISECP:
1. Failed API calls are recorded through `Modules::save_log()` — check **Settings > Others > Module Log** (or the `mod_logs` table) for the request URL, payload and response
2. Errors returned by the module are shown directly on the relevant admin action

On OpenPanel server:
1. [Enable `DEV_MODE` on OpenAdmin](https://dev.openpanel.com/cli/config.html#dev-mode)
2. Send requests from the WISECP server
3. View the logs in: `/var/log/openpanel/admin/api.log`

## Bug Reports

Report a [new issue on GitHub](https://github.com/stefanpejcic/openpanel-wisecp-module/issues/new/choose).
