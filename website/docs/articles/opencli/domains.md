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

## SSL

Check SSL for domain, add custom certificate, view files.

```bash
opencli domains-ssl <DOMAIN_NAME> [status|info|logs|auto|custom] [path/to/fullchain.pem path/to/key.pem]
```

### Examples

Display command examples for a specific domain:
```bash
opencli domains-ssl <DOMAIN_NAME>
```

### Status

Display current SSL status for a domain:
```bash
opencli domains-ssl <DOMAIN_NAME> status
```

### Info

View ceritificate files a domain:
```bash
opencli domains-ssl <DOMAIN_NAME> info
```

### Logs

View caddy SSL-related logs for a domain:
```bash
opencli domains-ssl <DOMAIN_NAME> logs
```

### Custom

Setup a custom SSL for a domain:
```bash
opencli domains-ssl <DOMAIN_NAME> custom <CERT_PATH> <KEY_PATH>
```

### Auto

Switch to autoSSL for a domain (default):
```bash
opencli domains-ssl <DOMAIN_NAME> auto
```

## Docroot

View and change docroot for a domain.

```bash
opencli domains-docroot <DOMAIN_NAME> [update </var/www/html/>] --debug
```

### View

To view the current docroot for a domain:
```bash
opencli domains-docroot <DOMAIN_NAME>
```

### Update

To update a docroot for a domain:
```bash
opencli domains-docroot <DOMAIN_NAME> update <docroot> [--debug]
```

## DNSSEC

Enable [DNSSEC](https://en.wikipedia.org/wiki/Domain_Name_System_Security_Extensions) for a domain and re-sign after changes the zone.

```bash
opencli domains-dnssec <DOMAIN_NAME> [--update | --check]
```

### Check

Check if domain has DNSSEC enabled:
```bash
opencli domains-dnssec <DOMAIN_NAME> --check
```

### Update

Configure/update DNSSEC for a domain:
```bash
opencli domains-dnssec <DOMAIN_NAME> --update
```


## HSTS

Manage [HSTS](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) for a domain.

```bash
opencli domains-hsts <DOMAIN_NAME> [on|off]
```

### Status
Check HSTS status for a domain:
```bash
opencli domains-hsts <DOMAIN_NAME> [on|off]
```

### Enable
Enable HSTS for a domain:
```bash
opencli domains-hsts <DOMAIN_NAME> on
```

### Disable
Disable HSTS for a domain:
```bash
opencli domains-hsts <DOMAIN_NAME> off
```


## Edit

Edit VirtualHosts file for a domain or enter its docroot.

```
opencli domains-edit <DOMAIN_NAME> [--ws]
```

### Docroot

`cd` into the domain docroot:
```
opencli domains-edit <DOMAIN_NAME>
```

### VHost

Open the VirtualHost file for a domain using `nano` edit, and upon saving restart user webserver:
```
opencli domains-edit <DOMAIN_NAME> --ws
```

## Stats

Parse caddy access logs for users domains and generate static html.

```bash
opencli domains-stats <USERNAME> --debug
```

Generate stats for all domains:
```bash
opencli domains-stats [--debug]
```

Generate stats for domains owned by a specific user:
```bash
opencli domains-stats <USERNAME> [--debug]
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

