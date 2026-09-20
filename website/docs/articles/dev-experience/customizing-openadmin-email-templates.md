# Customizing OpenAdmin Email Templates

OpenAdmin sends outbound notification emails (new user welcome emails, password/login alerts, admin notifications, and the daily usage report) using HTML templates built into the `admin` binary.

Administrators can override any of these templates with their own HTML, without editing any code and without recompiling: OpenAdmin picks up overrides once, at startup.

## Override directory

**Directory:** `/etc/openpanel/openadmin/email_templates/`

This directory doesn't exist by default. To customize a template, copy the original file below into the directory (creating it if needed) under the exact same name, then edit it as you like:

| File name | Used for | Original file |
| --- | --- | --- |
| `email_new_user.html` | Welcome email sent when a new user account is created | [view on GitHub](https://github.com/stefanpejcic/openadmin/blob/main/internal/webtemplates/email_new_user.html) |
| `email_user_notifications.html` | Password-change and new-login alerts sent to a user | [view on GitHub](https://github.com/stefanpejcic/openadmin/blob/main/internal/webtemplates/email_user_notifications.html) |
| `email_admin_notifications.html` | Generic admin notification emails | [view on GitHub](https://github.com/stefanpejcic/openadmin/blob/main/internal/webtemplates/email_admin_notifications.html) |
| `email_daily_system_report.html` | The daily usage report email (CPU/RAM/disk usage, user/domain/website counts) | [view on GitHub](https://github.com/stefanpejcic/openadmin/blob/main/internal/webtemplates/email_daily_system_report.html) |

Each file is a standard Go `html/template` file and can use the same `{{.Field}}` placeholders as the built-in template it replaces (e.g. `{{.Title}}`, `{{.Message}}`, `{{.Hostname}}`).

## Applying changes

Overrides are only read once, at `admin` service startup. After adding or editing a file in the override directory, restart the service:

```bash
systemctl restart admin
```

## Fallback behavior

- If a file for a given template name isn't present, OpenAdmin uses its built-in default for that email.
- If a file is present but fails to parse as a valid template, OpenAdmin logs the error and falls back to the built-in default for that email — a broken override never blocks outbound mail.
- You only need to add the files for the templates you want to change; the rest keep using the built-in defaults.
