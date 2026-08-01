---
sidebar_position: 1
---

# Service Status

The Service Status section allows you to view and control the status of system services and containers running on your server.

This table provides key details for each service:

* **Service** – Display name of the service.
* **Status** – Indicates if the service is active (running) or inactive.
* **Version** – Reserved for the service's current version; not populated in the current release.
* **Real Name** – Internal service name or container name (e.g., admin for OpenAdmin).
* **Type** – Identifies whether the service is a `system` process or a `docker` container (shown in the table as "container").
* **Port** – Reserved for the ports used by the service; not populated in the current release.
* **Monitoring** – Shows whether the service is actively being monitored and logged.
* **Action** – Options to start, stop, or restart the service.

## Edit Services

You can customize which services appear and are manageable from this section by clicking the **Edit Services** button.

Services are configured in JSON format:

* **name** – Display name for the service.
* **type** – Either `system` or `docker` (services of type `docker` are labeled "container" in the table).
* **real_name** – Internal service or container identifier.

Core services (as listed under the **Monitoring** column) are tracked and, when one of them becomes unresponsive, an alert is sent to administrators. If you intentionally stop a monitored service, remember to disable the corresponding alert via the [OpenAdmin Notifications](/docs/admin/notifications/) page to avoid unnecessary alerts.

