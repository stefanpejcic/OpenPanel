# Domain 

`opencli domain` command is used to set domain name for accessing the OpenPanel and OpenAdmin interfaces.

View current setting:
```bash
opencli domain
```

<details>
  <summary>Example output</summary>

```bash
# opencli domain
srv7.openpanel.org
```
</details>

Set domain:

```bash
opencli domain set <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domain set srv7.openpanel.org
srv7.openpanel.org is now set for accessing the OpenPanel and OpenAdmin interfaces.
```
</details>

Additional flags:

- `--debug` - display verbose information as the script is running.
- `--no-restart` - skip restarting services after the change.

Note: If domain is not pointed to the server via A record and SSL can not be generated, then `http://IP` will be used instead of `https://DOMAIN`.

Set IP address to be used instead of domain:

```bash
opencli domain ip
```

<details>
  <summary>Example output</summary>

```bash
# opencli domain ip
203.0.113.10 is now set for accessing the OpenPanel and OpenAdmin interfaces.
```
</details>

