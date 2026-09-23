# Websites

Manage websites.

### List all websites

List all websites hosted on the server:

```bash
opencli websites-all
```

### List websites for user

List all websites owned by user:

```bash
opencli websites-user <USERNAME>
```

### Add websites for user

Scan user files and add WordPress installations:
```bash
opencli websites-scan <USERNAME>
```

use `-all` flag to run for all users:
```bash
opencli websites-scan -all
```


### Secure

Server-level restrictions to harden WordPress instance. Rules are applied at the webserver level before PHP processing.

List all available rules:
```bash
opencli websites-secure --list-available-rules
```

Check if rules exist for a domain:
```bash
opencli websites-secure <domain>
```

List enabled rules for domain:
```bash
opencli websites-secure <domain> --list-active-rules
```

Enable rules for a domain:
```bash
opencli websites-secure <domain> --rules='rule1 rule2'
```
  
Disable all rules for a domain:
```bash
opencli websites-secure <domain> --disable-all
```

### Vulnerability

Check WordPress website for WP core, theme and plugin vulnerabilities:

```bash
opencli websites-vulnerability <WEBSITE>
```

Example:
```bash
opencli websites-vulnerability pejcic.rs
```

Check vulnerabilities for all WordPress websites hosted on server:

```bash
opencli websites-vulnerability --all
```

### PageSpeed

Get Google PageSpeed data for a single website:

```bash
opencli websites-pagespeed <WEBSITE>
```

Example:
```bash
opencli websites-pagespeed pejcic.rs/blog
```

Get Google PageSpeed data for all websites hosted on server:

```bash
opencli websites-pagespeed -all
```

> **Note:**  
> Since version **1.2.2**, users can provide their own [PageSpeed API key](https://developers.google.com/speed/docs/insights/v5/get-started) by creating a file named `pagespeed_api_key.txt` in their home directory at `/var/www/html/`.  
>  
> Administrators can also set a system-wide API key by creating the file `/etc/openpanel/openpanel/service/pagespeed.api` and placing the key inside.
