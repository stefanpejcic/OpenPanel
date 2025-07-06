---
sidebar_position: 1
---

# Service Status

The Service Status section allows you to view and control the status of system services and containers running on your server.

This table provides key details for each service:

* **Service** – Display name of the service.
* **Status** – Indicates if the service is active or inactive.
* **Version** – Current version (if available).
* **Real Name** – Internal service name or container name (e.g., admin for OpenAdmin).
* **Type** – Identifies whether the service is a system process or a container.
* **Port** – Lists the ports used by the service.
* **Monitoring** – Shows whether the service is actively being monitored and logged.
* **Action** – Options to start, stop, or restart the service.

## Edit Services

You can customize which services appear and are manageable from this section by clicking the **Edit Services** button.

Services are configured in JSON format:

* **name** – Display name for the service.
* **type** – Either `system` or `container`.
* **real_name** – Internal service or container identifier.

Default OpenPanel services are actively monitored by SentinelAI, which will try to diagnose and restart failed services automatically. If you manually stop a service, remember to disable this monitoring feature via the [OpenAdmin Notifications](/docs/admin/notifications/) page.

