---
sidebar_position: 2
---

# Domains

Manage domains: Add, Delete, detect, etc.

## List user domains

To list all domains owned by a user run the following command:


```bash
opencli domains-user <USERNAME>
```

Example:
```bash
# opencli domains-user stefan
panel.pejcic.rs
example.openpanel.co
```

## List all domains

To list all domains owned by all users run the following command:


```bash
opencli domains-all
```

Example:
```bash
# opencli domains-all
panel.pejcic.rs
example.openpanel.co
example.net
...
```

## Check who owns a domain name

To check which user owns a domain name run the following command:

```bash
opencli domains-whoowns <DOMAIN-NAME>
```

Example:
```bash
opencli domains-whoowns pejcic.rs
Owner of 'pejcic.rs': stefan
```

The `whoowns` script searches the database in order to determine which username added a domain.

## Parse domain access logs

To parse domain (Nginx) access logs and generate static reports for users domains accessible from `Domains > Access Logs` run the script:

```bash
opencli domains-stats [USERNAME]
```

Example:
```bash
opencli domains-stats stefan
Processing user: stefan
Processed domain pejcic.rs for user stefan
Processed domain openpanel.co for user stefan
```

To parse (Nginx) access logs for all active users and their domains run the script without username:

```bash
opencli domains-stats
```

## Enable modsecurity

To enable modsecurity for all domains owned by a specific user:

```bash
opencli domains-enable_modsec [USERNAME]
```

To enable modsecurity on all domains for all active users:

```bash
opencli domains-enable_modsec
```
