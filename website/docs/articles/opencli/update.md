# Update 

`opencli update` command is used to check if new update is available and to update.

Check if update is available:
```bash
opencli update --check
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --check
{"status": "Update available", "installed_version": "2.0.11", "latest_version": "2.0.12"}
```
</details>

Start update immediately if `autoupdate` or `autopatch` is enabled:
```bash
opencli update
```

<details>
  <summary>Example output</summary>

```bash
# opencli update
[INFO] Update available and will be automatically installed
===== Starting update to version 2.0.12 =====
[INFO] Update log: /var/log/openpanel/updates/2.0.12.log
---- Updating OpenCLI ----
[INFO] Downloading terminal scripts from github: https://github.com/stefanpejcic/opencli/archive/refs/heads/podman.tar.gz
[INFO] [✔] Terminal commands updated successfully
---- Updating OpenPanel container image ----
...
---- [✔] image openpanel/openpanel:2.0.12 downloaded successfully ----
---- Updating version in /root/.env ----
---- Restarting OpenPanel service ----
...
---- Cleaning up previous images ----
---- Updating installed locales ----
---- [✔] Locales updated: 2, failed: 0 ----
---- Updating OpenAdmin modules ----
---- Updating OpenAdmin ----
---- [✔] OpenAdmin update triggered (applying in background) ----
---- [✔] Minor update - skipping system updates ----
---- Checking for custom post-update scripts ----
---- Update completed successfully! ----
```
</details>

If both `autoupdate` and `autopatch` are disabled, nothing is installed:

<details>
  <summary>Example output</summary>

```bash
# opencli update
[INFO] [!] Autopatch and Autoupdate are both disabled. No updates will be installed automatically.
```
</details>

Start update immediately, regardless of `autoupdate` or `autopatch` setting:
```bash
opencli update --force
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --force
[INFO] [!] Forcing update, ignoring autopatch and autoupdate settings
[INFO] Update available and will be automatically installed
===== Starting update to version 2.0.12 =====
[INFO] Update log: /var/log/openpanel/updates/2.0.12.log
---- Updating OpenCLI ----
[INFO] Downloading terminal scripts from github: https://github.com/stefanpejcic/opencli/archive/refs/heads/podman.tar.gz
[INFO] [✔] Terminal commands updated successfully
---- Updating OpenPanel container image ----
...
---- [✔] image openpanel/openpanel:2.0.12 downloaded successfully ----
---- Updating version in /root/.env ----
---- Restarting OpenPanel service ----
...
---- Cleaning up previous images ----
---- Updating installed locales ----
---- [✔] Locales updated: 2, failed: 0 ----
---- Updating OpenAdmin modules ----
---- Updating OpenAdmin ----
---- [✔] OpenAdmin update triggered (applying in background) ----
---- [✔] Minor update - skipping system updates ----
---- Checking for custom post-update scripts ----
---- Update completed successfully! ----
```
</details>

## Admin

Update only OpenAdmin UI *(`/usr/local/admin/` from [stefanpejcic/openadmin](https://github.com/stefanpejcic/openadmin))*:
```bash
opencli update --admin
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --admin
Updating OpenAdmin
[✔] OpenAdmin update triggered (applying in background)
```
</details>

## Panel

Update only OpenPanel UI *(container image from [openpanel/openpanel-ui](https://hub.docker.com/r/openpanel/openpanel))*:
```bash
opencli update --panel
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --panel
...
openpanel
```
</details>

## Terminal

Update only OpenCLI *(`/usr/local/opencli/` from [stefanpejcic/opencli](https://github.com/stefanpejcic/opencli))*:
```bash
opencli update --cli
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --cli
Updating OpenCLI
[INFO] Downloading terminal scripts from github: https://github.com/stefanpejcic/opencli/archive/refs/heads/podman.tar.gz
[INFO] [✔] Terminal commands updated successfully
```
</details>

## Locales

Update only translation files *(locale files from [stefanpejcic/openpanel-translations](https://github.com/stefanpejcic/openpanel-translations))*:
```bash
opencli update --translations
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --translations
Updating installed locales
Compiling updated .mo files
[✔] Locales updated: 2, failed: 0
[INFO] Restarting OpenPanel service
[INFO] [✔] Translations updated and OpenPanel restarted
```
</details>

## System

Update system packages and kernel, remove older kernels and check if a reboot is required:
```bash
opencli update --system
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --system
[INFO] Updating system packages
System package update ready to proceed. A full server backup is recommended beforehand. Continue? [y/N] y
[INFO] Installing required system tools
[INFO] Updating system packages
[INFO] Cleaning up packages
[INFO] Removing old kernels (Debian/Ubuntu)
[INFO] Checking if reboot is required
[WARN] [!]  Reboot required!
```
</details>

You'll be asked to confirm before packages are updated. A full server backup is recommended beforehand.

## Modules

Update only the OpenAdmin modules/features list *(`/etc/openpanel/openadmin/config/features.json` from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration))*:
```bash
opencli update --modules
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --modules
Updating OpenAdmin modules
```
</details>

## Templates

Update only the `docker-compose.yml` template used for new users *(`/etc/openpanel/docker/compose/1.0/docker-compose.yml`)*:
```bash
opencli update --compose
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --compose
[INFO] Updating /etc/openpanel/docker/compose/1.0/docker-compose.yml
[INFO] Previous file saved as /etc/openpanel/docker/compose/1.0/docker-compose.yml.bak
[INFO] [✔] docker-compose.yml updated
```
</details>

Update only the `.env` template used for new users *(`/etc/openpanel/docker/compose/1.0/.env`)*:
```bash
opencli update --env
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --env
[INFO] Updating /etc/openpanel/docker/compose/1.0/.env
[INFO] Previous file saved as /etc/openpanel/docker/compose/1.0/.env.bak
[INFO] [✔] .env updated
```
</details>

The previous file is saved with a `.bak` extension. Existing users are not changed.

## PHP

Update PHP files and add new PHP versions to the template and to all existing users:
```bash
opencli update --php
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --php
===== Updating PHP =====
[INFO] Downloading configuration files from github
[INFO] PHP versions on github: 5.6 7.0 7.1 7.2 7.3 7.4 8.0 8.1 8.2 8.3 8.4 8.5
[INFO] Overwriting /etc/openpanel/php/
[INFO] [✔] /etc/openpanel/php/ updated
[INFO] Checking template in /etc/openpanel/docker/compose/1.0
[INFO] [template] up to date
[INFO] Checking user stefan (1)
[INFO] [stefan] mounts added: php-fpm-8.5:/etc/openpanel/php/ioncube/ioncube_loader_lin_8.5.so php-fpm-8.5:/etc/openpanel/php/ioncube_extension.ini
[INFO] [stefan] backups: /home/stefan/docker-compose.yml.php_20260924_103000.bak, /home/stefan/.env.php_20260924_103000.bak
[INFO] Checking user demo (2)
[INFO] [demo] added: php-fpm-8.5 PHP_FPM_8_5_CPU PHP_FPM_8_5_RAM PHP_FPM_8_5_PIDS php.ini/8.5.ini
[INFO] [demo] backups: /home/demo/docker-compose.yml.php_20260924_103000.bak, /home/demo/.env.php_20260924_103000.bak
===== PHP update summary =====
  template: unchanged
  stefan: changed: | mounts: php-fpm-8.5:/etc/openpanel/php/ioncube/ioncube_loader_lin_8.5.so php-fpm-8.5:/etc/openpanel/php/ioncube_extension.ini
  demo: changed: php-fpm-8.5 PHP_FPM_8_5_CPU PHP_FPM_8_5_RAM PHP_FPM_8_5_PIDS php.ini/8.5.ini
  Users checked: 2 | updated: 2 | up to date: 0 | skipped: 0 | failed: 0
```
</details>

This will:
- replace `/etc/openpanel/php/` (ini files, ionCube loaders, composer) with the latest files from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration),
- add any missing `php-fpm-X.Y` services to `docker-compose.yml` and their `PHP_FPM_X_Y_*` limits to `.env`,
- add missing PHP mounts (php.ini, ionCube) to the OpenLiteSpeed service,
- add missing ionCube mounts to existing `php-fpm-X.Y` services (e.g. when a loader for a PHP version becomes available),
- copy the `php.ini/X.Y.ini` file for each new version.

The template and every user in `/home/` are checked. Before changing a file, a backup is saved as `docker-compose.yml.php_<date>.bak` and `.env.php_<date>.bak`. If the new configuration is invalid, the backup is restored and that user is skipped. A summary is shown at the end.

Running containers are not restarted. New mounts take effect the next time the user's PHP or OpenLiteSpeed service is restarted.

## Service files

These options download the latest files for a single service from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration) and overwrite the files in `/etc/openpanel/`. Files that exist only on your server are kept. Several options can be combined, the files are downloaded only once.

### WP-CLI

Update WP-CLI *(`/etc/openpanel/wordpress/wp-cli.phar`)*:
```bash
opencli update --wp
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --wp
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/wordpress/wp-cli.phar updated
```
</details>

### OpenLiteSpeed

Update OpenLiteSpeed files *(`/etc/openpanel/openlitespeed/` from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration/tree/main/openlitespeed))*:
```bash
opencli update --ols
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --ols
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/openlitespeed updated
```
</details>

### Apache

Update Apache files *(`/etc/openpanel/apache/` from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration/tree/main/apache))*:
```bash
opencli update --apache
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --apache
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/apache updated
```
</details>

### ClamAV

Update ClamAV files *(`/etc/openpanel/clamav/`)*. If the `clamav` container is running, it is also removed, its image is deleted, and the container is started again with a newly downloaded image:
```bash
opencli update --clamav
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --clamav
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/clamav updated
[INFO] Removing clamav container
[INFO] Deleting image docker.io/clamav/clamav:1.4
[INFO] Downloading new clamav image
[INFO] Starting clamav container
[INFO] [✔] clamav container restarted with the new image
```
</details>

If the container is not running, only the files are updated:

<details>
  <summary>Example output</summary>

```bash
# opencli update --clamav
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/clamav updated
[INFO] clamav container is not running, skipping image update
```
</details>

### phpMyAdmin

Update phpMyAdmin files *(`/etc/openpanel/mysql/phpmyadmin/`)*. If the `phpmyadmin` container is running, it is also removed, its image is deleted, and the container is started again with a newly downloaded image:
```bash
opencli update --phpmyadmin
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --phpmyadmin
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/mysql/phpmyadmin updated
[INFO] Removing phpmyadmin container
[INFO] Deleting image docker.io/library/phpmyadmin:latest
[INFO] Downloading new phpmyadmin image
[INFO] Starting phpmyadmin container
[INFO] [✔] phpmyadmin container restarted with the new image
```
</details>

### PostgreSQL

Update PostgreSQL files *(`/etc/openpanel/postgres/` from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration/tree/main/postgres))*:
```bash
opencli update --postgres
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --postgres
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/postgres updated
```
</details>

### Skeleton

Update the files copied to new user accounts *(`/etc/openpanel/skeleton/` from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration/tree/main/skeleton))*:
```bash
opencli update --skeleton
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --skeleton
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/skeleton updated
```
</details>

### SSH

Update SSH files *(`/etc/openpanel/ssh/` from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration/tree/main/ssh))*:
```bash
opencli update --ssh
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --ssh
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/ssh updated
```
</details>

### Varnish

Update Varnish files *(`/etc/openpanel/varnish/` from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration/tree/main/varnish))*:
```bash
opencli update --varnish
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --varnish
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/varnish updated
```
</details>

### Cron

Update the OpenPanel cron jobs *(`/etc/openpanel/cron`)* and install them to `/etc/cron.d/openpanel`:
```bash
opencli update --cron
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --cron
[INFO] Downloading configuration files from github
[INFO] [✔] /etc/openpanel/cron updated
[INFO] [✔] /etc/cron.d/openpanel updated
```
</details>

## Combining options

Several options can be combined. They run in the order given:
```bash
opencli update --compose --env
```

<details>
  <summary>Example output</summary>

```bash
# opencli update --compose --env
[INFO] Updating /etc/openpanel/docker/compose/1.0/docker-compose.yml
[INFO] Previous file saved as /etc/openpanel/docker/compose/1.0/docker-compose.yml.bak
[INFO] [✔] docker-compose.yml updated
[INFO] Updating /etc/openpanel/docker/compose/1.0/.env
[INFO] Previous file saved as /etc/openpanel/docker/compose/1.0/.env.bak
[INFO] [✔] .env updated
```
</details>

