---
sidebar_position: 1
---

# Domains

To create a website, the first step is to [add a domain name](/docs/panel/domains/#create-a-new-domain).

If the **Domains** module is enabled on the server and your user account has access to it, you'll see a table listing all current domains, the total number of domains, a search bar, and an option to add a new domain.

![Domains page listing domains with their status, document root and PHP version](/img/openpanel-screenshots/domains/domains-list.png#gh-light-mode-only)
![Domains page listing domains with their status, document root and PHP version](/img/openpanel-screenshots/domains/domains-list_dark.png#gh-dark-mode-only)

From this interface, you can view:

* **Domain Status**: Active or Suspended
* **Document Root** (folder where website files will be stored)
* **SSL Settings**: Automatic or Custom
* **Domain Management Options**: Based on enabled features

## Available Actions

Depending on the features enabled on your server, the following actions are available from the **Actions** dropdown menu per domain:

* **Edit DNS Zone** — if the DNS feature is enabled
* **Manage WAF** — if the WAF feature is enabled
* **Change Document Root** — if the docroot feature is enabled
* **Edit VirtualHosts** — if the VHosts editor is enabled
* **Capitalize** — if the Capitalize feature is enabled
* **Suspend / Unsuspend** — if the Suspend feature is enabled
* **Delete**

![Actions menu of a domain with Edit DNS Zone, Manage WAF, Change docroot, Edit VirtualHosts, Capitalize, Suspend and Delete](/img/openpanel-screenshots/domains/domains-actions.png#gh-light-mode-only)
![Actions menu of a domain with Edit DNS Zone, Manage WAF, Change docroot, Edit VirtualHosts, Capitalize, Suspend and Delete](/img/openpanel-screenshots/domains/domains-actions_dark.png#gh-dark-mode-only)

In addition, if the **Redirects** feature is enabled, a dedicated **Redirect** column lets you create, edit, or delete a redirect directly from the table without opening the dropdown menu.

## Bulk Actions

Tick the checkbox of one or more domains, or the checkbox in the table header to select every domain shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![A domain selected with the bulk actions bar offering Suspend, Unsuspend, Change PHP version, Redirect to, Remove redirect, Enable WAF, Disable WAF and Delete](/img/openpanel-screenshots/domains/domains-bulk.png#gh-light-mode-only)
![A domain selected with the bulk actions bar offering Suspend, Unsuspend, Change PHP version, Redirect to, Remove redirect, Enable WAF, Disable WAF and Delete](/img/openpanel-screenshots/domains/domains-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Suspend** | Suspends the selected domains, visitors see a suspended page. |
| **Unsuspend** | Makes the selected domains available again. |
| **Change PHP version** | Switches the selected domains to the PHP version you pick. |
| **Redirect to** | Redirects the selected domains to the URL you enter, starting with `http://` or `https://`. |
| **Remove redirect** | Removes the redirect from the selected domains. |
| **Enable WAF** | Turns on the firewall for the selected domains. |
| **Disable WAF** | Turns off the firewall for the selected domains. |
| **Delete** | Permanently deletes the selected domains, including their websites, files and DNS zones. |

:::info
Actions only show when your plan includes that feature, for example **Suspend** and **Unsuspend** need domain suspension, and **Change PHP version** needs PHP. Selected domains that don't support an action, like a domain that is already suspended, are skipped.
:::

Click an action, confirm it, and it runs on the selected domains one after another. When it's done the page reloads with a notice listing the domains it worked for, or which ones failed and why.

## Create a New Domain

To add a new domain:

1. Click the **"New Domain"** button.
2. Enter the domain name.
3. Click **"Add Domain"** to save.

Unlike other panels, OpenPanel treats all domains equally. From this single interface, you can add **primary domains**, **addon domains**, or **subdomains**.

![New Domain form with the domain name and document root fields](/img/openpanel-screenshots/domains/new-form.png#gh-light-mode-only)
![New Domain form with the domain name and document root fields](/img/openpanel-screenshots/domains/new-form_dark.png#gh-dark-mode-only)

Once added, the system will automatically attempt to issue a free [Let’s Encrypt](https://letsencrypt.org/getting-started/) SSL certificate. If successful, the certificate will be applied immediately.

## Delete a Domain

To delete a domain:

1. Click the **"Delete"** option from the domain's dropdown menu.
2. A confirmation page will appear. Click **"Delete Domain"** to proceed.

![Delete domain confirmation page with the Delete Domain button](/img/openpanel-screenshots/domains/delete-form.png#gh-light-mode-only)
![Delete domain confirmation page with the Delete Domain button](/img/openpanel-screenshots/domains/delete-form_dark.png#gh-dark-mode-only)

> If the domain is linked to active applications (e.g. [Node.js](/docs/panel/applications/nodejs), [Python](/docs/panel/applications/python) or [WP Manager](/docs/panel/applications/wordpress)), deletion will be blocked until those applications are removed.
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
3. Make your changes and click **"Save Changes"**.

Once saved, OpenPanel will automatically restart the webserver to apply changes.
