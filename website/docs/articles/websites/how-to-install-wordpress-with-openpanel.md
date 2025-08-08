# How to Install WordPress® With OpenPanel

WordPress®, a web-based content management system, allows users to easily create a website or blog. This document describes how to install WordPress on your OpenPanel account.

![newDomain.png](/img/panel/v2/wpgDomain.png)

Before installing WordPress make sure you've added a domain that the WordPress website will be using, you can add a new domain on OpenPanel > Domains > Add New Domain.

## Install WordPress from WP Manager

OpenPanel's Site Manager feature allows you to install a new WordPress instance in a few simple steps:

![SiteManager_1.png](/img/panel/v2/wpgSitemanager1.png)

Access the Site Manager and press the "+New Website" button to begin.

![SiteManager_2.png](/img/panel/v2/wpgSitemanager2.png)

Choose WordPress by pressing the "Install WordPress" button.

![SiteManager_3.png](/img/panel/v2/wpgSitemanager3.png)

Choose a Website name, Description and the domain of your new WordPress website and press the "Start Installation" button to run the installer, the install log will be displayed below.

![SiteManager_4.png](/img/panel/v2/wpgSitemanager4.png)

When the manager is finished installing your new WordPress website you'll be redirected to the WordPress manager page where you can view and manage all of your WordPress websites.

## Install WordPress manually

The easiest way to download the official WordPress installation archive is through OpenPanel's File Manager tool.

![Manually_1.png](/img/panel/v2/wpgManual1.png)

First access the folder of your domain where you'll install a new WordPress instance.

![Manually_2.png](/img/panel/v2/wpgManual2.png)

When you're inside the domains directory press the Upload button on the top right.

![Manually_3.png](/img/panel/v2/wpgManual3.png)

On the Upload page press the "Download from URL Instead" button.

![Manually_4.png](/img/panel/v2/wpgManual4.png)

![Manually_5.png](/img/panel/v2/wpgManual5.png)

Paste the link https://wordpress.org/latest.zip to the URL field and press the "Download" button.

![Manually_6.png](/img/panel/v2/wpgManual6.png)

When the download finishes return to File Manager and extract the archive.

![Manually_7.png](/img/panel/v2/wpgManual7.png)

![Manually_8.png](/img/panel/v2/wpgManual8.png)

![Manually_9.png](/img/panel/v2/wpgManual9.png)

![Manually_10.png](/img/panel/v2/wpgManual10.png)

Access the "wordpress" folder, select All files using the "Select All" button and move them into your new domain's directory using the "Move" button.

![Manually_11.png](/img/panel/v2/wpgManual11.png)

Create a new database and database user with the Database Wizard tool, more info in our Database Wizard guide -> https://openpanel.com/docs/panel/mysql/wizard/ .

![Manually_12.png](/img/panel/v2/wpgManual12.png)

Start editing the wp-config-sample.php file by selecting it and pressing the "Edit" button.

![Manually_config.png](/img/panel/v2/wpgManualFinal.png)

Swap the vaules of DB_NAME, DB_USER and DB_PASSWORD with values you've set inside the Database Wizard, set the value of DB_HOST to 'mysql' and save the changes on the top right.

![Manually_configR.png](/img/panel/v2/wpgManualRename.png)

![Manually_configR2.png](/img/panel/v2/wpgManualRename2.png)

Finally, rename the wp-config-sample.php file to wp-config.php using the Rename button within the File Manager.

Your new WordPress instance is now created!

Finish the installation by accessing your domain via browser and completing the WordPress web installation wizard.

