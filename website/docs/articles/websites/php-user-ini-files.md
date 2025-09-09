# PHP settings per website (folder)

Using `.user.ini` files, you can set different PHP limits for each website or folder.

## Creating a `.user.ini` file

For example, to set custom PHP limits only for a specific website, such as `example.net`:

1. Open **OpenPanel > File Manager** and navigate to the domain's document root (the main folder for that site).
2. Click **New File**, enter `.user.ini` as the name, check **Open in File Editor after creation**, and then click **Create**.
3. In the editor, add the PHP configuration options you want to change. For example:

   ```
   upload_max_filesize = 64M
   post_max_size = 64M
   max_execution_time = 120
   ```
4. Save the file and exit the editor.

Finally, open your website and confirm the changes:

* For any PHP site, you can check via `phpinfo()`.
* For WordPress, go to **Tools > Site Health** and verify the new limits are applied.

---

> **NOTE**: Since the `.user.ini` is read from public directories, it's contents will be served to anyone requesting it and potentially show them sensitive configuration settings. Block access to it using .htaccess file in Apache, or via Vhost Editor for Nginx/OpenResty.
