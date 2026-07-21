# How to Install Custom or Older IonCube Loader Versions in OpenPanel

Starting with **OpenPanel version 1.7.2**, IonCube Loader is automatically available for all PHP versions that support it on new installations.

## Older Versions

If you want to downgrade to an older ionCube Loader bundle (or use a custom bundle) you can do so easily.

After placing the files on your server:

* **For a single user:** Edit their `docker-compose.yml` file and update the volume mount points for the **php-fpm-*** services so they reference your custom files.
* **For all new users:** Edit the template file located at `/etc/openpanel/docker/compose/1.0/docker-compose.yml`.

## Check if enabled

To confirm whether IonCube Loader is active for your PHP version, you can use any of the following methods:

From terminal run:

```bash
php -i | grep ioncube
```

Or create a file named **info.php** in your domain’s public directory with the following content:

```php
<?php
phpinfo();
```

Then open it in your browser:
`https://yourdomain.tld/info.php`

Look for **IonCube PHP Loader** in the output to verify that it is enabled.
