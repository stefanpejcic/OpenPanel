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
- **Enable Password Reset:** Allow users to reset passwords via the login form (not recommended for security reasons).
- **Display 2FA Widget:** Show a message on users' dashboards encouraging them to enable Two-Factor Authentication for enhanced security.
- **Display How-to Guides Widget:** Display helpful how-to articles on users’ dashboard pages. When enabled, a **How-to Articles** field appears below to edit the Knowledge Base articles shown (requires an active Enterprise license and a non-reseller account to edit).
- **Display Link to Report Bugs:** Show a “Found a bug? Let us know” link at the bottom of all user pages for easy bug reporting.
- **Display Country Flag Icons:** Show country flags next to the last login IP in the OpenPanel dashboard.

![Display section with the OpenPanel interface options](/img/openadmin-screenshots/settings/openpanel-display.png#gh-light-mode-only)
![Display section with the OpenPanel interface options](/img/openadmin-screenshots/settings/openpanel-display_dark.png#gh-dark-mode-only)

## Onboarding

- **Enable Onboarding Wizard:** Show a short guided setup wizard to new users on their first login, before they've added a domain or a database. See the [Onboarding page](/docs/panel/dashboard/onboarding) for what users see.

The wizard doesn't have its own separate list of steps to configure - each step just mirrors a feature already enabled in [Feature Manager](/docs/admin/plans/feature-manager) (Webserver, PHP, Database, Varnish, Backups, WAF, IP Blocker, 2FA, Passkeys). Disable a feature there for a plan and its onboarding step disappears too for users on that plan.

To reset the wizard for a single user (for testing, or if they ask to see it again), delete their marker file over SSH:

```bash
rm /home/USERNAME/onboarding.completed
```

They'll be offered the wizard again next time they log in, as long as they still have no domains or databases.

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

![File Manager section with the file manager options](/img/openadmin-screenshots/settings/openpanel-filemanager.png#gh-light-mode-only)
![File Manager section with the file manager options](/img/openadmin-screenshots/settings/openpanel-filemanager_dark.png#gh-dark-mode-only)

## Databases

Settings for user's databases:

- **Max Startup Time for MySQL:** Maximum number of seconds to wait for MySQL to initialize before timing out.
- **Max File Size for MySQL Import:** Maximum file size (in GB) allowed for a MySQL import.
- **Restricted MySQL (system) Users:** List of MySQL usernames that users are not allowed to access/manage.
- **Restricted MySQL (system) Databases:** List of MySQL databases that users are not allowed to access/manage.

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

![Statistics section with the resource usage and statistics options](/img/openadmin-screenshots/settings/openpanel-statistics.png#gh-light-mode-only)
![Statistics section with the resource usage and statistics options](/img/openadmin-screenshots/settings/openpanel-statistics_dark.png#gh-dark-mode-only)
