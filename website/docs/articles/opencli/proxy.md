# Proxy 

`opencli proxy` command is used to change `/openpanel` suffix to something else. 

View current setting:
```bash
opencli proxy
```

Set `/something` on every domain to redirect to OpenPanel interface:
```bash
opencli proxy set /something
```

Switch back to the default `/openpanel`:
```bash
opencli proxy default
```

The OpenPanel service is restarted in the background to apply the change. Add `--no-restart` to skip the restart.
