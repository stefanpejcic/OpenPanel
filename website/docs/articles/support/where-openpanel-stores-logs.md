# OpenPanel Log Files locations

OpenPanel generates the following logs:

| Log file | Description |
|----------|-------------|
|`/var/log/openpanel/user/access.log`| access logs for the OpenPanel|
|`docker logs openpanel`|verbose logs and errors for the OpenPanel when dev_mode is enabled|
|`/var/log/openpanel/admin/access.log`|access logs for the OpenAdmin |
|`/var/log/openpanel/admin/api.log`|OpenAdmin API log |
|`/var/log/openpanel/admin/login.log`|OpenAdmin successful logins|
|`/var/log/openpanel/admin/failed_login.log`|OpenAdmin failed login attempts |
|`/var/log/openpanel/user/failed_login.log`|OpenPanel failed login attempts |
|`/var/log/openpanel/admin/notifications.log`|OpenAdmin notifications (system and 8ser alerts) |
|`/var/log/openpanel/admin/cron.log`|OpenAdmin systme cron logs |
|`/var/log/openpanel/admin/error.log`|OpenAdmin error log |


:::info
By default, logging for OpenPanel, OpenAdmin, and the API is minimal in production. Administrators can [enable dev_mode](/docs/articles/support/how-to-troubleshoot-openpanel-error/#dev-mode) during troubleshooting to view more verbose logs.
:::


Services used by OpenPanel have the following logs:

| Log file | Description |
|----------|-------------|
|`docker logs openpanel_mysql`|MySQL error log|
|`docker logs openpanel_dns`|DNS service logs|


Logs can be viewed from the [OpenAdmin > Services > Log Viewer](/docs/admin/services/log_viewer/) and Administrators can even [add custom log files to OpenAdmin Log Viewer](/docs/admin/services/log_viewer/#how-to-add-more-files-to-openadmin-log-viewer).
