---
sidebar_position: 3
---

# Options  

The PHP Options page allows you to modify PHP settings stored in the `php.ini` file.

![PHP 8.5 Options page with editable values such as memory_limit, max_execution_time and upload limits](/img/openpanel-screenshots/php/options-page.png#gh-light-mode-only)
![PHP 8.5 Options page with editable values such as memory_limit, max_execution_time and upload limits](/img/openpanel-screenshots/php/options-page_dark.png#gh-dark-mode-only)

For more advanced users, the [**PHP.INI Editor**](/docs/panel/php/php_ini_editor/) can be used - if available.

---

## How to Change PHP Options

1. Open **OpenPanel** and navigate to **PHP > Options**.
2. Select the desired **PHP version** from the dropdown.
3. Modify the options you want to change.
4. Click **Save** to apply changes.

Settings will be saved to the `php.ini` file for the selected PHP version, and the PHP container will restart to apply changes immediately. These settings apply to **all domains** using that PHP version.

---

## Available Options

The list of configurable PHP options is determined by the system administrator. By default, the following options are available:

- `allow_url_fopen`
- `date.timezone`
- `disable_functions`
- `display_errors`
- `error_reporting`
- `file_uploads`
- `log_errors`
- `max_execution_time`
- `max_input_time`
- `max_input_vars`
- `memory_limit`
- `open_basedir`
- `output_buffering`
- `post_max_size`
- `short_open_tag`
- `upload_max_filesize`
- `zlib.output_compression`

**Customizing Available Options**

Administrators can customize the available options:

- **For all new users**: Edit `/etc/openpanel/php/options.txt`
- **For a specific user**: Edit `/home/USERNAME/php.ini/options.txt`

