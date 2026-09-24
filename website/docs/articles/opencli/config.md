# Config

`opencli config` allows you to change the settings for the user interface.


## Get
The `get` parameter allows you to view current settings.
```bash
opencli config get <OPTION>
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get api
off
```
</details>

## Update
The `update` parameter allows you to change the settings.
```bash
opencli config update <OPTION> <NEW-VALUE>
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update api on
Updated api to on
```
</details>

:::info
To apply the new setting, the OpenPanel UI container is restarted, except for these options: `email`, `autoupdate`, `autopatch`, `key`, `max_cpu`, `max_ram`.
:::

## Available Options
Currently available configuration options:

### logo
`logo` allows you to set a url for image *(suggested dimensions are 80x200px) that will be displayed to users on:

- every page
- login page
- notification emails footer

```bash
logo=https://http.cat/images/100.jpg
```

### `favicon`
URL of a favicon to be displayed on all OpenPanel pages.

**Check current value**:
```bash
opencli config get favicon
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get favicon
Parameter favicon has no value.
```
</details>

**Set new value**:
```bash
opencli config update favicon https://example.com/favicon.ico
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update favicon https://example.com/favicon.ico
Updated favicon to https://example.com/favicon.ico
```
</details>

### `dir`
Text direction of the OpenPanel interface: `ltr` (left-to-right) or `rtl` (right-to-left). Default is `ltr`.

**Check current value**:
```bash
opencli config get dir
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get dir
ltr
```
</details>

**Set new value**:
```bash
opencli config update dir rtl
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update dir rtl
Updated dir to rtl
```
</details>

### brand_name

`brand_name` allows you to brand the OpenPanel name that users see in their user panel with your custom brand name.

This brand name is displayed in the following positions:
- on top left when logo is not provided
- in the page title
- in the footer text on every page
  

```bash
brand_name=pcx3.com
```

### max_login_records

`max_login_records` determines how many records are kept for every users login history page. Default value, if empty is 20.

```bash
max_login_records=20
```

### default_php_version

The default PHP version for new users is no longer set in `openpanel.config`. It is set with `DEFAULT_PHP_VERSION` in the `.env` template for new users: `/etc/openpanel/docker/compose/1.0/.env`. To change it for an existing user, use [`opencli php-default`](/docs/articles/opencli/php/#change-the-default-version).

### domains_per_page

Set number of domains to show per page for the user, default value is 100.

```bash
domains_per_page=100
```

### domain_log_per_page

Set number of lines (table rows) to display per page on *Domains > Access Logs* and *Advanced > WAF Logs*. Default value is 500.

```bash
domain_log_per_page=500
```

### domain_log_max_for_show_all

Set maximum number of lines (table rows) to display **when user selects *show_all* option** on *Domains > Access Logs* and *Advanced > WAF Logs*. Default value is 10000.

```bash
domain_log_max_for_show_all=10000
```

### acccess_logs_per_page

Set number of lines to show per page from the domains activity log.

```bash
acccess_logs_per_page=1000
```

### resource_usage_items_per_page

Set number of rows to display per page on the 'Resource Usage History' page.

```bash
resource_usage_items_per_page=100
```

### resource_usage_retention

Set number of records to keep for each user, default value is 100.

```bash
resource_usage_retention=100
```

### resource_usage_charts_mode
`resource_usage_charts_mode` allows you to set the number of charts displayed on users _Historical Resource Usage_ page.


- _one_ - displays a single chart for both CPU and RAM usage.
- _two_ (Default) - displays two charts: one for CPU and another for RAM. 
- _none_ - displays no charts.

```bash
resource_usage_charts_mode=two
```

### found_a_bug_link
When set to 'yes', a *Found a bug? Let us know* link appears in the bottom-right corner of all pages in the OpenPanel and OpenAdmin UI, allowing users to report issues directly on GitHub.

```bash
found_a_bug_link=yes
```

### ip_county_flag
When set to 'yes', a *Cuntry Code Flag* icon appears in the Dashboard page for last login IP address.
```bash
ip_county_flag=yes
```

### login_ratelimit
The `login_ratelimit` parameter allows you to specify the maximum number of failed login attempts per minute for each IP address accessing the OpenPanel interface. The default setting is `5`.

```bash
login_ratelimit=5
```

### login_blocklimit

The `login_blocklimit` parameter allows you to specify the maximum number of failed login attempts per IP before it is temporary blocked for one hour. The default setting is `20`.

```bash
login_blocklimit=20
```


### session_lifetime

The `session_lifetime` parameter allows you to specify the number of minutes that user session is valid after they login. If there are no user interactions during this period then the session espires. The default setting is `300`.

```bash
session_lifetime=300
```


### session_duration

The `session_duration` parameter allows you to specify the number of minutes that session is extended for after the `session_lifetime` setting based on user interactions. The default setting is `30`.

```bash
session_duration=30
```

### terminal_timeout

`terminal_timeout` enables you to set time in seconds to limit command execution from *OpenPanel > Containers > Terminal* and *OpenAdmin > Server > Terminal*.

```bash
terminal_timeout=30
```

### weakpass

`weakpass` enables you to check passwords against the [weakpass.com](https://weakpass.com/wordlist) dictionaries and prevent the use of weak passwords. By default, it is set to 'yes,' allowing weak passwords to be used.

```bash
weakpass=yes
```

### password_reset

`password_reset` allows you to enable or disable the password reset functionality for the login page. By default password reset is disabled.

```bash
password_reset=yes
```


### permit_username_change_by_user

`permit_username_change_by_user` allows Administrators to enable or disable the username change functionality by OpenPanel users from their Account page. By default username change is disabled.

```bash
permit_username_change_by_user=no
```

### avatar_type

`avatar_type` allows you to set the type of images used for users profile pictures in the users panel.


Available options:

- _gravatar_ - displays the gravatar picture associated with the email address of the user.
- _letter_ - displays a CSS generated icon using the users first letter of the username.
- _icon_ - displays a placeholder bootstrap icon representing the user.

```bash
avatar_type=gravatar
```

### logout_url

`logout_url` allows administrators to define the destination to which users should be redirected after they log out. The default setting directs them to the `/login` page.

```bash
logout_url=https://google.com
```

### activity_items_per_page

`activity_items_per_page` allows administrators to define the default number of displayed items on the `/activity` page, the value defaults to 25.

```bash
activity_items_per_page=50
```

### activity_lines_retention

`activity_lines_retention` allows administrators to define the default number of lines to keep for each user `/activity` log. The value defaults to 1500.

```bash
activity_lines_retention=1500
```


### activity_max_size_bytes

`activity_max_size_bytes` allows administrators to define the total size in bytes for each user `/activity` log. The value defaults to 2000000 (2MB).

```bash
activity_max_size_bytes=2000000
```



### nameservers

The `ns1` `ns2` `ns3` `ns4` options allow you to set nameservers that will be used in dns zone files for newly added domains, and displayed to users on their panel dashboard.
Nameservers should be added in pairs, ns1 and ns2, ns3 and ns4.

```bash
ns1=ns1.openpanel.com
ns2=ns2.openpanel.com
ns3=ns3.openpanel.com
ns4=ns4.openpanel.com
```

### enabled_modules

In OpenPanel version 0.1.5, we implemented the Modules feature, which, by default, loads only the essential modules. Administrators also have the option to selectively disable modules they do not require.


### `permit_subdomain_sharing`
If set to `yes`, users can add subdomains of domains that another user owns. Default is `no`.

**Check current value**:
```bash
opencli config get permit_subdomain_sharing
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get permit_subdomain_sharing
no
```
</details>

**Set new value**:
```bash
opencli config update permit_subdomain_sharing yes
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update permit_subdomain_sharing yes
Updated permit_subdomain_sharing to yes
```
</details>

### `email`
Administrator email address used for notifications. Also set as `LF_ALERT_TO` in the CSF configuration.

**Check current value**:
```bash
opencli config get email
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get email
Parameter email has no value.
```
</details>

**Set new value**:
```bash
opencli config update email admin@example.com
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update email admin@example.com
Updated email to admin@example.com
```
</details>

### `password_strength`
Minimum password strength (`1`-`100`) for all password fields in OpenPanel: account, FTP, emails, databases. Default is `50`.

**Check current value**:
```bash
opencli config get password_strength
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get password_strength
50
```
</details>

**Set new value**:
```bash
opencli config update password_strength 70
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update password_strength 70
Updated password_strength to 70
```
</details>

### `email_plaintext_passwords`
If set to `yes`, the welcome email sent with `opencli user-add --send-email` includes the password in plain text. If `no`, the password is replaced with `********`. Default is `no`.

**Check current value**:
```bash
opencli config get email_plaintext_passwords
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get email_plaintext_passwords
no
```
</details>

**Set new value**:
```bash
opencli config update email_plaintext_passwords yes
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update email_plaintext_passwords yes
Updated email_plaintext_passwords to yes
```
</details>

### `validate_ip_address_cookie`
When enabled (`yes`), a session cookie is rejected if the request's IP address doesn't match the IP address it was issued to. Default is `yes`.

**Check current value**:
```bash
opencli config get validate_ip_address_cookie
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get validate_ip_address_cookie
yes
```
</details>

**Set new value**:
```bash
opencli config update validate_ip_address_cookie no
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update validate_ip_address_cookie no
Updated validate_ip_address_cookie to no
```
</details>

### `key`
The OpenPanel Enterprise license key. Use [`opencli license`](/docs/articles/opencli/license/) to add, verify or delete the key instead of setting it directly.

**Check current value**:
```bash
opencli config get key
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get key
Parameter key has no value.
```
</details>

**Set new value**:
```bash
opencli config update key enterprise-XXXXXXXXXX
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update key enterprise-XXXXXXXXXX
Updated key to enterprise-XXXXXXXXXX
```
</details>

### Captcha

Captcha for the OpenPanel login and signup forms.

### `captcha_provider`
Captcha provider to use: `none` (disabled), `google` (Google reCAPTCHA), `turnstile` (Cloudflare Turnstile) or `custom`. Default is empty (disabled).

**Check current value**:
```bash
opencli config get captcha_provider
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get captcha_provider
Parameter captcha_provider has no value.
```
</details>

**Set new value**:
```bash
opencli config update captcha_provider turnstile
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update captcha_provider turnstile
Updated captcha_provider to turnstile
```
</details>

### `recaptcha_site_key`
Site key for Google reCAPTCHA (used when `captcha_provider` is `google`).

**Check current value**:
```bash
opencli config get recaptcha_site_key
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get recaptcha_site_key
Parameter recaptcha_site_key has no value.
```
</details>

**Set new value**:
```bash
opencli config update recaptcha_site_key <SITE_KEY>
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update recaptcha_site_key <SITE_KEY>
Updated recaptcha_site_key to <SITE_KEY>
```
</details>

### `recaptcha_secret_key`
Secret key for Google reCAPTCHA (used when `captcha_provider` is `google`).

**Check current value**:
```bash
opencli config get recaptcha_secret_key
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get recaptcha_secret_key
Parameter recaptcha_secret_key has no value.
```
</details>

**Set new value**:
```bash
opencli config update recaptcha_secret_key <SECRET_KEY>
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update recaptcha_secret_key <SECRET_KEY>
Updated recaptcha_secret_key to <SECRET_KEY>
```
</details>

### `turnstile_site_key`
Site key for Cloudflare Turnstile (used when `captcha_provider` is `turnstile`).

**Check current value**:
```bash
opencli config get turnstile_site_key
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get turnstile_site_key
Parameter turnstile_site_key has no value.
```
</details>

**Set new value**:
```bash
opencli config update turnstile_site_key <SITE_KEY>
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update turnstile_site_key <SITE_KEY>
Updated turnstile_site_key to <SITE_KEY>
```
</details>

### `turnstile_secret_key`
Secret key for Cloudflare Turnstile (used when `captcha_provider` is `turnstile`).

**Check current value**:
```bash
opencli config get turnstile_secret_key
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get turnstile_secret_key
Parameter turnstile_secret_key has no value.
```
</details>

**Set new value**:
```bash
opencli config update turnstile_secret_key <SECRET_KEY>
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update turnstile_secret_key <SECRET_KEY>
Updated turnstile_secret_key to <SECRET_KEY>
```
</details>

### `custom_captcha_site_key`
Site key for a custom captcha (used when `captcha_provider` is `custom`).

**Check current value**:
```bash
opencli config get custom_captcha_site_key
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get custom_captcha_site_key
Parameter custom_captcha_site_key has no value.
```
</details>

**Set new value**:
```bash
opencli config update custom_captcha_site_key <SITE_KEY>
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update custom_captcha_site_key <SITE_KEY>
Updated custom_captcha_site_key to <SITE_KEY>
```
</details>

### logrotate_enable
Enable or disable **logrotate** for system logs: webserver, DNS, OpenPanel, FTP, MailServer, etc.
```bash
logrotate_enable=yes
```

* **yes** → Log rotation is active for supported services.
* **no** → Logs will grow indefinitely unless manually managed.

### `logrotate_size_limit`
Set the maximum log file size (in MB) before rotation occurs. Once a log file exceeds this size, it will be rotated automatically.

```bash
logrotate_size_limit=100
```

### `logrotate_retention`
Define the number of rotated log files to keep. Once the limit is reached, the oldest files are deleted.
```bash
logrotate_retention=10
```

* Example: If set to `10`, the system will keep 10 old log archives per log file.
* Helps control disk usage.

---

### `logrotate_keep_days`
Specify how many days rotated logs should be retained before deletion.
```bash
logrotate_keep_days=30
```

* Example: Setting to `30` means rotated logs older than 30 days will be removed.
* Works alongside `logrotate_retention` — whichever limit is hit first will trigger deletion.


### `goaccess_enable`

Enable or disable **GoAccess** web log analytics.

```bash
goaccess_enable=yes
```

* **yes** → GoAccess will process web server logs to generate analytics reports.
* **no** → Analytics will be disabled.

### `goaccess_schedule`

Set how often GoAccess reports should be generated.

```bash
goaccess_schedule=monthly
```

* Supported values: `daily`, `weekly`, `monthly`.
* Determines the frequency of log analysis and report generation.


### `goaccess_email`

Control whether GoAccess reports are emailed automatically.

```bash
goaccess_email=no
```

* **yes** → Reports are sent via email to the configured recipient.
* **no** → Reports are generated but not emailed.


### `goaccess_keep_days`

Specify how long GoAccess reports are retained before deletion.

```bash
goaccess_keep_days=365
```

* Example: `365` means reports older than one year will be removed automatically.
* Helps manage disk space while keeping historical analytics.



### autopatch

The `autopatch` option allows Administrator to automatically update OpenPanel to minor versions.

| Autopatch    | Update to minor version  | Update to major version  |
| -------- | ------- | ------- |
| yes | 1.0.2 will be updated to 1.0.3 | 1.0.2 will **NOT** be updated to 2.0.0 |
| no | 1.0.2 will **NOT** be updated to 1.0.3 | 1.0.2 will **NOT** be updated to 2.0.0 |

```bash
autopatch=yes
```

### autoupdate

The `autoupdate` option allows Aministrator to enable or disable automatic updates.

| Autoupdate    | Update to minor version  | Update to major version  |
| -------- | ------- | ------- |
| yes | 1.0.2 will **NOT** be updated to 1.0.3 | 1.0.2 will be updated to 2.0.0 |
| no | 1.0.2 will **NOT** be updated to 1.0.3 | 1.0.2 will **NOT** be updated to 2.0.0 |

```bash
autoupdate=yes
```


### twofa_nag

The `twofa_nag` option allows Administrator to set if 2FA nag should be displayed to users on their dashboard page.


```bash
twofa_nag=yes
```

### `twofa_enforce`
When enabled (`yes`), users have to set up 2FA before they can access any other page in OpenPanel. The 2FA module must be enabled. Default is `no`.

**Check current value**:
```bash
opencli config get twofa_enforce
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get twofa_enforce
no
```
</details>

**Set new value**:
```bash
opencli config update twofa_enforce yes
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update twofa_enforce yes
Updated twofa_enforce to yes
```
</details>

### `onboarding`
When enabled (`yes`), users who haven't finished it yet see a first-login setup wizard on their Dashboard: webserver, default PHP version, database server, backup destination and 2FA/passkey setup. Default is `yes`.

**Check current value**:
```bash
opencli config get onboarding
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get onboarding
yes
```
</details>

**Set new value**:
```bash
opencli config update onboarding no
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update onboarding no
Updated onboarding to no
```
</details>

### how_to_guides

The `how_to_guides` option allows Aministrator to enable or disable the How-to guides section from users dashboards.


```bash
how_to_guides=yes
```
If configured as "yes", the system will initially attempt to access a JSON file containing your custom how-to guides. In the event that the file is not available, articles from https://openpanel.com/docs/panel/dashboard/#how-to-guides will be shown instead.


### api

The `api` option allows the **Administrator** to enable or disable API functionality. The default value is `off`.

* **on** → API access is enabled.
* **off** → API access is disabled.

**Check current status**:
```bash
opencli config get how_to_guides
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get how_to_guides
yes
```
</details>

**Enable API**:
```bash
opencli config update how_to_guides on
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update how_to_guides on
Updated how_to_guides to on
```
</details>

**Disable API (default)**:
```bash
opencli config update how_to_guides off
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update how_to_guides off
Updated how_to_guides to off
```
</details>

### dev_mode

The `dev_mode` option allows Aministrator to view more verbose information in both OpenPanel and OpenAdmin, as well as their log files.

- displays detailed startup log
- verbose logging for API requests
- disables minified responses for both interfaces (HTML, JS, CSS and JSON)
- uses latest CDN for tailwindcss and alpinejs
- displays detailed processing of each request

* **on** → DEV_MODE is enabled.
* **off** → DEV_MODE is disabled.

**Check current status**:
```bash
opencli config get dev_mode
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get dev_mode
off
```
</details>

**Enable dev_mode**:
```bash
opencli config update dev_mode on
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update dev_mode on
Updated dev_mode to on
```
</details>

**Disable dev_mode (default)**:
```bash
opencli config update dev_mode off
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update dev_mode off
Updated dev_mode to off
```
</details>


### favicons

The `favicons` option allows you to specify api to use for favicon images.

To set api:
```bash
opencli config update favicons https://www.google.com/s2/favicons?domain=
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update favicons https://www.google.com/s2/favicons?domain=
Updated favicons to https://www.google.com/s2/favicons?domain=
```
</details>

```bash
opencli config update favicons https://favicon-api.openpanel.org/
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update favicons https://favicon-api.openpanel.org/
Updated favicons to https://favicon-api.openpanel.org/
```
</details>

For selfhosting favicon api, check [vemetric/favicon-api](https://github.com/Vemetric/favicon-api?tab=readme-ov-file#deployment)

### favorites_items

The `favorites_items` option controls total number of favorites items that user can have at a time. Default is 10.

To change value:

```bash
opencli config update favorites_items 5
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update favorites_items 5
Updated favorites_items to 5
```
</details>

### demo_mode

The `demo_mode` option allows Aministrator lock both OpenPanel and OpenAdmin UI in read-only mode, allowing users to login and view data, but not to make any changes.

* **on** → DEMO_MODE is enabled.
* **off** → DEMO_MODE is disabled.

**Check current status**:
```bash
opencli config get dev_mode
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get dev_mode
off
```
</details>

**Enable demo_mode**:
```bash
opencli config update dev_mode on
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update dev_mode on
Updated dev_mode to on
```
</details>

**Disable demo_mode (default)**:
```bash
opencli config update demo_mode off
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update demo_mode off
Updated demo_mode to off
```
</details>

### max_ram

The `max_ram` option allows Aministrator to set maximum % of available RAM that can be used by Docker service.

**Check current value**:
```bash
opencli config get max_ram
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get max_ram
90
```
</details>

**Set new value**:
```bash
opencli config update max_ram 85
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update max_ram 85
Updated max_ram to 85
```
</details>

### max_cpu

The `max_cpu` option allows Aministrator to set maximum % of available CPU that can be used by Docker service.

**Check current value**:
```bash
opencli config get max_cpu
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get max_cpu
95
```
</details>

**Set new value**:
```bash
opencli config update max_cpu 95
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update max_cpu 95
Updated max_cpu to 95
```
</details>

### screenshots
The `screenshots` option allows Aministrator to set a custom screenshots api to be used for generating website screenshots. 

- If 'local' is defined then screenshots will be generated locally on the server.
- If left empty, default is our remote api
- To define custom screenshots api add it's url as value.

```bash
screenshots=http://screenshots-api.openpanel.com/screenshot
```


### temporary_links

The `temporary_links` option allows Aministrator to set a custom domain name to be used to generate temporary sub-domains for WP installations.

```bash
temporary_links=https://preview.openpanel.org/index.php
```

### blacklist_useragents

The `blacklist_useragents` option allows Administrators to blacklist certain useragents from accessing OpenPanel UI.

**Check current value**:
```bash
opencli config get blacklist_useragents
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get blacklist_useragents
no
```
</details>

**Enable blacklist**:
```bash
opencli config update blacklist_useragents yes
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update blacklist_useragents yes
Updated blacklist_useragents to yes
```
</details>

**Disable blacklist**:
```bash
opencli config update blacklist_useragents no
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update blacklist_useragents no
Updated blacklist_useragents to no
```
</details>

### `mail_server`
Hostname or IP address of the SMTP mail server to be used for sending notifications to users.

**Check current value**:
```bash
opencli config get mail_server
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mail_server
Parameter mail_server has no value.
```
</details>

**Set new value**:
```bash
opencli config update mail_server smtp.example.com
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mail_server smtp.example.com
Updated mail_server to smtp.example.com
```
</details>


### `mail_port`
SMTP server port (common: 25, 465, 587).

**Check current value**:
```bash
opencli config get mail_port
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mail_port
465
```
</details>

**Set new value**:
```bash
opencli config update mail_port 465
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mail_port 465
Updated mail_port to 465
```
</details>


### `mail_use_tls`
Enable or disable STARTTLS encryption.

**Check current value**:
```bash
opencli config get mail_use_tls
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mail_use_tls
False
```
</details>

**Set new value**:
```bash
opencli config update mail_use_tls True
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mail_use_tls True
Updated mail_use_tls to True
```
</details>


### `mail_use_ssl`
Enable or disable SSL/TLS for SMTP connections.

**Check current value**:
```bash
opencli config get mail_use_ssl
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mail_use_ssl
True
```
</details>

**Set new value**:
```bash
opencli config update mail_use_ssl True
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mail_use_ssl True
Updated mail_use_ssl to True
```
</details>


### `mail_username`
SMTP username for authentication.

**Check current value**:
```bash
opencli config get mail_username
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mail_username
Parameter mail_username has no value.
```
</details>

**Set new value**:
```bash
opencli config update mail_username user@example.com
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mail_username user@example.com
Updated mail_username to user@example.com
```
</details>


### `mail_password`
SMTP password for authentication.

**Check current value**:
```bash
opencli config get mail_password
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mail_password
Parameter mail_password has no value.
```
</details>

**Set new value**:
```bash
opencli config update mail_password yourpassword
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mail_password yourpassword
Updated mail_password to yourpassword
```
</details>


### `mail_default_sender`
Default “From” address for outgoing emails.

**Check current value**:
```bash
opencli config get mail_default_sender
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mail_default_sender
Parameter mail_default_sender has no value.
```
</details>

**Set new value**:
```bash
opencli config update mail_default_sender no-reply@example.com
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mail_default_sender no-reply@example.com
Updated mail_default_sender to no-reply@example.com
```
</details>


### `mail_security_token`
Optional token for email API or internal security.

**Check current value**:
```bash
opencli config get mail_security_token
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mail_security_token
78bd237db2da
```
</details>

**Set new value**:
```bash
opencli config update mail_security_token 78bd237db2da
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mail_security_token 78bd237db2da
Updated mail_security_token to 78bd237db2da
```
</details>


### `mail_debug`
Enable debug logging for email operations (each email sent from the OpenAdmin is stored as file on the server).

**Check current value**:
```bash
opencli config get mail_debug
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mail_debug
False
```
</details>

**Set new value**:
```bash
opencli config update mail_debug True
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mail_debug True
Updated mail_debug to True
```
</details>




### `autopurge_trash`
Automatically purge files in the File Manager trash after a number of days.

**Check current value**:
```bash
opencli config get autopurge_trash
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get autopurge_trash
7
```
</details>

**Set new value**:
```bash
opencli config update autopurge_trash 7
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update autopurge_trash 7
Updated autopurge_trash to 7
```
</details>


### `filemanager_files_per_page`
Number of files displayed per page in the File Manager. Default is `1000`.

**Check current value**:
```bash
opencli config get filemanager_files_per_page
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_files_per_page
1000
```
</details>

**Set new value**:
```bash
opencli config update filemanager_files_per_page 500
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_files_per_page 500
Updated filemanager_files_per_page to 500
```
</details>

### `filemanager_buttons_style`

Set default buttons style on 'OpenPanel > File Manager' - Available 'classic' 'modern'

**Check current value**:
```bash
opencli config get filemanager_buttons_style
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_buttons_style
classic
```
</details>

Set 'classic':
```bash
opencli config update filemanager_buttons_style classic
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_buttons_style classic
Updated filemanager_buttons_style to classic
```
</details>

Set 'modern':
```bash
opencli config update filemanager_buttons_style modern
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_buttons_style modern
Updated filemanager_buttons_style to modern
```
</details>

### `menu_style`

Set the default OpenPanel menu style - Available 'classic' 'modern'. Users can still pick their own style from their profile menu, see [Menu Style](/docs/panel/dashboard/menu-style/). A missing or unknown value falls back to 'classic'.

**Check current value**:
```bash
opencli config get menu_style
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get menu_style
modern
```
</details>

Set 'classic':
```bash
opencli config update menu_style classic
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update menu_style classic
Updated menu_style to classic
```
</details>

Set 'modern':
```bash
opencli config update menu_style modern
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update menu_style modern
Updated menu_style to modern
```
</details>

### `filemanager_edit_size`
Maximum file size (in MB) allowed for editing files using the File Manager.

**Check current value**:
```bash
opencli config get filemanager_edit_size
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_edit_size
5
```
</details>

**Set new value**:
```bash
opencli config update filemanager_edit_size 5
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_edit_size 5
Updated filemanager_edit_size to 5
```
</details>


### `filemanager_view_size`
Maximum file size (in MB) allowed for viewing files in the File Manager.

**Check current value**:
```bash
opencli config get filemanager_view_size
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_view_size
5
```
</details>

**Set new value**:
```bash
opencli config update filemanager_view_size 5
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_view_size 5
Updated filemanager_view_size to 5
```
</details>


### `filemanager_download_size`
Maximum file size (in MB) allowed for downloading files using the File Manager.

**Check current value**:
```bash
opencli config get filemanager_download_size
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_download_size
2000
```
</details>

**Set new value**:
```bash
opencli config update filemanager_download_size 2000
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_download_size 2000
Updated filemanager_download_size to 2000
```
</details>


### `filemanager_upload_size`
Maximum file size (in MB) allowed for uploading files in the File Manager.

**Check current value**:
```bash
opencli config get filemanager_upload_size
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_upload_size
2000
```
</details>

**Set new value**:
```bash
opencli config update filemanager_upload_size 2000
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_upload_size 2000
Updated filemanager_upload_size to 2000
```
</details>


### `filemanager_compress_max_time`
Maximum time (in seconds) allowed for compressing files using File Manager.

**Check current value**:
```bash
opencli config get filemanager_compress_max_time
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_compress_max_time
5
```
</details>

**Set new value**:
```bash
opencli config update filemanager_compress_max_time 5
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_compress_max_time 5
Updated filemanager_compress_max_time to 5
```
</details>


### `filemanager_extract_max_time`
Maximum time (in seconds) allowed for extracting archives using File Manager.

**Check current value**:
```bash
opencli config get filemanager_extract_max_time
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_extract_max_time
5
```
</details>

**Set new value**:
```bash
opencli config update filemanager_extract_max_time 5
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_extract_max_time 5
Updated filemanager_extract_max_time to 5
```
</details>


### `filemanager_download_max_time`
Maximum time (in seconds) allowed for downloading files.

**Check current value**:
```bash
opencli config get filemanager_download_max_time
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_download_max_time
60
```
</details>

**Set new value**:
```bash
opencli config update filemanager_download_max_time 60
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_download_max_time 60
Updated filemanager_download_max_time to 60
```
</details>


### `filemanager_edit_extensions`

Allowed file extensions for editing in the File Manager.
**Check current value**:
```bash
opencli config get filemanager_edit_extensions
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_edit_extensions
".txt .md error_log .log env gitconfig cfg htaccess .ini .php .sh .html .json .htm .html5 .xml .py .php5 .php7 .php8 .sql .css .js .conf Gemfile .java"
```
</details>

**Set new value**:
```bash
opencli config update filemanager_edit_extensions ".txt .md error_log .log env gitconfig cfg htaccess .ini .php .sh .html .json .htm .html5 .xml .py .php5 .php7 .php8 .sql .css .js .conf"
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_edit_extensions ".txt .md error_log .log env gitconfig cfg htaccess .ini .php .sh .html .json .htm .html5 .xml .py .php5 .php7 .php8 .sql .css .js .conf"
Updated filemanager_edit_extensions to .txt .md error_log .log env gitconfig cfg htaccess .ini .php .sh .html .json .htm .html5 .xml .py .php5 .php7 .php8 .sql .css .js .conf
```
</details>


### `filemanager_image_extensions`
Allowed image file extensions to be viewed using Fle Manager.

**Check current value**:
```bash
opencli config get filemanager_image_extensions
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_image_extensions
".jpg .jpeg .png .gif .webp .avif"
```
</details>

**Set new value**:
```bash
opencli config update filemanager_image_extensions ".jpg .jpeg .png .gif .webp .avif"
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_image_extensions ".jpg .jpeg .png .gif .webp .avif"
Updated filemanager_image_extensions to .jpg .jpeg .png .gif .webp .avif
```
</details>


### `filemanager_archives_extensions`
Allowed archive file extensions for compression/extraction.

**Check current value**:
```bash
opencli config get filemanager_archives_extensions
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get filemanager_archives_extensions
".zip .tar .gz .tar.gz"
```
</details>

**Set new value**:
```bash
opencli config update filemanager_archives_extensions ".zip .tar .gz .tar.gz"
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update filemanager_archives_extensions ".zip .tar .gz .tar.gz"
Updated filemanager_archives_extensions to .zip .tar .gz .tar.gz
```
</details>


### `cron_max_file_size_kb`
Max size in KB for user's cronjob file `/home/{context}/crons.ini` to be editable via *OpenPanel > CronJobs* page. Default is `100`.

**Check current value**:
```bash
opencli config get cron_max_file_size_kb
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get cron_max_file_size_kb
100
```
</details>

**Set new value**:
```bash
opencli config update cron_max_file_size_kb 100
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update cron_max_file_size_kb 100
Updated cron_max_file_size_kb to 100
```
</details>


### `mysql_import_max_size_gb`
Max size in GB for file that is allowed to be imported via *OpenPanel > MySQL > Import* page. Default is `1`.

**Check current value**:
```bash
opencli config get mysql_import_max_size_gb
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mysql_import_max_size_gb
1
```
</details>

**Set new value**:
```bash
opencli config update mysql_import_max_size_gb 100
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mysql_import_max_size_gb 100
Updated mysql_import_max_size_gb to 100
```
</details>

### `mysql_startup_time`
Max time in seconds to wait on *OpenPanel > MySQL* page when initializing database (first time user open the page) before considering mysql failed. Default is 10.

**Check current value**:
```bash
opencli config get mysql_startup_time
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mysql_startup_time
30
```
</details>

**Set new value**:
```bash
opencli config update mysql_startup_time `0
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mysql_startup_time `0
Updated mysql_startup_time to `0
```
</details>


### `mysql_restricted_usernames`
List of usernames that will be considered system users and can not be edited via OpenPanel UI.

**Check current value**:
```bash
opencli config get mysql_restricted_usernames
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mysql_restricted_usernames
"mysql.sys mysql sys mariadb.sys phpmyadmin mysql.session mysql.infoschema root debian-sys-maint healthcheck percona.telemetry"
```
</details>

**Set new value**:
```bash
opencli config update mysql_restricted_usernames "mysql.sys mysql sys mariadb.sys phpmyadmin mysql.session mysql.infoschema root debian-sys-maint healthcheck"
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mysql_restricted_usernames "mysql.sys mysql sys mariadb.sys phpmyadmin mysql.session mysql.infoschema root debian-sys-maint healthcheck"
Updated mysql_restricted_usernames to mysql.sys mysql sys mariadb.sys phpmyadmin mysql.session mysql.infoschema root debian-sys-maint healthcheck
```
</details>


### `mysql_restricted_databases`
List of databases that will be considered system databases and can not be edited via OpenPanel UI.

**Check current value**:
```bash
opencli config get mysql_restricted_databases
```

<details>
  <summary>Example output</summary>

```bash
# opencli config get mysql_restricted_databases
"information_schema performance_schema mysql phpmyadmin sys mariadb.sys"
```
</details>

**Set new value**:
```bash
opencli config update mysql_restricted_databases "information_schema performance_schema mysql phpmyadmin sys mariadb.sys"
```

<details>
  <summary>Example output</summary>

```bash
# opencli config update mysql_restricted_databases "information_schema performance_schema mysql phpmyadmin sys mariadb.sys"
Updated mysql_restricted_databases to information_schema performance_schema mysql phpmyadmin sys mariadb.sys
```
</details>
