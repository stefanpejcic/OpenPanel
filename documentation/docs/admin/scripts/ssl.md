---
sidebar_position: 10
---

# SSL

Generate and manage SSL certificates for user domains and server hostname.


## Check SSL for all domains owned by user

This command allows you to check the SSL status of all user domains in Certbot and updates the ssl/.domain_status file of the user. This file is used to display SSL status and expiry date on [SSL Certificates page](/docs/panel/domains/SSL) in OpenPanel.

```bash
opencli ssl-user <username>
```

Example:
```bash
# opencli ssl-user stefan
panel.pejcic.rs:  2024-02-20 08:24:00+00:00 (VALID: 89 days)
example.com: None
```


## Check SSL for server hostname

By default when OpenPanel is installed, it will run the 'opencli ssl-hostname' and try to generate SSL for the server hostname, if succeeds, it will set the OpenPanel and AdminPanel to use the SSL and redirect all domains:2083 to the https://hostname.tld:2083 and https://hostname.tld:2087 

If you change the hostname and want the OpenPanel to use the new domain name, you can manually at any time run the following command to generate SSL for the new hostname and force it for OpenPanel: 

```bash
opencli ssl-hostname
```

Example:
```bash
opencli ssl-hostname
No SSL certificate found for server.openpanel.co. Proceeding to generate a new certificate...
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Account registered.
Requesting a certificate for server.openpanel.co

Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/server.openpanel.co/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/server.openpanel.co/privkey.pem
This certificate expires on 2024-02-18.
These files will be updated when the certificate renews.
Certbot has set up a scheduled task to automatically renew this certificate in the background.

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
If you like Certbot, please consider supporting our work by:
 * Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
 * Donating to EFF:                    https://eff.org/donate-le
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
ssl is now enabled and force_domain value in /usr/local/panel/conf/panel.config is set to 'server.openpanel.co'.
Restarting the panel services to apply the newly generated SSL and force domain server.openpanel.co.

- OpenPanel  is now available on: https://server.openpanel.co:2083
- AdminPanel is now available on: https://server.openpanel.co:2087


```
