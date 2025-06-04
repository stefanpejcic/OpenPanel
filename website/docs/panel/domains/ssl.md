---
sidebar_position: 5
---

# SSL

OpenPanel automatically generates and renews SSL certificates for all domains using **Let's Encrypt** or **ZeroSSL**.

You can also configure a **custom SSL certificate** for any domain via **OpenPanel > Domains > SSL**.

---

## Custom SSL

To use your own SSL certificate:

1. Upload your certificate files via **OpenPanel > FileManager**.
2. Go to **OpenPanel > Domains**, then click **SSL** next to the domain.
3. Under **Configure custom SSL**, enter the file paths for:
   * Certificate file (`.crt`)
   * Private key file (`.key`)
4. Click **Configure**.

Once configured, your custom certificate details will appear on the same page.

![screenshot of domain with custom ssl](/img/panel/v2/openpanel_customssl.png)


## AutoSSL

**AutoSSL** is the default option in OpenPanel.
If you're adding a new domain, no action is required.

To switch **from a custom certificate back to AutoSSL**:

1. Navigate to **OpenPanel > Domains** and click **SSL** for the domain.
2. Click **Switch back to AutoSSL**.

The certificate will be re-issued automatically after the domain is accessed via `https://`. Once generated, it will be displayed on the same page.

![screenshot of domain with autossl](/img/panel/v2/openpanel_autossl.png)


### Requirements

To ensure successful SSL generation:

* The **A record** for the domain must point to the server's **IPv4 address**.
* The DNS must be **fully propagated**. Use tools like [whatsmydns.net](https://www.whatsmydns.net/#A) to check.
* The domain must be accessed via `https://` at least once to trigger certificate generation. Open the domain in a browser using `https`.
If:

* The domain was just added,
* DNS is not yet pointed to the server, or
* The domain has not been accessed over `https://`,

Then the SSL section will show **“No certificate found.”**

![screenshot of domain with autossl but no ssl yet](/img/panel/v2/openpanel_autossl_no_ssl.png)
