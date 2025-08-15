# How to Install WordPress® With OpenPanel

[WordPress](https://wordpress.org/)® is a powerful, web-based content management system (CMS) that makes it easy to build websites and blogs. This guide walks you through two methods for installing WordPress on your OpenPanel account: **automatic installation via WP Manager** and **manual installation**.

Before installing WordPress, make sure you've added a domain where the WordPress site will be hosted.

To add a domain:
**OpenPanel > Domains > Add New Domain**

![newDomain.png](/img/panel/v2/wpgDomain.png)

## Install WordPress via WP Manager

OpenPanel's Site Manager lets you install WordPress in just a few clicks.

1. **Open Site Manager** and press the **+New Website"** button.
   ![SiteManager_1.png](/img/panel/v2/wpgSitemanager1.png)
2. Click **Install WordPress**.
   ![SiteManager_2.png](/img/panel/v2/wpgSitemanager2.png)
3. Fill in the **Website Name**, optionally add a **Site Description**, and choose your domain.
   Click **Start Installation** to begin.
   ![SiteManager_3.png](/img/panel/v2/wpgSitemanager3.png)
5. Once the installation is complete, you'll be redirected to the **WP Manager**, where you can manage all of your installed WordPress sites.
   ![SiteManager_4.png](/img/panel/v2/wpgSitemanager4.png)

## Install WordPress manually

The easiest way to download the official WordPress installation archive is through OpenPanel's File Manager tool.

First access the folder of your domain where you'll install a new WordPress instance.

![Manually_1.png](/img/panel/v2/wpgManual1.png)

When you're inside the domains directory press the Upload button on the top right.

![Manually_2.png](/img/panel/v2/wpgManual2.png)

On the Upload page press the "Download from URL Instead" button.

![Manually_3.png](/img/panel/v2/wpgManual3.png)

Paste the link https://wordpress.org/latest.zip to the URL field and press the "Download" button.

![Manually_4.png](/img/panel/v2/wpgManual4.png)

![Manually_5.png](/img/panel/v2/wpgManual5.png)

When the download finishes return to File Manager and extract the archive.

![Manually_6.png](/img/panel/v2/wpgManual6.png)

Access the "wordpress" folder, select All files using the "Select All" button and move them into your new domain's directory using the "Move" button.

![Manually_7.png](/img/panel/v2/wpgManual7.png)

![Manually_8.png](/img/panel/v2/wpgManual8.png)

![Manually_9.png](/img/panel/v2/wpgManual9.png)

![Manually_10.png](/img/panel/v2/wpgManual10.png)

Create a new database and database user with the Database Wizard tool, more info in our Database Wizard guide -> https://openpanel.com/docs/panel/mysql/wizard/ .

![Manually_11.png](/img/panel/v2/wpgManual11.png)

Start editing the wp-config-sample.php file by selecting it and pressing the "Edit" button.

![Manually_12.png](/img/panel/v2/wpgManual12.png)

Swap the vaules of DB_NAME, DB_USER and DB_PASSWORD with values you've set inside the Database Wizard, set the value of DB_HOST to 'mysql' and save the changes on the top right.

![Manually_config.png](/img/panel/v2/wpgManualFinal.png)

Finally, rename the wp-config-sample.php file to wp-config.php using the Rename button within the File Manager.

![Manually_configR.png](/img/panel/v2/wpgManualRename.png)

![Manually_configR2.png](/img/panel/v2/wpgManualRename2.png)

Your new WordPress instance is now created!

Finish the installation by accessing your domain via browser and completing the WordPress web installation wizard.

