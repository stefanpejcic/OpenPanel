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

Update only OpenAdmin UI *(`/usr/local/admin/` from [stefanpejcic/openadmin](https://github.com/stefanpejcic/openadmin))*:
```bash
opencli update --admin
```

Update only OpenPanel UI *(docker image from [openpanel/openpanel-ui](https://hub.docker.com/r/openpanel/openpanel-ui))*:
```bash
opencli update --panel
```

Update only OpenCLI *(`/usr/local/opencli/` from [stefanpejcic/opencli](https://github.com/stefanpejcic/opencli))*:
```bash
opencli update --cli
```
