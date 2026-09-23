---
sidebar_position: 6
---

# PHP Settings

Manage the default PHP version and configuration files for new user accounts.

### Default Version

The default PHP version used for new user accounts is not set on this page — it is configurable from the [Edit User Defaults](/docs/admin/settings/defaults/) page.

![Default version section linking to the User Defaults page](/img/openadmin-screenshots/settings/php-default.png)

### Available Options

These options determine which PHP settings users can modify from their **OpenPanel > PHP Options** page.

![Available Options section with the PHP options users can edit](/img/openadmin-screenshots/settings/php-options.png)

The default editable options include:

```
allow_url_fopen
date.timezone
disable_functions
display_errors
error_reporting
file_uploads
log_errors
max_execution_time
max_input_time
max_input_vars
memory_limit
open_basedir
output_buffering
post_max_size
short_open_tag
upload_max_filesize
zlib.output_compression
```

### Default PHP.INI Files

Here you can edit the PHP.INI configuration files that will be applied to new user accounts.

![Default PHP.INI Files section with a php.ini per PHP version and Restore Default and Save buttons](/img/openadmin-screenshots/settings/php-ini.png)

Select a PHP version to open its php.ini file for editing. After making changes, click **Save** to apply them.

To revert to the original PHP.INI settings, click **Restore Default**, then click **Save**.
