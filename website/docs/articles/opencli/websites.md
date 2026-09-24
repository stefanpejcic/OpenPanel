# Websites

Manage websites.

### List all websites

List all websites hosted on the server:

```bash
opencli websites-all
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-all
example.com
example.com/blog
shop.example.net
```
</details>

List only websites of a specific type:
```bash
opencli websites-all <TYPE>
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-all wordpress
example.com
example.com/blog
```
</details>

### List websites for user

List all websites owned by user:

```bash
opencli websites-user <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-user stefan
Domain: example.com
PHP version: php8.3
  Sites:
    Site Name: example.com
    Type: wordpress
    Version: 6.6.2
    Ports: 
    Path: /var/www/html/example.com

Domain: example.net
PHP version: php8.1
```
</details>

Additional flags:

- `--type=<TYPE>` - list only websites of this type. Run `opencli websites-user` without a username to see available types.
- `--domains=<domain1,domain2,...>` - list only websites on these domains.
- `--json` - display output as JSON.

<details>
  <summary>Example output</summary>

```bash
# opencli websites-user stefan --json
{ "example.com": {"php": "php8.3", "sites": [{"site_name": "example.com", "type": "wordpress", "version": "6.6.2", "ports": "", "path": "/var/www/html/example.com"}]},"example.net": {"php": "php8.1"} }
```
</details>

### Add websites for user

Scan user files and add WordPress installations:
```bash
opencli websites-scan <USERNAME>
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-scan stefan
- Parsing file: /var/www/html/example.com/wp-config.php
Adding website example.com to Site Manager
Fixing permissions and ownership for the directory /var/www/html/example.com
Scan completed. Detected 1 new WordPress installations:
- example.com, domain: example.com, email: stefan@example.com, version: 6.6.2
```
</details>

use `-all` flag to run for all users:
```bash
opencli websites-scan -all
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-scan -all
Processing user: stefan (1/2)
- Parsing file: /var/www/html/example.com/wp-config.php
  Site example.com already exists in the SiteManager - Skipping
Scan completed. No new WordPress installations detected, but the following 1 existing installations are present:
- example.com - domain: example.com, config: /var/www/html/example.com
------------------------------
Processing user: demo (2/2)
Scan completed. No WordPress installations detected.
------------------------------
DONE.
```
</details>


### Secure

Server-level restrictions to harden WordPress instance. Rules are applied at the webserver level before PHP processing.

List all available rules:
```bash
opencli websites-secure --list-available-rules
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-secure --list-available-rules
wp_manager_wp_config
wp_manager_xmlrpc
wp_manager_uploads_php
wp_manager_wp_includes_php
wp_manager_admin_script_concat
wp_manager_author_scan
wp_manager_sensitive_files
wp_manager_cache_php
wp_manager_env_files
wp_manager_bad_bots
```
</details>

Check if rules exist for a domain:
```bash
opencli websites-secure <domain>
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-secure example.com
Status for example.com: rules enabled
wp_manager_xmlrpc
wp_manager_uploads_php
```
</details>

List enabled rules for domain:
```bash
opencli websites-secure <domain> --list-active-rules
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-secure example.com --list-active-rules
Active rules for example.com:
wp_manager_xmlrpc
wp_manager_uploads_php
```
</details>

Enable rules for a domain:
```bash
opencli websites-secure <domain> --rules='rule1 rule2'
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-secure example.com --rules='wp_manager_xmlrpc wp_manager_uploads_php'
Rules updated for example.com
```
</details>

  
Disable all rules for a domain:
```bash
opencli websites-secure <domain> --disable-all
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-secure example.com --disable-all
All rules removed for example.com
```
</details>

### Vulnerability

Check WordPress website for WP core, theme and plugin vulnerabilities:

```bash
opencli websites-vulnerability <WEBSITE>
```

Example:
```bash
opencli websites-vulnerability pejcic.rs
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-vulnerability pejcic.rs
[*] pejcic.rs
   Core:
      └── WordPress 6.6.2
    PLUGINS:
      ├── akismet  5.3.3 (1/2)
      └── contact-form-7  5.9.8 (2/2)
    THEMES:
      └── twentytwentyfour  1.2 (1/1)
   [-] No vulnerabilities found.
```
</details>

Check vulnerabilities for all WordPress websites hosted on server:

```bash
opencli websites-vulnerability --all
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-vulnerability --all
[*] pejcic.rs [1 of 2]
   Core:
      └── WordPress 6.6.2
    ...
   [-] No vulnerabilities found.
[*] example.com/blog [2 of 2]
   Core:
      └── WordPress 6.5.5
    ...
   [+] Vulnerabilities detected and saved.

Summary: 2 Sites, 5 Plugins, 3 Themes in 00:00:14
```
</details>

### PageSpeed

Get Google PageSpeed data for a single website:

```bash
opencli websites-pagespeed <WEBSITE>
```

Example:
```bash
opencli websites-pagespeed pejcic.rs/blog
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-pagespeed pejcic.rs/blog
Google PageSpeed data saved to /etc/openpanel/openpanel/websites/pejcic.rs_blog.json
```
</details>

Get Google PageSpeed data for all websites hosted on server:

```bash
opencli websites-pagespeed -all
```

<details>
  <summary>Example output</summary>

```bash
# opencli websites-pagespeed -all
Google PageSpeed data saved to /etc/openpanel/openpanel/websites/pejcic.rs.json
Google PageSpeed data saved to /etc/openpanel/openpanel/websites/pejcic.rs_blog.json
```
</details>

> **Note:**  
> Since version **1.2.2**, users can provide their own [PageSpeed API key](https://developers.google.com/speed/docs/insights/v5/get-started) by creating a file named `pagespeed_api_key.txt` in their home directory at `/var/www/html/`.  
>  
> Administrators can also set a system-wide API key by creating the file `/etc/openpanel/openpanel/service/pagespeed.api` and placing the key inside.
