# Proxy 

`opencli proxy` command is used to change `/openpanel` suffix to something else. 

View current setting:
```bash
opencli proxy
```

<details>
  <summary>Example output</summary>

```bash
# opencli proxy
openpanel
```
</details>

Set `/something` on every domain to redirect to OpenPanel interface:
```bash
opencli proxy set /something
```

<details>
  <summary>Example output</summary>

```bash
# opencli proxy set /something
/something is now set for accessing the OpenPanel interface.
```
</details>

Switch back to the default `/openpanel`:
```bash
opencli proxy default
```

<details>
  <summary>Example output</summary>

```bash
# opencli proxy default
/openpanel is now set for accessing the OpenPanel interface.
```
</details>

The OpenPanel service is restarted in the background to apply the change. Add `--no-restart` to skip the restart.
