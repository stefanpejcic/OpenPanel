---
sidebar_position: 2
---

# OpenPanel

Configure nameservers, branding, and UI display settings for the OpenPanel interface.

---

## Branding

Customize the appearance of OpenPanel to match your brand:

- **Brand Name**  
  Set a custom name to appear in the OpenPanel sidebar and on login pages by entering it in the **"Brand name"** field.

- **Logo**  
  Display a logo instead of the brand name by providing a URL to the image in the **"Logo image"** field.  
  Supported formats: `.png` or `.svg`  
  **Recommended size:** `200px × 36px`

- **Logout URL**  
  Specify the URL users are redirected to after logging out from the panel (typically your main website).

## Nameservers

Before adding any domains, it is important to first create nameservers to ensure valid DNS zone files and proper propagation.

Configuring Nameservers involves two key steps:

1. **Create private nameservers (glue DNS records)** for your domain through your domain registrar.
2. **Add the nameservers** to the OpenPanel configuration.

Tutorials for Popular Domain Providers

- [Cloudflare](https://developers.cloudflare.com/dns/additional-options/custom-nameservers/zone-custom-nameservers/)  
- [GoDaddy](https://uk.godaddy.com/help/add-custom-hostnames-12320)  
- [NameCheap](https://www.namecheap.com/support/knowledgebase/article.aspx/768/10/how-do-i-register-personal-nameservers-for-my-domain/#:~:text=Click%20on%20the%20Manage%20option,5.)

To add nameservers in OpenPanel, navigate to **OpenAdmin > Settings > OpenPanel** and enter your nameservers in the `ns1` and `ns2` fields, then click **Save changes**.

Alternatively, you can set nameservers via terminal commands:

```bash
opencli config update ns1 your_ns1.domain.com
opencli config update ns2 your_ns2.domain.com
```

Currently, you can add up to 4 nameservers.

:::info
After creating nameservers it can take up to 12h for the records to be globally accessible. Use a tool such as [whatsmydns.net](https://www.whatsmydns.net/) to monitor the status.

If you still experience problems after the propagation process, then please check this guide: [dns server not responding to requests](https://community.openpanel.co/d/5-dns-server-does-not-respond-to-request-for-domain-zone).
:::


## Display

Additional display settings include:

- **Avatar Type:** Choose between Gravatar, Letter, or Icon for user avatars.
- **Charts Mode for Resource Usage:** Select to show 1 chart, 2 charts, or no charts on the Resource Usage page.
- **Check Passwords Against Weakpass.com:** When enabled, user passwords during account creation and reset are verified against Weakpass.com’s list of compromised passwords.
- **Enable Password Reset:** Allow users to reset passwords via the login form (not recommended for security reasons).
- **Display 2FA Widget:** Show a message on users' dashboards encouraging them to enable Two-Factor Authentication for enhanced security.
- **Display How-to Guides Widget:** Display helpful how-to articles on users’ dashboard pages.
- **Display Link to Report Bugs:** Show a “Found a bug? Let us know” link at the bottom of all user pages for easy bug reporting.
- **Display Country Flag Icons:** Show country flags next to the last login IP in the OpenPanel dashboard.

## File Manager

Configure the following settings for the File Manager:

- **Max File Size for Viewer:** Maximum file size (in MB) allowed to be opened in the Viewer. Recommended maximum is 20 MB.
- **Max File Size for Editor:** Maximum file size (in MB) allowed to be opened in the Code Editor. Recommended maximum is 10 MB.
- **Max File Size for Upload:** Maximum file size (in MB) allowed to be uploaded via the File Manager. Recommended maximum is 2000 MB.
- **Max File Size for Download:** Maximum file size (in MB) allowed to be downloaded via the File Manager. Recommended maximum is 2000 MB.
- **Auto-Purge Trash After:** Number of days files remain in the user’s trash bin before automatic deletion. Setting to 0 disables auto-purge. Recommended setting is 30 days.
- **Max Time for Compress Process:** Maximum number of minutes before aborting archive compression processes. Recommended to keep this to a few minutes.
- **Max Time for Extract Process:** Maximum number of minutes before aborting archive extraction processes. Recommended to keep this to a few minutes.
- **Enable View and Edit Options for (textual) Extensions:** Specify file extensions allowed to be opened and edited in the File Manager (should be textual file types).
- **Enable View Option for (base64 image) Extensions:** Specify image file extensions that can be displayed using base64 encoding in the Viewer.
- **Enable Extract and Archive Options for (archives) Extensions:** Specify archive file extensions that can be extracted using the File Manager.

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
