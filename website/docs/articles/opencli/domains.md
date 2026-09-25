# Domains


## All

Lists all domain names currently hosted on the server:

```bash
opencli domains-all [--docroot] [--php_version] [--json]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-all
example.com
example.net
blog.example.com

# opencli domains-all --docroot --php_version
example.com	/var/www/html/example.com	8.3
example.net	/var/www/html/example.net	8.1
blog.example.com	/var/www/html/blog.example.com	8.3

# opencli domains-all --json
{"data":{"example.com":{"document_root":"/var/www/html/example.com","php":{"php_version_id":"php83","version":"8.3","ini_path":"/home/stefan/php.ini/8.3.ini","is_native":true,"handler":"php-fpm"},"owner":"stefan"},...}}
```
</details>

- `--docroot` - also show the document root for each domain.
- `--php_version` - also show the PHP version for each domain.
- `--json` - display output as JSON.

## User

Lists all domain names currently owned by a specific user.

```bash
opencli domains-user <USERNAME> [--docroot|--php_version]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-user stefan
example.com
blog.example.com

# opencli domains-user stefan --docroot
example.com	/var/www/html/example.com
blog.example.com	/var/www/html/blog.example.com
```
</details>

## Add

Add a domain name for a user. Add `--debug` to see every step (checks, DNS zone, vhost, Caddy):

```bash
opencli domains-add <DOMAIN_NAME> <USERNAME> [--docroot DOCUMENT_ROOT] [--php_version N.N] [--skip_caddy] [--skip_vhost] [--skip_containers] [--skip_dns] [--skip-sentinel] [--hs_ed25519_public_key KEY --hs_ed25519_secret_key KEY] [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-add example.com stefan
Domain example.com added successfully
```
</details>

### Custom Docroot

To add a domain name with custom docroot:

```bash
opencli domains-add <DOMAIN_NAME> <USERNAME> --docroot DOCUMENT_ROOT
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-add example.com stefan --docroot /var/www/html/example.com/public
Domain example.com added successfully
```
</details>

### Custom PHP version

To add a domain name with custom PHP version:

```bash
opencli domains-add <DOMAIN_NAME> <USERNAME> --php_version N.N
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-add example.com stefan --php_version 8.3
Domain example.com added successfully
```
</details>


### Skip Flags

> **Warning:** These flags are intended primarily for automation and debugging purposes and should generally **not be used manually in the terminal**.
> Most skip flags are used by automation scripts when adding multiple domains at once, allowing services to be started only after all domains have been provisioned. The `--skip_dns`, `--skip_vhost`, and `--skip_caddy` options are also useful for troubleshooting and retry operations.

* **Skip DNS** – Add a domain without creating a DNS zone:

  ```bash
  opencli domains-add <DOMAIN_NAME> <USERNAME> --skip_dns
  ```

* **Skip Vhosts** – Add a domain without creating virtual host configurations or starting the user's PHP and web server containers:

  ```bash
  opencli domains-add <DOMAIN_NAME> <USERNAME> --skip_vhost
  ```

* **Skip Containers** – Add a domain without starting the user's PHP and web server containers:

  ```bash
  opencli domains-add <DOMAIN_NAME> <USERNAME> --skip_containers
  ```

* **Skip Caddy** – Add a domain without creating a Caddy virtual host configuration:

  ```bash
  opencli domains-add <DOMAIN_NAME> <USERNAME> --skip_caddy
  ```

* **Skip Sentinel** – Add a domain without sending the *domains_add* notification:

  ```bash
  opencli domains-add <DOMAIN_NAME> <USERNAME> --skip-sentinel
  ```

The output is the same with any of the skip flags:

<details>
  <summary>Example output</summary>

```bash
# opencli domains-add example.com stefan --skip_dns --skip_containers
Domain example.com added successfully
```
</details>

**Note:** These options are primarily intended for bulk provisioning workflows (like [cpanel account importer](/docs/articles/transfers/import-cpanel-backup-to-openpanel/) or [user transfer](/docs/articles/transfers/transfer-openpanel-account-to-another-server/)) and advanced troubleshooting scenarios.

### Onion (Tor) domains

When adding a `.onion` domain, you can pass an existing hidden service key pair instead of generating a new one:

```bash
opencli domains-add <ONION_DOMAIN> <USERNAME> --hs_ed25519_public_key <KEY> --hs_ed25519_secret_key <KEY>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-add abcdefghijklmnop.onion stefan --hs_ed25519_public_key <KEY> --hs_ed25519_secret_key <KEY>
Domain abcdefghijklmnop.onion added successfully
```
</details>

## Suspend

Suspend a domain name:

```bash
opencli domains-suspend <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-suspend example.com
Domain suspended successfully.
```
</details>

Add a reason for suspending the domain:

```bash
opencli domains-suspend <DOMAIN_NAME> --comment="<REASON>"
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-suspend example.com --comment="Unpaid invoice"
Domain suspended successfully.
```
</details>

## Unsuspend

Unsuspend a domain name:

```bash
opencli domains-unsuspend <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-unsuspend example.com
Domain unsuspended successfully.
```
</details>

## Delete

Delete a domain name:

```bash
opencli domains-delete <DOMAIN_NAME> [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-delete example.com
Domain example.com deleted successfully
```
</details>

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

<details>
  <summary>Example output</summary>

```bash
# opencli domains-ssl example.com
Usage examples for domain example.com:

- Check current SSL status for domain (AutoSSL, CustomSSL or No SSL):
  opencli domains-ssl example.com status
- Display fullchain and key files for the domain:
  opencli domains-ssl example.com info
- Set AutoSSL for the domain (default):
  opencli domains-ssl example.com auto
- Add custom certificate for the domain:
  opencli domains-ssl example.com custom /var/www/html/fullchain.pem /var/www/html/key.pem
- View SSL-related lines for the domain from Caddy logs:
  opencli domains-ssl example.com logs
```
</details>

### Status

Display current SSL status for a domain:
```bash
opencli domains-ssl <DOMAIN_NAME> status
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-ssl example.com status
AutoSSL
```
</details>

### Info

View ceritificate files a domain:
```bash
opencli domains-ssl <DOMAIN_NAME> info
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-ssl example.com info
-----BEGIN CERTIFICATE-----
MIIFBjCCA+6gAwIBAgISBL...
-----END CERTIFICATE-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIH...
-----END EC PRIVATE KEY-----
```
</details>

### Logs

View caddy SSL-related logs for a domain:
```bash
opencli domains-ssl <DOMAIN_NAME> logs
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-ssl example.com logs
Showing SSL-related log lines for example.com
-------------------------------------------------------
{"level":"info","ts":1727170200.12,"logger":"tls.obtain","msg":"obtaining certificate","identifier":"example.com"}
{"level":"info","ts":1727170204.87,"logger":"tls.obtain","msg":"certificate obtained successfully","identifier":"example.com"}
```
</details>

Show a specific number of lines, or follow the log live:
```bash
opencli domains-ssl <DOMAIN_NAME> logs 1000
opencli domains-ssl <DOMAIN_NAME> logs -f
```

### Custom

Setup a custom SSL for a domain:
```bash
opencli domains-ssl <DOMAIN_NAME> custom <CERT_PATH> <KEY_PATH>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-ssl example.com custom /var/www/html/fullchain.pem /var/www/html/key.pem
Adding custom certificate..
Updated example.com to use custom SSL.
```
</details>

### Auto

Switch to autoSSL for a domain (default):
```bash
opencli domains-ssl <DOMAIN_NAME> auto
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-ssl example.com auto
Updated example.com to use AutoSSL.
```
</details>

Email users about SSL certificates that need attention:
```bash
opencli domains-ssl --notify
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-ssl --notify
stefan: notified about 2 certificate(s).
```
</details>

It checks every domain that points to this server and emails its owner, if they have [SSL certificate problem](/docs/panel/account/notifications/#ssl-certificate-problem) turned on, when:
- an AutoSSL certificate has 7 days or less left, which means renewal is failing, since certificates are renewed about 30 days before they expire. The email includes the last error from the Caddy logs.
- any certificate, AutoSSL or custom, has 1 day or less left.

Each alert is sent once per certificate. It runs daily from cron.

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

<details>
  <summary>Example output</summary>

```bash
# opencli domains-docroot example.com
/var/www/html/example.com
```
</details>

### Update

To update a docroot for a domain:
```bash
opencli domains-docroot <DOMAIN_NAME> update <docroot> [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-docroot example.com update /var/www/html/example.com/public
Docroot updated to: /var/www/html/example.com/public for domain: example.com
```
</details>

## DNS

Manage DNS service and domain zones.


### Reconfig

To load new DNS zones into the Bind server:

```bash
opencli domains-dns reconfig
```
<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns reconfig
Loading new DNS zones..
```
</details>

`

### Check

To check and validate a DNS zone for a domain:

```bash
opencli domains-dns check <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns check example.com
Checking DNS zone for domain: example.com
zone example.com/IN: loaded serial 2026092401
OK
```
</details>

### Reload

To reload a DNS zone for a single domain:

```bash
opencli domains-dns reload <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns reload example.com
Reloading DNS zone for domain: example.com
zone reload queued
```
</details>

To reload all DNS zones, run the command without a domain:

```bash
opencli domains-dns reload
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns reload
Reloading all DNS zones..
server reload successful
```
</details>

### Show

To display the DNS zone for a single domain:

```bash
opencli domains-dns show <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns show example.com
DNS zone for domain: example.com - file: /etc/bind/zones/example.com.zone
$TTL 1h
@       IN      SOA     ns1.openpanel.com. admin.example.com. (
                        2026092401      ; Serial number
                        1h              ; Refresh interval
                        15m             ; Retry interval
                        1w              ; Expire interval
                        1h )            ; Minimum TTL

@ IN NS      ns1.openpanel.com.
@ IN NS      ns2.openpanel.com.
...
example.com.     14400     IN      A       203.0.113.10
example.com.     3600     IN      MX       0 example.com.
www     14400     IN      CNAME   example.com.
...
```
</details>

### List

To list all domains with DNS zones on the server:

```bash
opencli domains-dns list
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns list
Displaying all DNS zones on the server:
/etc/bind/zones/example.com.zone
/etc/bind/zones/example.net.zone
```
</details>

### Create

To create a DNS zone for a domain:

```bash
opencli domains-dns create <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns create example.com
Creating DNS zone for domain: example.com
No subdomains found for example.com
```
</details>

### Delete

To delete a DNS zone for a domain:

```bash
opencli domains-dns delete <DOMAIN_NAME> [-y]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns delete example.com -y
Deleting DNS zone for domain: example.com
```
</details>

Add `-y` to skip the confirmation prompt.

### Default

To restore the default DNS zone for a domain:

```bash
opencli domains-dns default <DOMAIN_NAME> [-y]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns default example.com -y
Deleting DNS zone for domain: example.com  and restoring default zone..
```
</details>

Add `-y` to skip the confirmation prompt.

### Count

To display the total number of DNS zones present on the server:

```bash
opencli domains-dns count
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns count
Total number of DNS zones on the server: 12
```
</details>

### Config

To check the main Bind configuration file for syntax errors:

```bash
opencli domains-dns config
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns config
Checking /etc/bind/named.conf configuration:
```
</details>

### Start

To start the DNS server:

```bash
opencli domains-dns start
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns start
Starting DNS service..
openpanel_dns
```
</details>

### Restart

To perform a soft restart of the Bind9 Docker container:

```bash
opencli domains-dns restart
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns restart
Restarting DNS service..
openpanel_dns
```
</details>

### Hard Restart

To perform a hard restart by terminating the container and starting it again:

```bash
opencli domains-dns hard-restart
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns hard-restart
Stopping DNS service..
openpanel_dns
openpanel_dns
Starting DNS service..
openpanel_dns
```
</details>

### Stop

To stop the DNS server:

```bash
opencli domains-dns stop
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dns stop
Stopping DNS service..
openpanel_dns
openpanel_dns
```
</details>


## DNSSEC

Enable [DNSSEC](https://en.wikipedia.org/wiki/Domain_Name_System_Security_Extensions) for a domain and re-sign after changes the zone.

```bash
opencli domains-dnssec <DOMAIN_NAME> [--update | --check]
```

### Enable

Enable DNSSEC for a domain (generates keys and signs the zone):
```bash
opencli domains-dnssec <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dnssec example.com
example.com.		IN DS 24568 7 1 3F2B0C6E4A8D9E1B7C5A2F0D6E8B4A1C9D7E3F5A
example.com.		IN DS 24568 7 2 9A1C7E5B3D2F8A6C4E0B9D7F1A3C5E7B9D2F4A6C8E0B1D3F5A7C9E2B4D6F8A0C
```
</details>

### Check

Check if domain has DNSSEC enabled (displays the DS records):
```bash
opencli domains-dnssec <DOMAIN_NAME> --check
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dnssec example.com --check
example.com.		IN DS 24568 7 1 3F2B0C6E4A8D9E1B7C5A2F0D6E8B4A1C9D7E3F5A
example.com.		IN DS 24568 7 2 9A1C7E5B3D2F8A6C4E0B9D7F1A3C5E7B9D2F4A6C8E0B1D3F5A7C9E2B4D6F8A0C
```
</details>

### Update

Re-sign the zone after DNS changes and reload the DNS service:
```bash
opencli domains-dnssec <DOMAIN_NAME> --update
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-dnssec example.com --update
Zone example.com has been re-signed and DNS service reloaded.
```
</details>


## HSTS

Manage [HSTS](https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security) for a domain.

```bash
opencli domains-hsts <DOMAIN_NAME> [enable|disable]
```

### Status
Check HSTS status for a domain:
```bash
opencli domains-hsts <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-hsts example.com
Strict-Transport-Security header is NOT set for domain example.com
```
</details>

### Enable
Enable HSTS for a domain:
```bash
opencli domains-hsts <DOMAIN_NAME> enable
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-hsts example.com enable
HSTS enabled for example.com
```
</details>

### Disable
Disable HSTS for a domain:
```bash
opencli domains-hsts <DOMAIN_NAME> disable
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-hsts example.com disable
HSTS disabled for example.com
```
</details>


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

<details>
  <summary>Example output</summary>

```bash
# opencli domains-edit example.com
Directory: /home/stefan/docker-data/volumes/stefan_html_data/_data/example.com
```
</details>

### VHost

Open the VirtualHost file for a domain using `nano` edit, and upon saving restart user webserver:
```bash
opencli domains-edit <DOMAIN_NAME> --ws
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-edit example.com --ws
Opening VirtualHosts file in nano: /home/stefan/docker-data/volumes/stefan_webserver_data/_data/example.com.conf
(nano editor opens; after saving and closing it:)
Restarting nginx webserver to save changes..
```
</details>

## Stats

Parse caddy access logs for users domains and generate static html. Requires the `goaccess` module to be enabled.

```bash
opencli domains-stats [USERNAME] [--debug]
```

Generate stats for all domains:
```bash
opencli domains-stats [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-stats
Processed domain example.com for user stefan
Processed domain example.net for user stefan
Skipped empty log file for domain blog.example.com of user stefan
```
</details>

Generate stats for domains owned by a specific user:
```bash
opencli domains-stats <USERNAME> [--debug]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-stats stefan
Processed domain example.com for user stefan
Processed domain example.net for user stefan
```
</details>



## Update NS

Change nameservers for a single or all dns zones.

```bash
opencli domains-update_ns <DOMAIN_NAME>|--all
```

Update the zone file for a specific domain:
```bash
opencli domains-update_ns <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-update_ns example.com
Nameservers have been updated and BIND9 zones reloaded.
```
</details>

Update all zone files:
```bash
opencli domains-update_ns --all [-y]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-update_ns --all
You are about to update all zone files. This action cannot be undone.
Proceed? (y/n)
y
Nameservers have been updated and BIND9 zones reloaded.
```
</details>





## Varnish

Check Varnish status for domain, enable/disable Varnish caching.

```bash
opencli domains-varnish <DOMAIN_NAME> [on|off] [--short]
```

### Status

Display Varnish Cache status for domain:
 
```bash
opencli domains-varnish <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-varnish example.com
Varnish is OFF for domain example.com

# opencli domains-varnish example.com --short
off
```
</details>

### Enable

Enable Varnish Cache for a domain :
 
```bash
opencli domains-varnish <DOMAIN_NAME> on [--short]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-varnish example.com on
Varnish is not running, starting..
Caddy container reloaded.
```
</details>

### Disable

Disable Varnish Cache for a domain:
 
```bash
opencli domains-varnish <DOMAIN_NAME> off [--short]
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-varnish example.com off
Caddy container reloaded.
```
</details>

## Whoowns

Check which username owns a certain domain name.

To check owner for a domain:

```bash
opencli domains-whoowns <DOMAIN_NAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-whoowns example.com
Owner of 'example.com': stefan
```
</details>

To check owner and docker context for a domain:

```bash
opencli domains-whoowns <DOMAIN_NAME> --context
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-whoowns example.com --context
stefan stefan
```
</details>

To check owner and docroot for a domain:

```bash
opencli domains-whoowns <DOMAIN_NAME> --docroot
```

<details>
  <summary>Example output</summary>

```bash
# opencli domains-whoowns example.com --docroot
Owner of 'example.com': stefan | docroot: /var/www/html/example.com
```
</details>

