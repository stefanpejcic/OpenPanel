---
sidebar_position: 2
---

# OpenPanel

Configure nameservers, branding, and UI display settings for the OpenPanel interface.

---

## Branding

Customize the appearance of OpenPanel to match your brand:

- **Brand Name**  

![Branding section with the brand name, logo, favicon and logout URL](/img/openadmin-screenshots/settings/openpanel-branding.png#gh-light-mode-only)
![Branding section with the brand name, logo, favicon and logout URL](/img/openadmin-screenshots/settings/openpanel-branding_dark.png#gh-dark-mode-only)
  Set a custom name to appear in the OpenPanel sidebar and on login pages by entering it in the **"Brand name"** field.

- **Logo**  
  Display a logo instead of the brand name by providing a URL to the image in the **"Logo image"** field.  
  Supported formats: `.png` or `.svg`  
  **Recommended size:** `200px × 36px`

- **Logout URL**  
  Specify the URL users are redirected to after logging out from the panel (typically your main website).

## Nameservers

- ns1
- ns2
- ns3
- ns4

![Nameservers section with the ns1 to ns4 fields](/img/openadmin-screenshots/settings/openpanel-nameservers.png#gh-light-mode-only)
![Nameservers section with the ns1 to ns4 fields](/img/openadmin-screenshots/settings/openpanel-nameservers_dark.png#gh-dark-mode-only)

[Guide on how to properly configure nameservers](/docs/articles/domains/how-to-configure-nameservers-in-openpanel)

If no nameservers are set but the DNS module is enabled, a recommendation is shown to either set at least ns1 and ns2, or [disable the DNS module](/docs/admin/settings/modules) if it won't be used, so BIND9 isn't autostarted and DNS zone files aren't created for every domain.

## Users

Settings for OpenPanel user accounts:

- **Allow Users to Change Username:** If allowed, users can change their username from the *Account > Settings* page.
- **Allow Subdomain Sharing:** If allowed, users can add subdomains for domains that another user owns.
- **Forbidden Usernames:** List of usernames that can not be used, one per line.
- **Restricted Domains:** List of domains that can not be used, one per line.

![Users section with toggles for what OpenPanel users can change](/img/openadmin-screenshots/settings/openpanel-users.png#gh-light-mode-only)
![Users section with toggles for what OpenPanel users can change](/img/openadmin-screenshots/settings/openpanel-users_dark.png#gh-dark-mode-only)

## Display

Additional display settings include:

- **Menu Style:** The default menu for OpenPanel users - **Classic** (expandable groups) or **Modern** (one link per area, with that area's pages as tabs at the top of each page). Users can still switch to the other style for themselves, see [Menu Style](/docs/panel/dashboard/menu-style/). The same setting also switches OpenAdmin's own menu, right after saving.
- **Avatar Type:** Choose between Gravatar, Letter, or Icon for user avatars.
- **Charts Mode for Resource Usage:** Select to show 1 chart, 2 charts, or no charts on the Resource Usage page.
- **Enable Password Reset:** Allow users to reset passwords via the login form (not recommended for security reasons). When enabled without SMTP configured, a warning links to [SMTP settings](/docs/admin/settings/notifications#smtp-settings), since reset emails won't reach users until it's set up.
- **Display 2FA Widget:** Show a message on users' dashboards encouraging them to enable Two-Factor Authentication for enhanced security.
- **Display How-to Guides Widget:** Display helpful how-to articles on users’ dashboard pages. When enabled, a **How-to Articles** field appears below to edit the Knowledge Base articles shown (requires an active Enterprise license and a non-reseller account to edit).
- **Display Link to Report Bugs:** Show a “Found a bug? Let us know” link at the bottom of all user pages for easy bug reporting.
- **Display Country Flag Icons:** Show country flags next to the last login IP in the OpenPanel dashboard.

![Display section with the OpenPanel interface options](/img/openadmin-screenshots/settings/openpanel-display.png#gh-light-mode-only)
![Display section with the OpenPanel interface options](/img/openadmin-screenshots/settings/openpanel-display_dark.png#gh-dark-mode-only)

## Onboarding

- **Enable Onboarding Wizard:** Show a short guided setup wizard to new users on their first login. Steps the user has already set up are skipped (Webserver once they have a domain, Database once they have a database, Backups once a destination is configured, 2FA once it's enabled). See the [Onboarding page](/docs/panel/dashboard/onboarding) for what users see.

The wizard doesn't have its own separate list of steps to configure - each step just mirrors a feature already enabled in [Feature Manager](/docs/admin/plans/feature-manager) (Webserver, PHP, Database, Varnish, Backups, WAF, IP Blocker, 2FA, Passkeys). Disable a feature there for a plan and its onboarding step disappears too for users on that plan.

To reset the wizard for a single user (for testing, or if they ask to see it again), delete their marker file over SSH:

```bash
rm /home/USERNAME/onboarding.completed
```

They'll be offered the wizard again next time they log in, minus any steps they've already set up.

You can also add your own custom steps to the wizard - for example, a step to enable Redis - without modifying OpenPanel itself. See [Add Custom Steps to the Onboarding Wizard](/docs/articles/dev-experience/add-custom-steps-to-onboarding-wizard).

## File Manager

Configure the following settings for the File Manager:

- **Files per Page:** Number of files/folders displayed per page in the File Manager.
- **Max File Size for Viewer:** Maximum file size (in MB) allowed to be opened in the Viewer. Recommended maximum is 20 MB.
- **Max File Size for Editor:** Maximum file size (in MB) allowed to be opened in the Code Editor. Recommended maximum is 10 MB.
- **Max File Size for Upload:** Maximum file size (in MB) allowed to be uploaded via the File Manager. Recommended maximum is 2000 MB.
- **Max File Size for Download:** Maximum file size (in MB) allowed to be downloaded via the File Manager. Recommended maximum is 2000 MB.
- **Auto-Purge Trash After:** Number of days files remain in the user’s trash bin before automatic deletion. Setting to 0 disables auto-purge. Recommended setting is 30 days.
- **Max Time for Compress Process:** Maximum number of minutes before aborting archive compression processes. Recommended to keep this to a few minutes.
- **Max Time for Extract Process:** Maximum number of minutes before aborting archive extraction processes. Recommended to keep this to a few minutes.
- **Max Time for Download Process:** Maximum number of minutes before aborting file download processes. Recommended to keep this to a few minutes.
- **Enable View and Edit Options for (textual) Extensions:** Specify file extensions allowed to be opened and edited in the File Manager (should be textual file types).
- **Enable View Option for (base64 image) Extensions:** Specify image file extensions that can be displayed using base64 encoding in the Viewer.
- **Enable Extract and Archive Options for (archives) Extensions:** Specify archive file extensions that can be extracted using the File Manager.

A warning is shown when:

- a max file size for the Viewer, Editor, Upload or Download is bigger than the RAM limit of the OpenPanel container (set on [Service Limits](/docs/admin/services/limits)), since files that big can fail.
- Auto-Purge Trash is set to 0, since deleted files then pile up in users' Trash folders until they empty it themselves.
- Auto-Purge Trash is over 30 days, recommending to change the `opencli files-purge_trash` schedule to weekly on [Scheduled Actions](/docs/admin/advanced/crons).

![File Manager section with the file manager options](/img/openadmin-screenshots/settings/openpanel-filemanager.png#gh-light-mode-only)
![File Manager section with the file manager options](/img/openadmin-screenshots/settings/openpanel-filemanager_dark.png#gh-dark-mode-only)

## Databases

Settings for user's databases:

- **Max Startup Time for MySQL:** Maximum number of seconds to wait for MySQL to initialize before timing out.
- **Max File Size for MySQL Import:** Maximum file size (in GB) allowed for a MySQL import.
- **Restricted MySQL (system) Users:** List of MySQL usernames that users are not allowed to access/manage.
- **Restricted MySQL (system) Databases:** List of MySQL databases that users are not allowed to access/manage.

A warning is shown when:

- Max Startup Time for MySQL is under 10 seconds, since MySQL needs time to initialize on first start. If a custom `MYSQL_VERSION` is set for new users, at least 60 seconds is recommended, since the first start downloads that image from Docker Hub.
- Max File Size for MySQL Import is bigger than the RAM limit of the OpenPanel container.

![Databases section with the database options](/img/openadmin-screenshots/settings/openpanel-databases.png#gh-light-mode-only)
![Databases section with the database options](/img/openadmin-screenshots/settings/openpanel-databases_dark.png#gh-dark-mode-only)

## Security

Captcha, 2FA, and password protections for login and signup forms:

- **Captcha Provider:** Choose between Disabled, Google reCAPTCHA, Cloudflare Turnstile, or Custom.
- Depending on the provider selected, enter the corresponding **site key** and **secret key** (or custom site key for the Custom provider).

![Security section with the OpenPanel login and session options](/img/openadmin-screenshots/settings/openpanel-security.png#gh-light-mode-only)
![Security section with the OpenPanel login and session options](/img/openadmin-screenshots/settings/openpanel-security_dark.png#gh-dark-mode-only)

:::info
Captcha requires the [captcha plugin](https://github.com/stefanpejcic/captcha/) to be installed — these settings have no effect until it is installed.
:::

- **Minimum Password Strength:** A value between 1-100 setting the minimum password strength required for all password input fields in OpenPanel (account, FTP, Emails, Databases, etc).
- **Check Passwords Against Weakpass.com:** When enabled, user passwords during account creation and reset are verified against Weakpass.com’s list of compromised passwords.
- **Enforce 2FA:** On login redirects users to 2FA setup page and prevents them from accessing other pages until 2FA is configured.
- **Validate IP Address on Session Cookie:** When enabled, a session cookie is rejected if the request's IP address doesn't match the IP it was issued to.

A warning is shown when Minimum Password Strength is below 50 (weak passwords are accepted), when the Weakpass.com check is off (passwords found in data breaches can be used), and when Validate IP Address on Session Cookie is off (a cookie stolen from a user, for example through phishing, can be reused from any IP).

## Statistics

Configure the following settings related to user login attempts, session management, and data retention:

- **Failed Logins per Minute Before User is Rate-Limited:** Number of failed login attempts allowed per minute from a single IP before that IP is temporarily rate-limited to prevent brute-force attacks.
- **Failed Logins per Minute Before User is Blocked for 1 Hour:** Threshold of failed login attempts per minute that triggers a 1-hour block for the offending IP address.
- **Session Duration (in Minutes):** Length of time a user session remains active before requiring re-authentication.
- **Session Lifetime (in Minutes):** Total maximum lifetime of a session, after which the user will be logged out regardless of activity.
- **Login Records to Keep per User:** Number of recent login attempts stored per user for audit and security tracking purposes.
- **Activity Records to Store per User:** Maximum number of user activity logs retained for reviewing past actions within the panel.
- **Activity Items per Page:** Number of activity log entries displayed per page in the user interface.
- **Resource Usage Items to Display per Page:** Number of resource usage records shown per page when viewing user or system resource statistics.
- **Resource Usage Items to Log per User:** Number of resource usage entries recorded and stored per user for historical analysis.
- **Domains per Page:** Number of domain entries displayed per page in domain management lists.
- **Terminal Commands Timeout (in Seconds):** Maximum number of seconds a Docker Terminal command can run before being timed out.
- **PageSpeed API Key:** If set, this API key is used to fetch data from Google PageSpeed Insights.

A warning is shown when Terminal Commands Timeout is under 30 seconds (it applies to both the OpenPanel web terminal and OpenAdmin), and when there are more than 10 sites on the server but no PageSpeed API key.

![Statistics section with the resource usage and statistics options](/img/openadmin-screenshots/settings/openpanel-statistics.png#gh-light-mode-only)
![Statistics section with the resource usage and statistics options](/img/openadmin-screenshots/settings/openpanel-statistics_dark.png#gh-dark-mode-only)

## AI Assistant

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
