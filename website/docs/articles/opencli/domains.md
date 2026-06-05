# Domains


## User

Lists all domain names currently owned by a specific user.

```bash
opencli domains-user <USERNAME> [--docroot|--php_version]
```

## Suspend

Suspend a domain name:

```bash
opencli domains-suspend <DOMAIN_NAME>
```

## Unsuspend

Unsuspend a domain name:

```bash
opencli domains-unsuspend <DOMAIN_NAME>
```

## Update NS

Change nameservers for a single or all dns zones.

```bash
opencli domains-update_ns <DOMAIN_NAME>|--all
```

Update the zone file for a specific domain:
```bash
opencli domains-update_ns <DOMAIN_NAME>
```

Update all zone files:
```bash
opencli domains-update_ns --all [-y]
```



## Varnish

Check Varnish status for domain, enable/disable Varnish caching.

```bash
opencli domains-varnish <DOMAIN_NAME> <on|off> <--short>
```

### Status

Display Varnish Cache status for domain:
 
```bash
opencli domains-varnish <DOMAIN_NAME>
```

### Enable

Enable Varnish Cache for a domain :
 
```bash
opencli domains-varnish <DOMAIN_NAME> on [--short]
```

### Disable

Disable Varnish Cache for a domain:
 
```bash
opencli domains-varnish <DOMAIN_NAME> off [--short]
```

## Whoowns

Check which username owns a certain domain name.

To check owner for a domain:

```bash
opencli domains-whoowns <DOMAIN_NAME>
```

To check owner and docker context for a domain:

```bash
opencli domains-whoowns <DOMAIN_NAME> --context
```

To check owner and docroot for a domain:

```bash
opencli domains-whoowns <DOMAIN_NAME> --docroot
```

