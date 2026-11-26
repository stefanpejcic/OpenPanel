---
sidebar_position: 2
---

# OpenPanel

Settings for the OpenPanel user panel, grouped in sections: how it looks, what users see on the dashboard, login and account rules, domains, the File Manager, databases, history, WordPress and the AI assistant.

---

## Branding

- **Brand name:** Shown in the OpenPanel sidebar and on login pages when no logo is set.
- **Favicon:** Icon shown in the browser tab on every OpenPanel page. Enter the full URL of a `.ico`, `.png` or `.svg` image, at least 32×32px. Leave empty to use the OpenPanel icon.
- **Logo:** Shown instead of the brand name on light backgrounds: the login page in light mode, and the sidebar when the theme gives it a light colour.
- **Logo for dark backgrounds:** A light-coloured version of the logo for the login page in dark mode and a dark sidebar, which is the Modern theme's default in both modes. Leave empty to use the Logo everywhere.
- **Logout URL:** Where users are sent after logging out, usually your main website.

Each logo is either uploaded or set as a URL. When no logo is set, the field offers an upload, and **Use an image URL instead** switches to a URL field. Once a logo is set, its path or URL is shown, with **Upload a new image** to replace it. Uploads can be PNG, JPEG, WebP, GIF or SVG, up to 1 MB, recommended size 200×36px. They're saved as `/etc/openpanel/openpanel/brand/logo.<ext>` and `logo_dark.<ext>` and served by OpenPanel at `/brand/`. A new upload replaces the old file of the same logo, and switching a logo to a URL or clearing it removes its upload. The other logo is never touched.

![Branding section with the brand name, favicon, logo and logout URL](/img/openadmin-screenshots/settings/openpanel-branding.png#gh-light-mode-only)
![Branding section with the brand name, favicon, logo and logout URL](/img/openadmin-screenshots/settings/openpanel-branding_dark.png#gh-dark-mode-only)

## Theme & layout {#display}

- **Theme:** The look of the OpenPanel user panel. **Default** uses the templates built into OpenPanel. Below it are the installed themes (folders in `/etc/openpanel/openpanel/templates/`, shown with their screenshot, version, author and license from the theme's `readme.txt`) and the themes from the official list that aren't installed yet, each with an **Install** button. The chosen theme is saved as `template=` and applies after OpenPanel restarts. To build your own theme or submit one to the list, see [How to create and install themes](/docs/articles/dev-experience/openpanel-themes/).
  - **Update to x.y.z** shows on an installed theme when the official list has a newer version. It replaces the theme files and keeps the colours saved on its Customize page.
  - **Uninstall** deletes the theme's files. Users who picked it for themselves go back to the server theme, and if it was the server theme, OpenPanel goes back to Default.
- <span id="users-can-switch-theme"></span>**Users can switch theme:** **No, everyone uses the selected theme** (default) or **Yes, users can pick any installed theme**. When it's on, users pick any installed theme for themselves from their account menu and in the first step of [Onboarding](/docs/panel/dashboard/onboarding/#steps), and the theme selected above stays the default for everyone who doesn't. Saved as `allow_user_theme=` in `openpanel.config`.
- **Menu style:** The default menu for OpenPanel users, **Classic** (expandable groups) or **Modern** (one link per area, with that area's pages as tabs at the top of each page). Users can still switch to the other style for themselves, see [Menu Style](/docs/panel/dashboard/menu-style/). The same setting also switches OpenAdmin's own menu, right after saving.
- **Avatar type:** Gravatar, Letter, or Icon for user avatars.
- **Charts mode for resource usage:** 1 chart, 2 charts, or no charts on the Resource Usage page.
- **Display country flag icons:** Show country flags next to the last login IP on the OpenPanel dashboard.

![Theme and layout section with the theme picker and the OpenPanel interface options](/img/openadmin-screenshots/settings/openpanel-display.png#gh-light-mode-only)
![Theme and layout section with the theme picker and the OpenPanel interface options](/img/openadmin-screenshots/settings/openpanel-display_dark.png#gh-dark-mode-only)

## Dashboard

- **Display onboarding wizard:** See [Onboarding wizard](#onboarding) below.
- **Display 2FA widget:** Show a message on users' dashboards encouraging them to enable Two-Factor Authentication.
- **Display how-to guides widget:** Show how-to articles on users' dashboards. When enabled, a **How-to articles** field appears below it to edit the Knowledge Base articles shown (needs an active Enterprise license and a non-reseller account to edit).
- **Applications shown on the dashboard:** Which applications are listed on the users' dashboard.
- <span id="docker-apps"></span>**Docker apps on the Auto Installer:** Which one-click Docker apps users with the **Docker** feature see in the *Docker applications* section of the [Auto Installer](/docs/panel/applications/autoinstaller/#docker-applications). Every app from `/etc/openpanel/docker/templates/applications.yml` is shown as a button, blue ones are shown to users, click one to hide or show it, and **Restore default** shows them all again. Only the hidden apps are saved, as `hidden_docker_apps=` in the `[PANEL]` section of `openpanel.config`, so apps added to the list in a later update show up on their own. It changes what the Auto Installer lists, not what users can run: with the Docker feature they can still add any image from *Containers*. Applies after OpenPanel restarts.
- **Display link to report bugs:** Show a "Found a bug? Let us know" link at the bottom of all user pages.

![Dashboard section with the onboarding wizard, 2FA and how-to widgets, dashboard applications, Docker apps on the Auto Installer and the report bugs link](/img/openadmin-screenshots/settings/openpanel-dashboard.png#gh-light-mode-only)
![Dashboard section with the onboarding wizard, 2FA and how-to widgets, dashboard applications, Docker apps on the Auto Installer and the report bugs link](/img/openadmin-screenshots/settings/openpanel-dashboard_dark.png#gh-dark-mode-only)

### Onboarding wizard {#onboarding}

When enabled, new users get a short guided setup wizard on their first login. Steps the user has already set up are skipped (Webserver once they have a domain, Database once they have a database, Backups once a destination is configured, 2FA once it's enabled). See the [Onboarding page](/docs/panel/dashboard/onboarding) for what users see.

The wizard doesn't have its own separate list of steps to configure - each step just mirrors a feature already enabled in [Feature Manager](/docs/admin/plans/feature-manager) (Webserver, PHP, Database, Varnish, Backups, WAF, IP Blocker, 2FA, Passkeys). Disable a feature there for a plan and its onboarding step disappears too for users on that plan.

To reset the wizard for a single user (for testing, or if they ask to see it again), delete their marker file over SSH:

```bash
rm /home/USERNAME/onboarding.completed
```

They'll be offered the wizard again next time they log in, minus any steps they've already set up.

You can also add your own custom steps to the wizard - for example, a step to enable Redis - without modifying OpenPanel itself. See [Add Custom Steps to the Onboarding Wizard](/docs/articles/dev-experience/add-custom-steps-to-onboarding-wizard).

## Login & sessions {#login-sessions}

- **Enable password reset on login:** Allow users to reset passwords from the login form (not recommended for security reasons). When enabled without SMTP configured, a warning links to [SMTP settings](/docs/admin/settings/notifications#smtp-settings), since reset emails won't reach users until it's set up.
- **Captcha provider:** Disabled, Google reCAPTCHA, Cloudflare Turnstile, or Custom. Depending on the provider, enter its **site key** and **secret key** (or the custom site key for Custom).
- **Failed logins per minute before rate limiting:** Failed login attempts allowed per minute from one IP before that IP is temporarily rate-limited.
- **Failed logins per minute before a 1 hour block:** Failed login attempts per minute that block the IP for 1 hour.
- **Validate IP address on session cookie:** A session cookie is rejected when the request's IP doesn't match the IP it was issued to.
- **Session duration (in minutes):** How long a session stays active without activity before the user has to log in again.
- **Session lifetime (in minutes):** The longest a session can last, after which the user is logged out regardless of activity.
- **Terminal commands timeout (in seconds):** How long a Docker Terminal command can run before it's stopped.

:::info
Captcha requires the [captcha plugin](https://github.com/stefanpejcic/captcha/) to be installed, these settings have no effect until it is installed.
:::

A warning is shown when Validate IP address on session cookie is off (a cookie stolen from a user, for example through phishing, can be reused from any IP), and when Terminal commands timeout is under 30 seconds (it applies to both the OpenPanel web terminal and OpenAdmin).

![Login and sessions section with the captcha, rate limit and session options](/img/openadmin-screenshots/settings/openpanel-security.png#gh-light-mode-only)
![Login and sessions section with the captcha, rate limit and session options](/img/openadmin-screenshots/settings/openpanel-security_dark.png#gh-dark-mode-only)

## Passwords & 2FA {#passwords}

- **Minimum password strength:** A value between 1-100 setting the minimum password strength for all password fields in OpenPanel (account, FTP, emails, databases and so on).
- **Check passwords against Weakpass.com:** Passwords set on account creation and reset are checked against Weakpass.com's list of compromised passwords.
- **Enforce 2FA:** After login, users are sent to the 2FA setup page and can't open other pages until 2FA is configured.

A warning is shown when Minimum password strength is below 50 (weak passwords are accepted) and when the Weakpass.com check is off (passwords found in data breaches can be used).

![Passwords and 2FA section with the minimum password strength, Weakpass check and Enforce 2FA](/img/openadmin-screenshots/settings/openpanel-passwords.png#gh-light-mode-only)
![Passwords and 2FA section with the minimum password strength, Weakpass check and Enforce 2FA](/img/openadmin-screenshots/settings/openpanel-passwords_dark.png#gh-dark-mode-only)

## Accounts

- **Allow users to change username:** Users can change their username from the *Account > Settings* page.
- **Forbidden usernames:** Usernames that can't be used, one per line.

![Accounts section with what OpenPanel users can change](/img/openadmin-screenshots/settings/openpanel-users.png#gh-light-mode-only)
![Accounts section with what OpenPanel users can change](/img/openadmin-screenshots/settings/openpanel-users_dark.png#gh-dark-mode-only)

## Domains & DNS {#nameservers}

- **NS1 to NS4:** The nameservers shown to users. [Guide on how to properly configure nameservers](/docs/articles/domains/how-to-configure-nameservers-in-openpanel). If no nameservers are set but the DNS module is enabled, a recommendation is shown to either set at least ns1 and ns2, or [disable the DNS module](/docs/admin/settings/modules) if it won't be used, so BIND9 isn't autostarted and DNS zone files aren't created for every domain.
- **DNS record types users can add and edit:** Record types offered in the [DNS Zone Editor](/docs/panel/domains/dns/). See [DNS record types](#dns-record-types).
- **Allow subdomain sharing:** Users can add subdomains for domains that another user owns.
- **Restricted domains:** Domains that can't be added, one per line.
- **Domains per page:** Domains shown per page on the Domains list.

![Domains and DNS section with the nameserver fields](/img/openadmin-screenshots/settings/openpanel-nameservers.png#gh-light-mode-only)
![Domains and DNS section with the nameserver fields](/img/openadmin-screenshots/settings/openpanel-nameservers_dark.png#gh-dark-mode-only)

### DNS record types

Choose which record types users can add and edit in the OpenPanel [DNS Zone Editor](/docs/panel/domains/dns/). Click a type to turn it on or off, or click **Restore default** to go back to the default list.

| Type | On by default | Used for |
|---|---|---|
| `A`, `AAAA` | Yes | IPv4 and IPv6 address of a name |
| `CAA` | Yes | Which certificate authorities can issue SSL for the domain |
| `CNAME` | Yes | Alias to another name |
| `MX` | Yes | Mail servers for the domain |
| `SRV` | Yes | Host and port of a service |
| `TLSA` | Yes | Certificate pinning with DANE |
| `TXT` | Yes | Text values, such as SPF and verification records |
| `HTTPS`, `SVCB` | No | Connection hints for browsers and other clients |
| `NS` | No | Delegates a subdomain to other nameservers, the domain itself is refused |
| `SSHFP` | No | SSH host key fingerprints |
| `URI` | No | Maps a name to a URL |
| `DMARC` | No | Form helper that saves a `TXT` record on `_dmarc`, not a real record type |

Save the settings and restart OpenPanel for the change to apply.

What changes for users:

- The add and edit forms only offer the allowed types, and adding or editing any other type is refused, also through the API.
- Existing records of other types stay in the zone and keep working. Users can't edit them, but they can delete them.
- The [Advanced Editor](/docs/panel/domains/dns/#advanced-editor) edits the zone file directly, so it isn't limited by this setting.

The value is saved as `dns_record_types` and can also be [set with OpenCLI](/docs/opencli/config/#dns_record_types).

## File manager {#files}

The options are grouped in **Display**, **Size limits**, **Time limits**, **File types** and **Trash**:

- **Files per page:** Files and folders shown per page.
- **File Manager button style:** How the toolbar buttons look.
- **Max file size for the viewer:** Largest file (in MB) that opens in the Viewer. Recommended maximum is 20 MB.
- **Max file size for the editor:** Largest file (in MB) that opens in the Code Editor. Recommended maximum is 10 MB.
- **Max file size for uploads:** Largest file (in MB) that can be uploaded. Recommended maximum is 2000 MB.
- **Max file size for downloads:** Largest file (in MB) that can be downloaded. Recommended maximum is 2000 MB.
- **Max time for compressing, extracting and downloads:** Minutes before an archive compression, extraction or download is stopped. Recommended to keep these to a few minutes.
- **Extensions that can be viewed and edited (text):** File extensions that open in the editor, should be text file types.
- **Extensions that can be viewed (images):** Image extensions that the Viewer shows.
- **Extensions that can be extracted (archives):** Archive extensions the File Manager can extract.
- **Auto-purge trash after:** Days files stay in the user's trash before they're deleted. 0 turns auto-purge off. Recommended is 30 days.

A warning is shown when:

- a max file size for the Viewer, Editor, Upload or Download is bigger than the RAM limit of the OpenPanel container (set on [Service Limits](/docs/admin/services/limits)), since files that big can fail.
- Auto-purge trash is set to 0, since deleted files then pile up in users' Trash folders until they empty it themselves.
- Auto-purge trash is over 30 days, recommending to change the `opencli files-purge_trash` schedule to weekly on [Scheduled Actions](/docs/admin/advanced/crons).

![File Manager section with the file manager options](/img/openadmin-screenshots/settings/openpanel-filemanager.png#gh-light-mode-only)
![File Manager section with the file manager options](/img/openadmin-screenshots/settings/openpanel-filemanager_dark.png#gh-dark-mode-only)

## Databases

Settings for user's databases:

- **Max Startup Time for MySQL:** Maximum number of seconds to wait for MySQL to initialize before timing out.
- **Max File Size for MySQL Import:** Maximum file size (in GB) allowed for a MySQL import.
- **Restricted MySQL (system) Users:** List of MySQL usernames that users are not allowed to access/manage.
- **Restricted MySQL (system) Databases:** List of MySQL databases that users are not allowed to access/manage.
- **Enforce username prefix:** MySQL usernames and database names created by users must start with their `context_` prefix.

A warning is shown when:

- Max Startup Time for MySQL is under 10 seconds, since MySQL needs time to initialize on first start. If a custom `MYSQL_VERSION` is set for new users, at least 60 seconds is recommended, since the first start downloads that image from Docker Hub.
- Max File Size for MySQL Import is bigger than the RAM limit of the OpenPanel container.

![Databases section with the database options](/img/openadmin-screenshots/settings/openpanel-databases.png#gh-light-mode-only)
![Databases section with the database options](/img/openadmin-screenshots/settings/openpanel-databases_dark.png#gh-dark-mode-only)

## History & logs {#history}

- **Login records to keep per user:** Recent logins stored per user.
- **Activity records to store per user:** Activity log entries kept per user.
- **Activity items per page:** Activity log entries shown per page.
- **Resource usage items to log per user:** Resource usage entries stored per user.
- **Resource usage items per page:** Resource usage entries shown per page.

![History and logs section with the login, activity and resource usage history options](/img/openadmin-screenshots/settings/openpanel-history.png#gh-light-mode-only)
![History and logs section with the login, activity and resource usage history options](/img/openadmin-screenshots/settings/openpanel-history_dark.png#gh-dark-mode-only)

## WordPress

Plugins and themes installed on new WordPress sites, and the profiles users can pick from on *OpenPanel > WordPress > Install*. These options are available on all license types and are saved together with the rest of the page.

### Default plugins set

The global default list of WordPress plugins, installed on new WordPress sites when the user keeps **Default** in the *Plugins & themes set* field of the install form (or when there are no [WordPress profiles](#wordpress-profiles)). If the user has their own [plugins set](/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/), theirs is used instead.
Enter one item per row. Supported formats:

* **wp\_org\_slug**: the plugin slug from the WordPress.org plugin page
* **URL**: a direct link to a `.zip` plugin file hosted online

Stored in `/etc/openpanel/wordpress/sets/plugins.txt` file.

### Default themes set

The global default list of WordPress themes, installed on new WordPress sites when the user keeps **Default** in the *Plugins & themes set* field of the install form (or when there are no [WordPress profiles](#wordpress-profiles)). If the user has their own [themes set](/docs/articles/websites/wordpress-plugins-themes-sets-in-openpanel/), theirs is used instead.
Enter one item per row. Supported formats:

* **wp\_org\_slug**: the theme slug from the WordPress.org themes page
* **URL**: a direct link to a `.zip` theme file hosted online

Stored in `/etc/openpanel/wordpress/sets/themes.txt` file.

### WordPress profiles

Named sets of plugins and themes, for example *E-commerce (WooCommerce)*, *Blog* or *Business website*. When at least one profile exists, users get a **Plugins & themes set** field on *OpenPanel > WordPress > Install* where they pick:

- **Default**: the global [plugins set](#default-plugins-set) and [themes set](#default-themes-set) above, or the user's own sets.
- One of your profiles: only that profile's plugins and themes are installed.
- **None, clean WordPress**: no extra plugins or themes are installed.

There are no profiles by default, and until you add one the install form looks the same as before.

To add a profile, open **Add WordPress profile**, enter a name and list the plugins and themes (click **+ Add another profile** to add more in the same save), one per line, in the same formats as the global sets (WordPress.org slug or a URL to a `.zip` file, lines starting with `#` are ignored). Click **Save changes**. Existing profiles are listed above it: open one to rename it or change its plugins and themes and click **Save changes**, or click **Delete** to remove it.

Each profile is stored as a JSON file in `/etc/openpanel/wordpress/sets/profiles/`, for example:

```json
{
  "name": "E-commerce (WooCommerce)",
  "plugins": ["woocommerce", "woocommerce-payments"],
  "themes": ["storefront"]
}
```

The file name (without `.json`) is the profile's ID, which can also be passed as `wp_set` to the [WordPress install API](/docs/panel/api/#install-wordpress) together with `none` for a clean install.

## AI assistant

Settings for the [AI Assistant](/docs/panel/dashboard/dashboard#ai-assistant), the chat users open from the **AI** button in the OpenPanel header. It answers questions about the panel using the OpenPanel documentation, the user's plan features and the page they're on, and can look things up or make changes in the user's account.

To use it, enable the **AI Assistant** module on [Settings > Modules](/docs/admin/settings/modules/), add it to the feature sets of the plans that should have it, and set an API key here.

- **What the assistant can do:**
  - *Answer questions only*: it can't see the user's data, it only explains how to do things.
  - *Answer questions and look up account data* (default): it can list domains, websites, email accounts, usage and similar, but can't change anything.
  - *Answer questions, look up data and make changes*: it can also add, edit and remove things, after the user approves each change.
- **Provider:** OpenRouter, OpenAI, Anthropic (Claude), or Custom for any OpenAI-compatible API such as Ollama.
- **API key:** The key from your provider. Optional for Custom providers that don't need one.
- **Model:** Model name, for example `gpt-5-mini`. Leave empty to use `openrouter/auto` with OpenRouter or `claude-opus-5-5` with Anthropic. Required for OpenAI and Custom.
- **API base URL:** Only for Custom, the URL that serves `/chat/completions`, for example `http://ollama:11434/v1`.
- **Assistant name:** Shown at the top of the chat. Defaults to *AI Assistant*.
- **Requests per user per month:** How many questions each user can ask per month, `0` for unlimited. Default is 100.
- **Max tokens per answer:** Upper limit for the length of each answer. Default is 2048.
- **Extra instructions:** Added to the assistant's instructions, for example your support email or policies.

![AI Assistant section with what the assistant can do, the provider, API key, model, limits and the usage per user this month](/img/openadmin-screenshots/settings/openpanel-ai.png#gh-light-mode-only)
![AI Assistant section with what the assistant can do, the provider, API key, model, limits and the usage per user this month](/img/openadmin-screenshots/settings/openpanel-ai_dark.png#gh-dark-mode-only)

### Actions

The assistant uses the same actions as the [OpenPanel API](/docs/panel/account/api/), but only those for features on the user's plan, and every action goes through the same checks as the API. Logins, account email and password changes, 2FA, passkeys, sessions, MCP tokens, the MySQL root password, autologin links, downloads and uploads are never available to it.

- Looking things up runs right away.
- Changes are shown to the user as a card with the exact values, and run only after they click **Approve**. Users can click **Always allow** to skip the card next time for the same kind of action.
- Actions that delete or overwrite data, like removing a domain, deleting files, restoring backups or changing passwords, always ask and can't be set to **Always allow**.
- After the assistant reads content others can write, like files or logs, every change asks again for the rest of that question, even if the user set it to **Always allow**. This stops text hidden in a file from triggering changes.
- Each question can run at most 12 actions.

Everything the assistant does is written to the user's activity log, for example `used the AI Assistant to run: Add a new domain (POST /api/domains {"domain":"example.com"}), approved by the user`. Passwords and other secrets are hidden.

### Usage

Below the settings, **Usage this month** lists every user who used the assistant this month, with the number of questions, actions, input and output tokens, and when they last used it. Tokens are what your provider bills for, so this shows who uses the assistant most and what it costs you.

The settings are saved in the `[AI]` section of `/etc/openpanel/openpanel/conf/openpanel.config`. OpenPanel needs a restart to use the new values.

Each question sends the matching parts of the documentation along with the conversation to the provider, and every action adds another call, so API usage is billed by your provider. Use **Requests per user per month** to keep costs in check.

## Integrations

- **PageSpeed API key:** Used to fetch data from Google PageSpeed Insights. A warning is shown when there are more than 10 sites on the server but no key is set.
