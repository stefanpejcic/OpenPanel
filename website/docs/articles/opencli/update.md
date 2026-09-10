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
