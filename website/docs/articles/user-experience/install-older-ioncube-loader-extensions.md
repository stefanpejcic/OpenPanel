# How to Install Custom or Older Ioncube Loader Versions in OpenPanel

:::info
This guide explains how to install custom or older Ioncube Loader versions for PHP in OpenPanel Docker-based services (php-fpm-8.x).  
The previous method using `/etc/openpanel/php/ioncube.txt` is outdated.
:::

Follow these steps to set up a custom Ioncube Loader:

1. **Download and extract the Ioncube Loader.**  
   Download the official bundle from [Ioncube Loaders](https://www.ioncube.com/loaders.php).  
   Extract it and locate the loader for your PHP version, for example: `ioncube_loader_lin_8.2.so`.  
   Place the file in a directory inside the user’s home, e.g.: `/home/USERNAME/ioncube/ioncube_loader_lin_8.2.so`.

2. **Start the user’s PHP service.**  
   ```bash
   docker --context=USERNAME compose up -d php-fpm-8.2

3. **Check the PHP extension directory.**
    ```bash
    docker --context=USERNAME exec php-fpm-8.2 bash -c "php -i | grep extension_dir"
    ```
    Example output: ```/usr/local/lib/php/extensions/no-debug-non-zts-20220829```

4. **Bind-mount the loader in docker-compose.yml.**
   Edit the user’s docker-compose.yml and add:
   ```bash
   services:
     php-fpm-8.2:
       volumes:
         - ./ioncube/ioncube_loader_lin_8.2.so:/usr/local/lib/php/extensions/no-debug-non-zts-20220829/ioncube_loader.so
    ```
5. **Create a persistent PHP configuration file for Ioncube.**
   ```bash
   echo "zend_extension=ioncube_loader.so" > /home/USERNAME/ioncube/docker-php-ext-ioncube.ini
   chown USERNAME:USERNAME /home/USERNAME/ioncube/docker-php-ext-ioncube.ini
    ```
   Then mount it in docker-compose.yml
   ```bash
   services:
     php-fpm-8.2:
       volumes:
         - ./ioncube/docker-php-ext-ioncube.ini:/usr/local/etc/php/conf.d/docker-php-ext-ioncube.ini
   ```

6. **Restart the PHP service.**
   ```bash
   cd /home/USERNAME/ && \
   docker --context=USERNAME compose down php-fpm-8.2 && \
   docker --context=USERNAME compose up -d php-fpm-8.2
   ```

7. **Verify the installation.**
   Check via CLI:
   ```bash
   docker --context=USERNAME exec php-fpm-8.2 php -v
   ```
   Or check via a phpinfo page: https://yourdomain.tld/info.php
   You should see: ionCube PHP Loader ... enabled

Once completed, the Ioncube Loader is installed persistently and fully compatible with OpenPanel’s Docker PHP stack.



