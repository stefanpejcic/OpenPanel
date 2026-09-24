# Update 

`opencli update` command is used to check if new update is available and to update.

Check if update is available:
```bash
opencli update --check
```

Start update immediately if `autoupdate` or `autopatch` is enabled:
```bash
opencli update
```

Start update immediately, regardless of `autoupdate` or `autopatch` setting:
```bash
opencli update --force
```

## Admin

Update only OpenAdmin UI *(`/usr/local/admin/` from [stefanpejcic/openadmin](https://github.com/stefanpejcic/openadmin))*:
```bash
opencli update --admin
```

## Panel

Update only OpenPanel UI *(docker image from [openpanel/openpanel-ui](https://hub.docker.com/r/openpanel/openpanel))*:
```bash
opencli update --panel
```

## Terminal

Update only OpenCLI *(`/usr/local/opencli/` from [stefanpejcic/opencli](https://github.com/stefanpejcic/opencli))*:
```bash
opencli update --cli
```

## Locales

Update only translation files *(locale files from [stefanpejcic/openpanel-translations](https://github.com/stefanpejcic/openpanel-translations))*:
```bash
opencli update --translations
```

## System

Update system packages and kernel, remove older kernels and check if a reboot is required:
```bash
opencli update --system
```
You'll be asked to confirm before packages are updated. A full server backup is recommended beforehand.

## Modules

Update only the OpenAdmin modules/features list *(`/etc/openpanel/openadmin/config/features.json` from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration))*:
```bash
opencli update --modules
```

## Templates

Update only the `docker-compose.yml` template used for new users *(`/etc/openpanel/docker/compose/1.0/docker-compose.yml`)*:
```bash
opencli update --compose
```

Update only the `.env` template used for new users *(`/etc/openpanel/docker/compose/1.0/.env`)*:
```bash
opencli update --env
```

The previous file is saved with a `.bak` extension. Existing users are not changed.

## PHP

Update PHP files and add new PHP versions to the template and to all existing users:
```bash
opencli update --php
```

This will:
- replace `/etc/openpanel/php/` (ini files, ionCube loaders, composer) with the latest files from [stefanpejcic/openpanel-configuration](https://github.com/stefanpejcic/openpanel-configuration),
- add any missing `php-fpm-X.Y` services to `docker-compose.yml` and their `PHP_FPM_X_Y_*` limits to `.env`,
- add missing PHP mounts (php.ini, ionCube) to the OpenLiteSpeed service,
- add missing ionCube mounts to existing `php-fpm-X.Y` services (e.g. when a loader for a PHP version becomes available),
- copy the `php.ini/X.Y.ini` file for each new version.

The template and every user in `/home/` are checked. Before changing a file, a backup is saved as `docker-compose.yml.php_<date>.bak` and `.env.php_<date>.bak`. If the new configuration is invalid, the backup is restored and that user is skipped. A summary is shown at the end.

Running containers are not restarted. New mounts take effect the next time the user's PHP or OpenLiteSpeed service is restarted.

## Combining options

Several options can be combined. They run in the order given:
```bash
opencli update --compose --env
```
