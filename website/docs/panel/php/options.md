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
4. Click **Save Changes** to apply changes. Unticking an On/Off option saves it as `Off` (or `0`).

Settings will be saved to the `php.ini` file for the selected PHP version, and the PHP container will restart to apply changes immediately. These settings apply to **all domains** using that PHP version.

---

## Optimize PHP

When you open the options of a PHP version, OpenPanel checks that version in the background and suggests values that fit your sites. While it checks you'll see *Checking this PHP version for possible optimizations...*.

If it finds something to improve, an **Optimize PHP** banner shows at the top of the page and an optimize icon (⚡) shows next to each option that has a suggestion.

1. Click the optimize icon next to an option, or **Review all** in the banner.
2. The **Confirm Changes** dialog shows the previous value, the new value and why it's suggested for each option.
3. Click **Save Changes** to apply them and restart the PHP service, or **Cancel** to close the dialog without changing anything.

The banner also shows what the suggestions are based on: the PHP version, the memory limit and current memory use of the PHP service, how many of your domains use this version, how many of them run WordPress and how many PHP workers the service can run at once.

Suggestions are based on:

- **The PHP service**: its memory limit and whether it was stopped for running out of memory. For example `memory_limit` is lowered when a single script may use more memory than the whole service has, or when it's unlimited (`-1`).
- **Your domains**: how many use this PHP version and how many of them are WordPress sites. WordPress sites get suggestions such as `memory_limit = 256M`, `upload_max_filesize = 64M` and `max_input_vars = 3000`.
- **Your settings**: for example `post_max_size` smaller than `upload_max_filesize`, `max_execution_time = 0`, `display_errors` on, `log_errors` off or an empty `date.timezone`.
- **The PHP-FPM log**: *Allowed memory size exhausted*, *Maximum execution time exceeded*, *POST Content-Length exceeds the limit* and *Input variables exceeded* errors raise the matching option. Running out of PHP workers is shown as a note.

If no domain uses the selected PHP version yet, the banner says so and only suggestions based on settings and memory are shown.

:::info
Optimizations are general suggestions and may not result in better performance for all sites.
:::

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

