# Port 

`opencli port` command is used to change port (`2083`) for accessing OpenPanel.

View current port:
```bash
opencli port
```

Set custom port:

```bash
opencli port set <NUMBER>
```
Example:
```bash
opencli port set 5000
```

Set default port `2083`:

```bash
opencli port default
```

Caddy and OpenPanel are restarted and the new port is opened in the CSF firewall. Add `--no-restart` to skip both:
```bash
opencli port set <NUMBER> --no-restart
```
