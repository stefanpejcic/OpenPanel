---
sidebar_position: 1
---

# Domains

To create a website, the first step is to [add a domain name](/docs/panel/domains/#create-a-new-domain).

If the **Domains** module is enabled on the server and your user account has access to it, you'll see a table listing all current domains, the total number of domains, a search bar, and an option to add a new domain.

From this interface, you can view:

* **Domain Status**: Active or Suspended
* **Document Root** (folder where website files will be stored)
* **SSL Settings**: Automatic or Custom
* **Domain Management Options**: Based on enabled features

## Available Actions

Depending on the features enabled on your server, the following actions are available per domain:

* **Edit DNS Zone** — if the DNS feature is enabled
* **Manage WAF** — if the WAF feature is enabled
* **Change Document Root**
* **Edit VirtualHosts**
* **Redirect** — if the Redirect feature is enabled
* **Capitalize** — if the Capitalize feature is enabled
* **Suspend / Unsuspend Domain**
* **Delete Domain**

## Create a New Domain

To add a new domain:

1. Click the **"Add Domain"** button.
2. Enter the domain name.
3. Click **"Add Domain"** to save.

Unlike other panels, OpenPanel treats all domains equally. From this single interface, you can add **primary domains**, **addon domains**, or **subdomains**.

Once added, the system will automatically attempt to issue a free [Let’s Encrypt](https://letsencrypt.org/getting-started/) SSL certificate. If successful, the certificate will be applied immediately.

## Delete a Domain

To delete a domain:

1. Click the **"Delete Domain"** option from the domain's dropdown menu.
2. A confirmation page will appear. Click **"Delete"** to proceed.

> If the domain is linked to active applications (e.g. [PM2](/docs/panel/applications/pm2) or [WP Manager](/docs/panel/applications/wordpress)), deletion will be blocked until those applications are removed.
> This prevents accidental removal of domains tied to running websites.

Deleting a domain will **permanently remove** the following:

1. Nginx configuration file
2. DNS zone file
3. SSL certificate
4. IP Blocker rules for the domain
5. Redirects associated with the domain

## Redirects

### Add Redirect

To create a redirect:

1. Click the **"Create Redirect"** button next to the domain.
2. Enter the full URL (must start with `http://` or `https://`).
3. Click **"Save"** to apply or **"Cancel"** to discard.

### Edit Redirect

Click the **pencil icon** next to an existing redirect URL to modify it.

### Delete Redirect

Click the **cross icon** next to the redirect URL to remove it.

## Edit VirtualHosts File

The **VirtualHosts** file defines the configuration for the domain within Nginx or Apache. It includes settings such as:

* Access logs
* PHP version
* Application runners
* Redirect rules
* Custom directives

To edit this file:

1. Click **"Edit VirtualHosts"** from the domain’s dropdown menu.
2. A new page will open with the Vhost file content.
3. Make your changes and click **"Save"**.

Once saved, OpenPanel will automatically restart the webserver to apply changes.
