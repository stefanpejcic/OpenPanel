---
sidebar_position: 10
---

# Custom code

The Custom Code section, accessible via OpenAdmin > Settings > Custom Code, allows you to inject and manage custom code that extends or modifies the behavior and appearance of the OpenPanel UI.

:::info
Custom CSS, Custom JS, Code in Header, Code in Footer, Custom Section, and Welcome Email below require an active **Enterprise** license (and a non-reseller account) to edit. The remaining options on this page (WordPress Plugins/Themes Sets, WordPress profiles, After Update, Before Startup) are available on all license types.
:::

:::info
How-to Articles, PageSpeed API Key, Forbidden Usernames, and Restricted Domains moved to [OpenAdmin > Settings > OpenPanel](/docs/admin/settings/openpanel).
:::

## Custom CSS
Inject your own CSS styles that will be applied across all pages of the OpenPanel UI. 

![Custom CSS editor with the Insert Example button](/img/openadmin-screenshots/settings/custom_code-css.png#gh-light-mode-only)
![Custom CSS editor with the Insert Example button](/img/openadmin-screenshots/settings/custom_code-css_dark.png#gh-dark-mode-only)

Stored in `/etc/openpanel/openpanel/custom_code/custom.css` file.

## Custom JS
Add custom JavaScript to all pages. This is ideal for extending UI functionality, adding widgets, or integrating third-party tools.

![Custom JS editor](/img/openadmin-screenshots/settings/custom_code-js.png#gh-light-mode-only)
![Custom JS editor](/img/openadmin-screenshots/settings/custom_code-js_dark.png#gh-dark-mode-only)

Stored in `/etc/openpanel/openpanel/custom_code/custom.js` file.

## Code in header
Insert custom code directly into the head tag of every page. Useful for meta tags, analytics scripts, and global settings.

![Code in Header editor for the head tag of every OpenPanel page](/img/openadmin-screenshots/settings/custom_code-header.png#gh-light-mode-only)
![Code in Header editor for the head tag of every OpenPanel page](/img/openadmin-screenshots/settings/custom_code-header_dark.png#gh-dark-mode-only)

Stored in `/etc/openpanel/openpanel/custom_code/in_header.html` file.

## Code in footer
Insert custom code directly into the footer section of all pages. This is commonly used for tracking scripts, analytics, or deferred JavaScript.

![Code in Footer editor](/img/openadmin-screenshots/settings/custom_code-footer.png#gh-light-mode-only)
![Code in Footer editor](/img/openadmin-screenshots/settings/custom_code-footer_dark.png#gh-dark-mode-only)

Stored in `/etc/openpanel/openpanel/custom_code/in_footer.html` file.

## Custom section

You can add a **custom section** with icon-based items to the *Dashboard* in **OpenPanel**.

![Custom Section editor for the OpenPanel dashboard](/img/openadmin-screenshots/settings/custom_code-section.png#gh-light-mode-only)
![Custom Section editor for the OpenPanel dashboard](/img/openadmin-screenshots/settings/custom_code-section_dark.png#gh-dark-mode-only)

custom section supports the following fields:

* **`section_title`** *(string)*:
  The title displayed at the top of your custom section.

* **`section_position`** *(string)*:
  Determines where your section appears relative to other built-in sections.
  Acceptable values:

  * `before_files`
  * `before_domains`
  * `before_mysql`
  * `before_postgresql`
  * `before_applications`
  * `before_emails`
  * `before_cache`
  * `before_php`
  * `before_docker`
  * `before_advanced`
  * `before_account`

* **`items`** *(array of objects)*:
  A list of clickable items shown as cards with icons. Each item has:

  * **`label`** *(string)* – The text shown on the card.
  * **`icon`** *(string)* – The icon class from [Bootstrap Icons](https://icons.getbootstrap.com/). Example: `bi bi-person-fill-gear`.
  * **`url`** *(string)* – The link to navigate to when the item is clicked.

Example:
```json
{
  "section_title": "Billing Account",
  "section_position": "before_domains",
  "items": [
    {
      "label": "Manage Profile",
      "icon": "bi bi-person-fill-gear",
      "url": "https://panel.hostio.rs/clientarea.php?action=details"
    },
    {
      "label": "Manage Billing Information",
      "icon": "bi bi-credit-card",
      "url": "https://panel.unlimited.rs/clientarea.php?action=details"
    },
    {
      "label": "View Email History",
      "icon": "bi bi-envelope-open",
      "url": "https://panel.unlimited.rs/clientarea.php?action=emails"
    },
    {
      "label": "News & Announcements",
      "icon": "bi bi-megaphone-fill",
      "url": "https://panel.unlimited.rs/index.php?rp=/announcements"
    },
    {
      "label": "Knowledgebase",
      "icon": "bi bi-book-half",
      "url": "https://panel.unlimited.rs/index.php?rp=/knowledgebase"
    },
    {
      "label": "Server Status",
      "icon": "bi bi-hdd-network",
      "url": "https://panel.unlimited.rs/serverstatus.php"
    },
    {
      "label": "Invoices",
      "icon": "bi bi-receipt",
      "url": "https://panel.unlimited.rs/clientarea.php?action=invoices"
    },
    {
      "label": "Support Tickets",
      "icon": "bi bi-life-preserver",
      "url": "https://panel.unlimited.rs/supporttickets.php"
    },
    {
      "label": "Open Ticket",
      "icon": "bi bi-journal-plus",
      "url": "https://panel.unlimited.rs/submitticket.php"
    },
    {
      "label": "Register New Domain",
      "icon": "bi bi-globe",
      "url": "https://panel.unlimited.rs/cart.php?a=add&domain=register"
    },
    {
      "label": "Transfer Domain",
      "icon": "bi bi-arrow-repeat",
      "url": "https://panel.unlimited.rs/cart.php?a=add&domain=transfer"
    }
  ]
}

```

## Welcome email

HTML message that is placed straight into a new mailbox when the user ticks **Send a welcome email with instructions to set up a mail client** on *OpenPanel > Emails > New*. It is saved into the inbox on the server with `doveadm`, so it is never sent over the internet and doesn't go to any external address.

Click **Restore default** to load the [default template](https://raw.githubusercontent.com/stefanpejcic/openpanel-configuration/refs/heads/main/openpanel/custom_code/welcome.html) into the editor, then click Save.

A live preview of the HTML is shown next to the editor and updates as you type. The preview is sandboxed, so scripts in the HTML don't run.

These placeholders are replaced when the message is created:

| Placeholder | Replaced with |
|---|---|
| `{{EMAIL}}` | the new email address, also used as the username |
| `{{DOMAIN}}` | the domain of the email address |
| `{{MAIL_HOST}}` | the mail server hostname, same as *Emails > Connect Devices* in OpenPanel |
| `{{IMAP_PORT}}` | `993` if the hostname has an SSL certificate, otherwise `143` |
| `{{SMTP_PORT}}` | `465` if the hostname has an SSL certificate, otherwise `587` |
| `{{SECURITY}}` | `SSL/TLS` or `STARTTLS` |
| `{{WEBMAIL_URL}}` | the webmail link |

If the file is empty or missing, no welcome message is added and the email account is still created.

Stored in `/etc/openpanel/openpanel/custom_code/welcome.html` file.

## WordPress plugins set

The global default list of WordPress plugins, installed on new WordPress sites when the user keeps **Default** in the *Plugins & themes set* field of the install form (or when there are no [WordPress profiles](#wordpress-profiles)). If the user has their own [plugins set](/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/), theirs is used instead.
Enter one item per row. Supported formats:

* **wp\_org\_slug** — the plugin slug from the WordPress.org plugin page
* **URL** — a direct link to a `.zip` plugin file hosted online

![WordPress Plugins Set editor](/img/openadmin-screenshots/settings/custom_code-plugins.png#gh-light-mode-only)
![WordPress Plugins Set editor](/img/openadmin-screenshots/settings/custom_code-plugins_dark.png#gh-dark-mode-only)

Stored in `/etc/openpanel/wordpress/sets/plugins.txt` file.

## WordPress themes set

The global default list of WordPress themes, installed on new WordPress sites when the user keeps **Default** in the *Plugins & themes set* field of the install form (or when there are no [WordPress profiles](#wordpress-profiles)). If the user has their own [themes set](/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/), theirs is used instead.
Enter one item per row. Supported formats:

* **wp\_org\_slug** — the theme slug from the WordPress.org themes page
* **URL** — a direct link to a `.zip` theme file hosted online

![WordPress Themes Set editor](/img/openadmin-screenshots/settings/custom_code-themes.png#gh-light-mode-only)
![WordPress Themes Set editor](/img/openadmin-screenshots/settings/custom_code-themes_dark.png#gh-dark-mode-only)

Stored in `/etc/openpanel/wordpress/sets/themes.txt` file.

## WordPress profiles

Named sets of plugins and themes, for example *E-commerce (WooCommerce)*, *Blog* or *Business website*. When at least one profile exists, users get a **Plugins & themes set** field on *OpenPanel > WordPress > Install* where they pick:

- **Default**: the global [plugins set](#wordpress-plugins-set) and [themes set](#wordpress-themes-set) above, or the user's own sets.
- One of your profiles: only that profile's plugins and themes are installed.
- **None, clean WordPress**: no extra plugins or themes are installed.

There are no profiles by default, and until you add one the install form looks the same as before.

To add a profile, open **Add WordPress profile**, enter a name and list the plugins and themes, one per line, in the same formats as the global sets (WordPress.org slug or a URL to a `.zip` file, lines starting with `#` are ignored). Click **Save files**. Existing profiles are listed above it: open one to rename it or change its plugins and themes and click **Save files**, or click **Delete** to remove it.

Each profile is stored as a JSON file in `/etc/openpanel/wordpress/sets/profiles/`, for example:

```json
{
  "name": "E-commerce (WooCommerce)",
  "plugins": ["woocommerce", "woocommerce-payments"],
  "themes": ["storefront"]
}
```

The file name (without `.json`) is the profile's ID, which can also be passed as `wp_set` to the [WordPress install API](/docs/panel/api/#install-wordpress) together with `none` for a clean install.

## After update
Define custom bash commands that will automatically run after each OpenPanel update. Ideal for restoring customizations or triggering automation scripts.

![After update script editor](/img/openadmin-screenshots/settings/custom_code-after.png#gh-light-mode-only)
![After update script editor](/img/openadmin-screenshots/settings/custom_code-after_dark.png#gh-dark-mode-only)

This powerful customization layer helps ensure OpenPanel fits seamlessly into your environment.

Examples:

- [Custom email templates](/docs/articles/dev-experience/customizing-openadmin-email-templates/)


Stored in `/root/openpanel_run_after_update` file.

## Before startup

Add custom bash code to be executed before starting OpenPanel.

![Before startup script editor](/img/openadmin-screenshots/settings/custom_code-before.png#gh-light-mode-only)
![Before startup script editor](/img/openadmin-screenshots/settings/custom_code-before_dark.png#gh-dark-mode-only)

Stored in `/root/openpanel_run_on_startup` file.


