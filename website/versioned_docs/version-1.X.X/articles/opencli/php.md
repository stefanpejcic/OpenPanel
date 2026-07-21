# PHP

Manage users' PHP versions: list enabled, list available, change version, etc.

### Get version for a domain

To view the current PHP version used by a domain, run the following command:

```bash
opencli php-domain <DOMAIN-NAME>
```

Example:
```bash
# opencli php-domain pejcic.rs
Domain 'pejcic.rs' (owned by user: stefan) uses PHP version: php8.1
```

### Change version for a domain

To change a PHP version for a domain name run the `domain` script with `--update` flag::

```bash
opencli php-domain <DOMAIN-NAME> --update <PHP-VERSION>
```

Example:
```bash
# opencli php-domain pejcic.rs --update 8.3
Updating PHP version to: 8.3
Domain 'pejcic.rs' (owned by user: stefan) uses PHP version: php8.3
Updating PHP version in the Apache configuration file...
 * Reloading Apache httpd web server apache2
 *
Updated PHP version in the configuration file to 8.3
```

### View default version

The default PHP version for a user determines which PHP version will be used for all domains that the user adds in the future. It does not change the PHP version for any existing domains.

To list the currently set default PHP version for a user, run the following command:

```bash
opencli php-default <USERNAME>
```

Example:
```bash
# opencli php-default stefan
Default PHP version for user 'stefan' is: php8.3
```

### Change the default version

To update the default PHP version for a user use the php-default with `--update` flag and provide the new PHP version.

```bash
opencli php-default <USERNAME> --update <VERSION>
```

Example:
```bash
# opencli php-default stefan --update 8.1
PHP version for user 'stefan' updated to: 8.1
```

### List versions

To list all available PHP versions for a user, grep 'php-fpm' in docker-compose.yml file for the user:

```bash
grep php-fpm- /home/ <USERNAME>/docker-compose.yml
```


Example:
```bash
# grep php-fpm- /home/demo/docker-compose.yml
  php-fpm-5.6:
    container_name: php-fpm-5.6
  php-fpm-7.0:
    container_name: php-fpm-7.0
  php-fpm-7.1:
    container_name: php-fpm-7.1
  php-fpm-7.2:
    container_name: php-fpm-7.2
  php-fpm-7.3:
    container_name: php-fpm-7.3
  php-fpm-7.4:
    container_name: php-fpm-7.4
  php-fpm-8.0:
    container_name: php-fpm-8.0
  php-fpm-8.1:
    container_name: php-fpm-8.1
  php-fpm-8.2:
    container_name: php-fpm-8.2
  php-fpm-8.3:
    container_name: php-fpm-8.3
  php-fpm-8.4:
    container_name: php-fpm-8.4
```
