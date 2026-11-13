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

When the mailserver is installed, it is switched to the domain's SSL certificate (IMAP, SMTP and Roundcube connect to `ssl://DOMAIN`). The command waits up to 60 seconds for Caddy to issue the certificate. If it's still missing after that, it prints the message below and the mailserver is switched in the background as soon as the certificate is issued (checked every 10 seconds for 10 minutes):

```bash
SSL certificate for srv7.openpanel.org is not issued yet, mailserver will switch to it once Caddy gets one.
```

If `/webmail` points to the default webmail address (`http://IP:8080` or `https://webmail.<old domain>`), it is changed to `https://webmail.<DOMAIN>`, and back to `http://IP:8080` with `opencli domain ip`. A webmail domain set with [`opencli email-webmail domain`](/docs/articles/opencli/email/) is left unchanged.

The new domain is also written to the hosts file that users' containers start with, see [`opencli docker-hosts`](/docs/articles/opencli/docker/#hosts).

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

