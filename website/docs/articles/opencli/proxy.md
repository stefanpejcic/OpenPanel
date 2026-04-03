# Proxy 

`opencli proxy` command is used to change `/openpanel` suffix to something else. 

View current setting:
```bash
opencli proxy
```

Set `/something` on every domain to redirect to OpenPanel interface:
```bash
opencli domain set something
```

Switch back to the default `/openpanel`:
```bash
opencli domain set default
```


You can also pass `--no-restart` flag to avoid interrupting the Caddy or OpenPanel services.
