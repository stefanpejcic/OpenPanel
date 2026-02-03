# OpenPanel WHMCS Module 😎
WHMCS module for [OpenPanel](https://openpanel.com)

![screenshot](/screenshots/account_manage.png)

---


## Usage

### Requirements

- Server with OpenPanel Enterprise license
- WHMCS

### Installation

1. Login to SSH for WHMCS server
2. Navigate to `path_to_whmcs/modules/servers`
3. Run this command to create a new folder and in it download the module:
   ```bash
   git clone https://github.com/stefanpejcic/openpanel-whmcs-module.git openpanel
   ```

### Supported Operations

From WHMCS admin area:
- Create a new user account + add domain
- Auto-login as user
- Suspend account
- Unsuspend account
- Change package
- Change password
- Terminate account
- View user information as admin

From WHMCS client area:
- Auto-login as user
- Change password

### Configuration

How to setup WHMCS and OpenPanel: https://openpanel.com/docs/articles/extensions/openpanel-and-whmcs/

###  Troubleshooting

On WHMCS:
1. set OPENPANEL_DEBUG to `true`
2. [Enable Moodule Logs](https://developers.whmcs.com/provisioning-modules/module-logging/)
3. Run actions in WHMCS
4. Check the log on `YOUR_WHMCS_ADMIN_URI/index.php?rp=/admin/logs/module-log`

On OpenPanel server:
1. [Enable `DEV_MODE` on OpenAdmin](https://dev.openpanel.com/cli/config.html#dev-mode)
2. Send requests from WHMCS server
3. View the logs in: `/var/log/openpanel/admin/api.log`


### Update

1. Login to SSH for WHMCS server
2. Navigate to `path_to_whmcs/modules/servers/openpanel`
3. Run this command to download newer files:
   ```bash
   git pull
   ```

## Bug Reports

Report [new issue on github](https://github.com/stefanpejcic/openpanel-whmcs-module/issues/new/choose)

