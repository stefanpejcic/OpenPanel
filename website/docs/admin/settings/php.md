---
sidebar_position: 6
---

# PHP Settings

Manage the default PHP version and configuration files for new user accounts.

### Default Version

The default PHP version used for new user accounts is not set on this page — it is configurable from the [Edit User Defaults](/docs/admin/settings/defaults/) page.

![Default version section linking to the User Defaults page](/img/openadmin-screenshots/settings/php-default.png#gh-light-mode-only)
![Default version section linking to the User Defaults page](/img/openadmin-screenshots/settings/php-default_dark.png#gh-dark-mode-only)

### Available Options

These options determine which PHP settings users can modify from their **OpenPanel > PHP Options** page.

![Available Options section with the PHP options users can edit](/img/openadmin-screenshots/settings/php-options.png#gh-light-mode-only)
![Available Options section with the PHP options users can edit](/img/openadmin-screenshots/settings/php-options_dark.png#gh-dark-mode-only)

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

![Default PHP.INI Files section with a php.ini per PHP version and Restore Default and Save buttons](/img/openadmin-screenshots/settings/php-ini.png#gh-light-mode-only)
![Default PHP.INI Files section with a php.ini per PHP version and Restore Default and Save buttons](/img/openadmin-screenshots/settings/php-ini_dark.png#gh-dark-mode-only)

Select a PHP version to open its php.ini file for editing. After making changes, click **Save** to apply them.

![Default PHP.INI Files section with the PHP 8.5 INI file expanded in an editable text area](/img/openadmin-screenshots/settings/php-ini-open.png#gh-light-mode-only)
![Default PHP.INI Files section with the PHP 8.5 INI file expanded in an editable text area](/img/openadmin-screenshots/settings/php-ini-open_dark.png#gh-dark-mode-only)

To revert to the original PHP.INI settings, click **Restore Default**, then click **Save**.
