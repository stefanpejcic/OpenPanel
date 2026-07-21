---
sidebar_position: 5
---

# SSL

The SSL/TLS Settings page enables you to check whether domains have SSL certificates generated, view the private key or public file for the certificate, renew SSL, generate new certificates, or delete existing ones.

![domain_ssl.png](/img/panel/v1/domains/domain_ssl.png)

The system will attempt to automatically generate a free Let's Encrypt SSL [when a new domain is added](/docs/panel/domains/#adding-a-domain). If successful, you will see the 'AutoSSL Domain Validated' message next to the domain name.

However, if SSL generation fails, the page will display 'No SSL Installed' next to the domain name, and you will have the option to generate a new SSL by clicking the 'Generate' button.

### Generate SSL

To generate a Free Let'sEncrypt SSL for the domain click on the 'Generate' button.

![domain_ssl_generate.png](/img/panel/v1/domains/domain_ssl_generate.png)

### Renew SSL

The system will automatically check all your certificates every night and renew them if needed. However, you can also manually renew a certificate by clicking on the 'Renew' button next to it.

![domain_ssl_renew.png](/img/panel/v1/domains/domain_ssl_renew.png)

### Delete an SSL

To remove an SSL certificate for a domain, click on the 'Delete' button next to it.

![domain_ssl.png](/img/panel/v1/domains/domain_ssl.png)


### View Certificate

To view the certificate file for a domain, click on the 'View Certificate' link. This will open a modal displaying the certificate in plain text.

![domain_ssl_view_certificate.png](/img/panel/v1/domains/domain_ssl_view_certificate.png)

### View Public Key

To view the private key file for a certificate, click on the 'Private Key' link. This will open a modal displaying the certificate private key in plain text.

![domain_ssl_view_private.png](/img/panel/v1/domains/domain_ssl_view_private.png)

### Renew All SSLs

The system will automatically check all your certificates every night and renew them if necessary. However, you can also manually run the renewal for all certificates by clicking on the 'Run AutoSSL for All Domains' button.

![domains_ssl_renew_all.png](/img/panel/v1/domains/domains_ssl_renew_all.png)
