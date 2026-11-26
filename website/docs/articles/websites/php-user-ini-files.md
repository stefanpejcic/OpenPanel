---
sidebar_label: "PHP settings per website (folder)"
---

# How to Set PHP Settings per Website or Folder (.user.ini)

Using `.user.ini` files, you can set different PHP limits for each website or folder. The other websites that use the same PHP version keep the settings from [PHP > Options](/docs/panel/php/options/).

:::warning
`.user.ini` files work with the **Apache**, **Nginx** and **OpenResty** web servers. They are not supported with **OpenLiteSpeed**, which runs PHP through its own LiteSpeed PHP (LSAPI) and ignores `.user.ini` files and `php_value` lines in `.htaccess`. On OpenLiteSpeed, change the limits for the whole PHP version in [PHP > Options](/docs/panel/php/options/) instead, or switch to another web server.
:::

## Creating a `.user.ini` file

For example, to set custom PHP limits only for a specific website, such as `example.net`:

1. Open **OpenPanel > Files > File Manager** and navigate to the domain's document root (the main folder for that site).
2. Click **New File**, enter `.user.ini` as the name, check **Open in File Editor after creation**, and then click **Create**.
3. In the editor, add the PHP configuration options you want to change. For example:

   ```ini
   upload_max_filesize = 64M
   post_max_size = 64M
   max_execution_time = 120
   ```
4. Save the file and exit the editor.

## When the changes apply

PHP caches `.user.ini` files for up to **5 minutes** (the `user_ini.cache_ttl` setting), so new values can take a few minutes to show up.

To apply them right away, open **PHP > PHP.INI Editor**, select the PHP version the domain uses and click **Save**. This restarts that PHP service.

Then open your website and confirm the changes:

* For any PHP site on Apache, Nginx or OpenResty, you can check via `phpinfo()`.
* For WordPress, go to **Tools > Site Health** and verify the new limits are applied.

## Which folders it applies to

A `.user.ini` file applies to PHP scripts in the folder where it is placed **and in all of its subfolders**. PHP looks for `.user.ini` files from the document root down to the folder of the script that runs.

If a subfolder has its own `.user.ini`, its values are applied after the parent's values, so the subfolder's values win. For example, a `.user.ini` in `example.net/` that sets `upload_max_filesize = 64M` and one in `example.net/shop/` that sets `upload_max_filesize = 256M` give `64M` to the main site and `256M` to the shop.

## What you can set

Only settings that PHP allows to change per folder work in `.user.ini`. In the [PHP manual's list of ini settings](https://www.php.net/manual/en/ini.list.php), these are the ones marked `PHP_INI_ALL` or `PHP_INI_PERDIR` in the **Changeable** column.

The most commonly used ones:

| Area | Examples |
|------|----------|
| Limits | `memory_limit = 512M`, `max_execution_time = 300`, `max_input_time = 300`, `max_input_vars = 5000` |
| Uploads | `upload_max_filesize = 128M`, `post_max_size = 128M`, `max_file_uploads = 50` |
| Errors | `display_errors = Off`, `log_errors = On`, `error_log = /var/www/html/example.net/php_errors.log`, `error_reporting = E_ALL & ~E_DEPRECATED` |
| Time and charset | `date.timezone = Europe/Berlin`, `default_charset = "UTF-8"` |
| Sessions | `session.gc_maxlifetime = 7200`, `session.cookie_secure = 1`, `session.cookie_httponly = 1`, `session.cookie_samesite = Lax` |
| Security | `open_basedir = /var/www/html/example.net:/tmp`, `allow_url_fopen = Off` |
| Output | `output_buffering = 4096`, `zlib.output_compression = On`, `auto_prepend_file = /var/www/html/example.net/prepend.php` |
| OPcache | `opcache.enable = 0`, `opcache.revalidate_freq = 0` |

:::tip
`opcache.enable = 0` is useful for applications that ask for OPcache to be disabled, such as WHMCS. It turns OPcache off only for that website, while all other websites on the same PHP version keep the speed benefit. You can turn OPcache **off** from `.user.ini`, but not back on if it is disabled for the PHP version.
:::

## What you cannot set

Settings marked `PHP_INI_SYSTEM` are ignored in `.user.ini`. They can only be changed for the whole PHP version in [PHP > Options](/docs/panel/php/options/) or the [PHP.INI Editor](/docs/panel/php/php_ini_editor/). These include:

* Loading extensions: `extension`, `zend_extension` (to enable extensions such as ionCube, use [PHP > Extensions](/docs/panel/php/extensions/)).
* Disabling functions and classes: `disable_functions`, `disable_classes`.
* OPcache memory and cache size: `opcache.memory_consumption`, `opcache.jit_buffer_size`, `opcache.max_accelerated_files`.
* The `.user.ini` settings themselves: `user_ini.filename`, `user_ini.cache_ttl`.
* PHP-FPM pool settings, such as the number of PHP processes. These are not ini settings at all.

Also, if a value is set with `php_admin_value` or `php_admin_flag` in the domain's VirtualHost, `.user.ini` cannot override it.
