---
sidebar_position: 9
---

# phpMyAdmin

phpMyAdmin is an advanced MySQL database management tool and is recommended only for experienced users.

![databases_phpmyadmin.png](/img/panel/v1/databases/databases_phpmyadmin.png)

To access phpMyAdmin click on **Databases > phpMyAdmin** link in the sidebar. You will be automatically logged into the phpMyAdmin interface where you can view all existing databases and their tables.

:::warning
Editing and manipulating tables directly from phpMyAdmin could break your site if not done correctly. If you are not comfortable doing this, please check with a developer first.
:::

Once you’re logged in to phpMyAdmin, you can view your database tables, run queries, drop tables, import data, export your WordPress database, and more.

For more information about using phpMyAdmin, refer to [the official phpMyAdmin documentation](https://www.phpmyadmin.net/docs/).

----

:::info
phpMyAdmin runs on PHP version that user defines as [default php version](/docs/panel/advanced/server_settings#set-default-php-version) in OpenPanel. But due to phpmyadmin minimum requirements, if PHP version is less than 8.0 then the [default php version defined by the administrator](/docs/admin/settings/openpanel) will be used instead. To increase limits for the phpmyadmin interface, [edit the php.ini file](/docs/panel/advanced/server_settings#phpini-editor) of the version it is running on.
:::
