---
sidebar_position: 5
---

# SSL

OpenPanel automatically generates and renews SSL certificates for all domains using **Let's Encrypt**.

You can also configure a **custom SSL certificate** for any domain via **OpenPanel > Domains > SSL Certificates**.

## Custom SSL

To use your own SSL certificate:

1. Go to **OpenPanel > Domains**, then click **SSL** next to the domain.
2. Under **Configure custom SSL**, paste the contents of your certificate and private key (PEM format, including the `-----BEGIN CERTIFICATE-----`/`-----BEGIN PRIVATE KEY-----` lines) into the **Certificate** and **Private Key** fields.
3. Click **Configure Custom Certificate**.

Once configured, your custom certificate details will appear on the same page, and the SSL status will change to **Custom SSL**.

![Configure custom SSL form with fields for the certificate and the private key](/img/openpanel-screenshots/domains/ssl-custom.png#gh-light-mode-only)
![Configure custom SSL form with fields for the certificate and the private key](/img/openpanel-screenshots/domains/ssl-custom_dark.png#gh-dark-mode-only)


## AutoSSL

**AutoSSL** is the default option in OpenPanel.
If you're adding a new domain, no action is required — a certificate is requested automatically the first time the domain is accessed over `https://`.

If the status shows **Auto SSL** but no certificate has been issued yet, click **Generate now** to reload the webserver and trigger issuance immediately.

To switch **from a custom certificate back to AutoSSL**:

1. Navigate to **OpenPanel > Domains** and click **SSL** for the domain.
2. Click **Switch to Let's Encrypt and generate**.

This switches the domain back to AutoSSL and immediately attempts to generate the certificate. Once issued, it will be displayed on the same page.

![SSL Status row showing Auto SSL with the Generate now button](/img/openpanel-screenshots/domains/ssl-status.png#gh-light-mode-only)
![SSL Status row showing Auto SSL with the Generate now button](/img/openpanel-screenshots/domains/ssl-status_dark.png#gh-dark-mode-only)


### Requirements

To ensure successful SSL generation:

* The **A record** for the domain must point to the server's **IPv4 address**.
* The DNS must be **fully propagated**. Use tools like [whatsmydns.net](https://www.whatsmydns.net/#A) to check.
* The domain must be accessed via `https://` at least once to trigger certificate generation. Open the domain in a browser using `https`.

If:

* The domain was just added,
* DNS is not yet pointed to the server, or
* The domain has not been accessed over `https://`,

Then the SSL section will show **“No Certificate!”**

![SSL page for a domain with the AutoSSL status, the Generate now button, and the certificate details and files](/img/openpanel-screenshots/domains/ssl-page.png#gh-light-mode-only)
![SSL page for a domain with the AutoSSL status, the Generate now button, and the certificate details and files](/img/openpanel-screenshots/domains/ssl-page_dark.png#gh-dark-mode-only)
