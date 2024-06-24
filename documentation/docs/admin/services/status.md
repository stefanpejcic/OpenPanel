---
sidebar_position: 1
---

# Services Status

*OpenAdmin > Services > Status* allows Administrators to view current status and manage monitored system services or Docker containers.

![services_page](/img/admin/openadmin_services_status.png)

To add custom services to the list, edit the `/etc/openpanel/openadmin/config/services.json` file.

Default OpenPanel services are actively monitored by the SentinelAI and on failure, OpenAdmin will automatically try to determinate the root cause and restart the m. If you manually stop a service, you should also disable this feature from the  [OpenAdmin Notifications](/docs/admin/notifications/) page.

