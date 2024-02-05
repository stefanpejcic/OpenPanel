---
sidebar_position: 8
---

# PHP

Manage users' PHP versions: list enabled, list available, change version, etc.

## Get version for a domain

To view the current PHP version used by a domain, run the following command:

```bash
opencli php-domain_php <DOMAIN-NAME>
```

Example:
```bash
# opencli php-domain_php pejcic.rs
Domain 'pejcic.rs' (owned by user: stefan) uses PHP version: php8.1
```

## Change version for a domain

To change a PHP version for a domain name run the `domain_php` script with `--update` flag::

```bash
opencli php-domain_php <DOMAIN-NAME> --update <PHP-VERSION>
```

Example:
```bash
# opencli php-domain_php pejcic.rs --update 8.3
Updating PHP version to: 8.3
Domain 'pejcic.rs' (owned by user: stefan) uses PHP version: php8.3
Updating PHP version in the Apache configuration file...
 * Reloading Apache httpd web server apache2
 *
Updated PHP version in the configuration file to 8.3
```

## List the default version

The default PHP version for a user determines which PHP version will be used for all domains that the user adds in the future. It does not change the PHP version for any existing domains.

To list the currently set default PHP version for a user, run the following command:

```bash
opencli php-default_php_version <USERNAME>
```

Example:
```bash
# opencli php-default_php_version stefan
Default PHP version for user 'stefan' is: php8.3
```

## Change the default version

To update the default PHP version for a user use the php-default_php_version with `--update` flag and provide the new PHP version.

```bash
opencli php-default_php_version <USERNAME> --update <VERSION>
```

Example:
```bash
# opencli php-default_php_version stefan --update 8.1
PHP version for user 'stefan' updated to: 8.1
```

## List installed versions

To list all installed PHP versions for a user, run the following command:

```bash
opencli php-enabled_php_versions <USERNAME>
```

Example:
```bash
# opencli php-enabled_php_versions stefan
php7.4
php8.1
php8.2
```

## List available versions

To get available (that can be installed) PHP versions for users' OS, run the following command:

```bash
opencli php-get_available_php_versions <USERNAME>
```

Example:
```bash
# opencli php-get_available_php_versions stefan
....
PHP versions for user stefan have been updated and stored in /home/stefan/etc/.panel/php/php_available_versions.json.
```

The script will by default update users' available PHP versions setting for the UI, optionally you can add `--show` flag to display the available versions.

```bash
opencli php-get_available_php_versions <USERNAME> --show
```

Example:
```bash
# opencli php-get_available_php_versions stefan --show
....
Available PHP versions for user stefan:
php8.1-fpm
php5.6-fpm
php7.0-fpm
php7.1-fpm
php7.2-fpm
php7.3-fpm
php7.4-fpm
php8.0-fpm
php8.2-fpm
php8.3-fpm
```

The `get_available_php_versions` script performs various actions:

- Runs `apt-get update` inside users contianer
- Lists available PHP versions from remote repositories
- Saves the list to `/php_available_versions.json` in user home directory
- optionally display the list

## Install a new version

To install a a new PHP version run the following command:

```bash
opencli php-install_php_version <USERNAME> <PHP-VERSION>
```

Example:
```bash
# opencli php-install_php_version stefan 8.2
Hit:1 https://ppa.launchpadcontent.net/ondrej/php/ubuntu jammy InRelease
Hit:2 http://security.ubuntu.com/ubuntu jammy-security InRelease
Hit:3 http://archive.ubuntu.com/ubuntu jammy InRelease
Hit:4 http://archive.ubuntu.com/ubuntu jammy-updates InRelease
Hit:5 http://archive.ubuntu.com/ubuntu jammy-backports InRelease
Reading package lists... Done
Reading package lists... Done
Building dependency tree... Done
Reading state information... Done
The following additional packages will be installed:
  php8.1-cli php8.1-common php8.1-opcache php8.1-readline
Suggested packages:
  php-pear
The following NEW packages will be installed:
  php8.1-cli php8.1-common php8.1-fpm php8.1-opcache php8.1-readline
0 upgraded, 5 newly installed, 0 to remove and 5 not upgraded.
Need to get 4804 kB of archives.
After this operation, 21.1 MB of additional disk space will be used.
..
.
```

