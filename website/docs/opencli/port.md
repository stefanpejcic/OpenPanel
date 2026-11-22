# Port 

`opencli port` command is used to change port (`2083`) for accessing OpenPanel.

View current port:
```bash
opencli port
```

<details>
  <summary>Example output</summary>

```bash
# opencli port
2083
```
</details>

Set custom port:

```bash
opencli port set <NUMBER>
```
Example:
```bash
opencli port set 5000
```

<details>
  <summary>Example output</summary>

```bash
# opencli port set 5000
5000 is now set for accessing the OpenPanel interface.
```
</details>

Set default port `2083`:

```bash
opencli port default
```

<details>
  <summary>Example output</summary>

```bash
# opencli port default
2083 is now set for accessing the OpenPanel interface.
```
</details>

Caddy and OpenPanel are restarted and the new port is opened in the CSF firewall. Add `--no-restart` to skip both:
```bash
opencli port set <NUMBER> --no-restart
```
