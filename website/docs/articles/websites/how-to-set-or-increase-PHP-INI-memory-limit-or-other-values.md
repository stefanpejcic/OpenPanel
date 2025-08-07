# How to set or increase PHP INI memory_limit or other values?

There are multiple ways to manage PHP settings in OpenPanel. This guide walks you through the available methods and when to use each one.

## For OpenPanel users

Depending on which modules are enabled, users can manage PHP settings using one or more of the following interfaces:

- **PHP Limits** – Accessible if the `php` module is enabled. Allows editing of basic PHP limits.
- **PHP.INI Editor** – Available when the `php_ini` module is enabled. Provides full access to edit the php.ini file.
- **PHP Options** – Available with the `php_options` module. Lets users edit only pre-configured options from the ini file.

### Setting the Default PHP Version

The default PHP version used for new domains can be configured via:

**OpenPanel > PHP > Default version**

1. Select your desired PHP version.
2. Click 'Change' to apply.

![change default](/img/panel/v2/openpanel_cahnge_default_php_version.gif)

### PHP Limits

The **PHP Limits** interface allows you to configure key PHP directives. You **must** set the following here:

- `max_execution_time`
- `max_input_time`
- `max_input_vars`
- `memory_limit`
- `post_max_size`
- `upload_max_filesize`

[PHP Limits Documentation](/docs/panel/php/limits)

![openpanel_edit_php_limits](/img/panel/v2/openpanel_edit_php_limits.gif)

> NOTE: Settings from the *PHP Limits* page override those made in *PHP.INI Editor* or *PHP Options*.

---

### PHP Options

This interface allows you to modify commonly used PHP settings—limited to those pre-configured by an Administrator.

[PHP Options Documentation](/docs/panel/php/options)

![php options page](/img/panel/v2/php_options.png)

---

### PHP.INI Editor

For full control, the **PHP.INI Editor** lets users directly modify the php.ini file for their selected PHP version.

[PHP.INI Editor Documentation](/docs/panel/php/php_ini_editor)

![openpanel_php_ini_editor](/img/panel/v2/openpanel_php_ini_editor.gif)

---

## For OpenAdmin users

Administrators can configure system-wide PHP defaults and manage what settings users can change.

### Default PHP version

Set the default PHP version for new users via:
[**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/)

1. Choose the default PHP version.
2. Click 'Save' to confirm.

---

### PHP Limits

Administrators can define default PHP limits available to users via:
[**OpenAdmin > Settings > User Defaults**](/docs/admin/settings/defaults/)

---

### PHP Options

Users can only modify options that have been explicitly allowed by the administrator. These keys are pre-defined and applied from the system `.ini` files.

Administrators can manage these options in two ways:

From the panel:
[**OpenAdmin > Settings > PHP Settings**](/docs/admin/settings/php/) page.

Or via terminal:
- **Global default (all users)**: `/etc/openpanel/php/options.txt`
- **User-specific settings**: `/home/USERNAME/php.ini/options.txt`

---

### PHP.INI Editor

Administrators can configure the default php.ini files for all PHP versions.

From the panel:
[**OpenAdmin > Settings > PHP Settings**](/docs/admin/settings/php/) page.

Or via terminal:
- **Global default (all users)**: `/etc/openpanel/php/ini/`
- **User-specific settings**: `/home/USERNAME/php.ini/`

