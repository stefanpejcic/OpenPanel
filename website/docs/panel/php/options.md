---
sidebar_position: 3
---

# Options  

The PHP Options page lets you modify the PHP version settings by editing the PHP.INI file for that version and restarting the service to immediately apply the changes to all websites using it.

Options are limited to the options that Administrator permitted.

![ph options page](/img/panel/v2/php_options.png)

For more advanced users, the PHP.INI Editor can be used, if available.

---

Options are limited to the options that Administrator permitted, only configured keys from the .ini file will be applied.

Administrators can configure options:

- for all users: `/etc/openpanel/php/options.txt` , if these are set by the administrator they are used for all users .
- for a specific user: `/home/USERNAME/php.ini/options.txt` , if global options are not set by the administrator these will be applied instead.
- if per user options file does not exist, system will fallback to these defaults: https://github.com/stefanpejcic/openpanel-configuration/blob/main/php/options.txt
