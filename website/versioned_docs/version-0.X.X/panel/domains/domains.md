---
sidebar_position: 1
---

# Domain Management

To create a website, it's necessary to first [add a domain name](/docs/panel/domains/#adding-domains).

![domains.png](/img/panel/v1/domains/domains.png)

## Available Actions

Through the Domains interface, you have the ability to perform various domain-related actions, including:

- **Add a Domain Name**: add a new domain name, either a subdomain or top-level domain.
- **Delete a Domain Name**: delete a domain name permanently removes it from the account.
- **Redirect to Another Domain**: set up a domain redirection, specifying that one domain should redirect to another.
- **Edit DNS Records**: modify DNS records associated with the domain.
- **Edit Virtual Hosts File**: capability to edit the virtual hosts configuration file for the domain.
- **Force HTTPS**: if domain has SSL Certificate installed you can force all incoming traffic trough https.


  
## Adding a Domain


To add a new domain, click on the 'Add Domain' button, enter the domain name, and then click on the 'Add Domain' button.

![add_domain.png](/img/panel/v1/domains/add_domain.png)

Unlike other hosting panels, OpenPanel has a single interface 'Domains' where you can add your primary domain, addon domains or subdomains.

When new domain name is added, the system will automatically try to generate a free [Let's Encrypt](https://letsencrypt.org/getting-started/) certificate. If successful, the certificate is automatically applied.

:::info
OpenPanel supports [Internationalized domain names (IDNs)](https://en.wikipedia.org/wiki/Internationalized_domain_name), and if such a domain is added, it will be automatically converted to [punycode](https://en.wikipedia.org/wiki/Punycode).
:::

## Delete a Domain

To delete a domain name click on the 'Delete domain' button in the dropdown options for the domain.

![delete_domain_1.png](/img/panel/v1/domains/delete_domain_1.png)

A new modal will appear, asking you to confirm the deletion process by clicking on the 'Delete' button. Upon clicking the button, the domain will be instantly deleted.

![delete_domain_2.png](/img/panel/v1/domains/delete_domain_2.png)

If the domain name has websites associated with it, for instance, in the [PM2](/docs/panel/applications/pm2) or [WP Manager](/docs/panel/applications/wordpress) interfaces, the system will block the deletion of the domain name until the listed websites are deleted first.

This is a fallback mechanisam designed to prevent users from accidentally deleting domains that have running webistes.

![delete_domain_3.png](/img/panel/v1/domains/delete_domain_3.png)

Deleting a domain name will also permanently delete the following files:

1. Nginx configuration file
2. DNS zone file
3. SSL certificate
4. IP Blocker configuration for the domain
5. Redirects for the domain

:::info
It's important to note that domain deletion does not affect any website files or databases associated with the domain. Those assets remain intact and are not impacted by the domain removal process.
:::


## Redirects


### Add redirect

To redirect a domain name to another URL, click on the 'Create Redirect' button next to it. In the field, set the URL to redirect to, ensuring that the redirect starts with either the `http://` or `https://` prefix.

To save, click on the 'Save' button, and to discard, click on the 'Cancel' button.

![domain_redirect_1.png](/img/panel/v1/domains/domain_redirect_1.png)

### Edit Redirect

To edit an existing redirect, click on the pencil icon next to the redirect URL for that domain.
![domain_redirect_2.png](/img/panel/v1/domains/domain_redirect_2.png)

### Delete Redirect

To delete a redirect, click on the cross icon next to the redirect URL for that domain.

![domain_redirect_2.png](/img/panel/v1/domains/domain_redirect_2.png)

## HTTP/2


HTTP/2 is an HTTP version that makes applications faster, simpler, and more robust in comparison with HTTP/1. To be able to use HTTP/2 you need to set up the connection through HTTPS. To make your website support HTTPS, enable the option Force HTTPS fot that domain.

Starting OpenPanel [version 0.1.4](/docs/changelog/0.1#014) the HTTP/2 support is enabled automatically for all domains that have SSL and no further setting is needed.


## HTTP/3

HTTP/3 is the latest version of the HTTP protocol, designed to improve web performance, security, and reliability even further than HTTP/2. It uses QUIC, a transport layer network protocol that reduces latency and improves connection speeds by allowing multiplexed connections over UDP, instead of TCP. This results in faster page loads and smoother browsing experiences, especially on mobile networks.

To enable HTTP/3 for your domain, you need to ensure that your website is running over HTTPS. You can enable the **Force HTTPS** option for that domain to ensure all traffic is served securely.

Starting from OpenPanel [version 0.3.0](/docs/changelog/0.3.0/#-polish), HTTP/3 support is automatically enabled for all domains that have SSL, so no additional configuration is required if you already have SSL set up. This ensures that your website can benefit from the speed and performance improvements of HTTP/3 seamlessly.



## Force HTTPS

The force HTTPS option can be enabled for each domain name that has a valid SSL certificate.

When enable, the option will automatically upgrade all http requests for the domain to https protocol.

To force https for a domain check the force https toggle next to it.


## Edit VirtualHosts file

The VirtualHosts file serves as the configuration file for the domain within Nginx or Apache webservers. This file contains essential information about access logs, the PHP version in use, redirects, running applications, and other relevant settings.

To edit the Vhost file for a domain click on the 'Edit VirtualHosts' button in the dropdown for that domain:

![domain_edit_vhost_1.png](/img/panel/v1/domains/domain_edit_vhost_1.png)


A new page will open with the content of the VirtualHosts file for the domain. Make the necessary changes and click on the 'Save' button when finished.

Upon saving, Nginx/Apache service restart will be triggered to immediately apply the changes.

![domain_edit_vhost_2.png](/img/panel/v1/domains/domain_edit_vhost_2.png)

:::danger
Editing VirtualHosts file is intended for advanced users only. This action can potentially lead to server misconfiguration or downtime if not done correctly.

Please be cautious and make sure you understand the changes you are making. OpenPanel uses a rollback mechanism to check for errors, and if any issues are detected, system will try to revert the changes. However, it's important to have a backup of your configurations before making any changes.
:::
