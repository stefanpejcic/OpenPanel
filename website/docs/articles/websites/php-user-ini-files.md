---
sidebar_label: "PHP settings per website (folder)"
---

# How to Set PHP Settings per Website or Folder (.user.ini)

Using `.user.ini` files, you can set different PHP limits for each website or folder.

:::warning
`.user.ini` files work with the **Apache**, **Nginx** and **OpenResty** web servers. They are not supported with **OpenLiteSpeed**, which runs PHP through its own LiteSpeed PHP (LSAPI) and ignores `.user.ini` files and `php_value` lines in `.htaccess`. On OpenLiteSpeed, change the limits for the whole PHP version in [PHP > Options](/docs/panel/php/options/) instead, or switch to another web server.
:::

## Creating a `.user.ini` file

For example, to set custom PHP limits only for a specific website, such as `example.net`:

1. Open **OpenPanel > Files > File Manager** and navigate to the domain's document root (the main folder for that site).
2. Click **New File**, enter `.user.ini` as the name, check **Open in File Editor after creation**, and then click **Create**.
3. In the editor, add the PHP configuration options you want to change. For example:

   ```
   upload_max_filesize = 64M
   post_max_size = 64M
   max_execution_time = 120
   ```
4. Save the file and exit the editor.

Finally, open your website and confirm the changes:

* For any PHP site on Apache, Nginx or OpenResty, you can check via `phpinfo()`.
* For WordPress, go to **Tools > Site Health** and verify the new limits are applied.

---

> **NOTE**: Since the `.user.ini` is read from public directories, it's contents will be served to anyone requesting it and potentially show them sensitive configuration settings. Block access to it using .htaccess file in Apache, or via Vhost Editor for Nginx/OpenResty.
