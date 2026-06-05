# Domains

## Varnish

Check Varnish status for domain, enable/disable Varnish caching.

```bash
opencli domains-varnish <domain> <on|off> <--short>
```

### Status

Display Varnish Cache status for domain:
 
```bash
opencli domains-varnish <domain>
```

### Enable

Enable Varnish Cache for a domain :
 
```bash
opencli domains-varnish <domain> on [--short]
```

### Disable

Disable Varnish Cache for a domain:
 
```bash
opencli domains-varnish <domain> off [--short]
```

## Whoowns

Check which username owns a certain domain name.

To check owner for a domain:

```bash
opencli domains-whoowns <DOMAIN-NAME>
```

To check owner and docker context for a domain:

```bash
opencli domains-whoowns <DOMAIN-NAME> --context
```

To check owner and docroot for a domain:

```bash
opencli domains-whoowns <DOMAIN-NAME> --docroot
```

