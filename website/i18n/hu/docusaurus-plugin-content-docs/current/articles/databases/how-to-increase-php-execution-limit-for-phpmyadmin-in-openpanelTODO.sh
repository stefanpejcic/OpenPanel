# How to change PHP execution limit for phpMyAdmin in OpenPanel

phpMyAdmin runs on [PHP version that user sets as default php version](/docs/panel/advanced/server_settings/#set-default-php-version) in OpenPanel. 

But due to [phpmyadmin minimum requirements](https://docs.phpmyadmin.net/en/latest/require.html#php), if PHP version is less than 8.0 then [the default php version defined by the administrator](/docs/admin/settings/openpanel/#other-settings) will be used instead.

To increase limits for the phpMyAdmin interface, edit the php.ini file of the version it is running on.

### Step 1. Check PHP version

There are 2 ways to check which PHP version is the phpMyAdmin interface running:

1. From phpMyAdmin interface

Login to phpMyAdmin from the OpenPanel interface and note the version under 'Server Settings':

![phpmyadmin check php version](https://i.postimg.cc/bw1TsZKG/2024-08-02-19-10.png)

2. From OpenPanel > Process Manager

Login to OpenPanel and navigate to Process Manager to view under which PHP version is the phpMyAdmin interface running.

![process manager check php version](https://i.postimg.cc/2jXTJRQ5/2024-08-02-19-07.png)

### Step 2. Edit php.ini file

From OpenPanel navigate to Server Settings > PHP > PHP.INI editor and select that PHP version. Then make the changes, for example increase PHP `max_execution_time` value, and click on save to apply changes.

![phpini editor openpanel](https://i.postimg.cc/z8BckpCB/2024-08-02-19-09.png)

### Step 3. Kill phpMyAdmin processes

Since phpMyAdmin is running in its own webserver, we need to terminate its current processes in order to start it again with the new limits applied.

To do this, navigate to OpenPanel > Process Manager and kill all processes that mention phpMyAdmin.

![openpanel process kill](https://i.postimg.cc/KvkQhqfD/2024-08-02-19-08.png)

### Step 4. Start phpMyAdmin

Finally, start phpMyAdmin again with new limits, by opening phpMyAdmin from the OpenPanel menu.

![phpmyadmin menu item in openpanel](https://i.imgur.com/c2nA3vg.png)

Limits have been increased and you can proceed.
