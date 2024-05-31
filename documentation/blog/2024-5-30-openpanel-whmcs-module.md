---
title: OpenPanel WHMCS Module
description: OpenPanel is now integrated in WHMCS Module 🎉
slug: openpanel-whmcs-module
authors: stefanpejcic
tags: [OpenPanel, news, tutorial]
image: https://openpanel.co/img/blog/openpanel-whmcs-module.png
hide_table_of_contents: true
---

OpenPanel WHMCS module allows users to integrate billing automations with their OpenPanel server.

<!--truncate-->

WHMCS module is now available for OpenPanel.
Currently supported actions are:

- ✅ create account
- ✅ change password
- ✅ suspend account
- ✅ unsuspend account
- ✅ terminate account
- ✅ autologin from frontend
- ✅ autologin from backend
- ❌ get disk usage for account
- ✅ change package

To setup WHMCS to use your OpenPanel server follow these steps:

1. **Enable API access on OpenPanel server**
   First make sure that API access is enabled by going to `OpenAdmin > API` or by running `opencli config get api` from the terminal:
   ![enable_api](https://i.postimg.cc/L6vwMQ4t/image.png)
   If API is not enabled, click on the "Enable API access" button or from terminal run `opencli config update api on`.

   We recommend creating new Administrator user for API, to create a new user navigate to *OpenAdmin > OpenAdmin Settings* and create new admin user, or from terminal run: `opencli admin new USERNAME_HERE PASSWORD_HERE`

2. **Install OpenPanel WHMC Module**
    Login to SSH for WHMCS server
    Navigate to `path_to_whmcs/modules/servers`
    Run this command to create a new folder and in it download the module: `git clone https://github.com/stefanpejcic/openpanel-whmcs-module.git openpanel`

3. **Establish connection between the two servers**
   On OpenPanel server make sure that the OpenAdmin port 2087 is open on `OpenAdmin > Firewall` or whitelist the IP adress of your WHMCS server.
   to whitelist ip address from terminal run: `ufw allow from WHMCS_IP_HERE`

   On WHMCS server also make sure that the 2087 port is opened or whitelist the IP address of your OpenPanel server.
   From WHMS navigate to: *System Settings > Products & Services > Servers*
    ![screenshot](https://i.postimg.cc/MHWpL3tc/image.png)
    Click on *Create New Server* and under module select **OpenPanel** then add OpenPanel server IP, username and password for the OpenAdmin panel:
   ![create_whmcs_group](https://i.postimg.cc/3Jh3nqWY/image.png)
   
4. **Create hosting packages**
  Hosting packages need to be created on both OpenPanel and WHMCS servers.
  On OpenPanel server login to admin panel and on `OpenAdmin > Plans` create hosting packages that you will be assinging to users on WHMCS.

  On the WHMCS server create first a new group and then create new plans under this group. When creating products, make sure to select OpenPanel for Module and the newly created group
  ![screenshot2](https://i.postimg.cc/NLvF4GSc/image.png)

5. **Test creating new accounts**
   Create an order and create a new order to test OpenPanel API.
