---
sidebar_position: 1
---

# Service Status

The Service Status section allows you to view and control the status of system services and containers running on your server.

This table provides key details for each service:

- **Service** – Display name of the service.

- **Service Status** – Indicates if the service is active or inactive.

- **Version** – Current version (if available).

- **Real Name** – Internal service name or container name (e.g., admin for OpenAdmin).

- **Type** – Identifies whether the service is a system process or a container.

- **Port** – Lists the ports used by the service.

- **Monitoring** – Shows whether the service is actively being monitored and logged.

- **Action** – Options to start, stop, or restart the service.

## Edit Monitored Services
You can customize which services appear and are manageable from this section by clicking Edit Services.

Services are configured in JSON format:

- name – Display name for the service.

- type – Either system or container.

- on_dashboard – Whether to display it on the dashboard.

- real_name – Internal service or container identifier.

Default OpenPanel services are actively monitored by the SentinelAI and on failure, OpenAdmin will automatically try to determinate the root cause and restart them. If you manually stop a service, you should also disable this feature from the  [OpenAdmin Notifications](/docs/admin/notifications/) page.

