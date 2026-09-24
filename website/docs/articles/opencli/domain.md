# Domain 

`opencli domain` command is used to set domain name for accessing the OpenPanel and OpenAdmin interfaces.

View current setting:
```bash
opencli domain
```

Set domain:

```bash
opencli domain set <DOMAIN_NAME>
```
Example:
```bash
opencli domain set srv7.openpanel.org
```

Additional flags:

- `--debug` - display verbose information as the script is running.
- `--no-restart` - skip restarting services after the change.

Note: If domain is not pointed to the server via A record and SSL can not be generated, then `http://IP` will be used instead of `https://DOMAIN`.

Set IP address to be used instead of domain:

```bash
opencli domain ip
```
