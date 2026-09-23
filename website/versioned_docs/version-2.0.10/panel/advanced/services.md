---
sidebar_position: 2
---

# Services

The Services page allows you to view current real-time resource usage for your services and manage them.

:::info
The Services page appears only if the 'services' module is enabled on the server and included as a feature in your hosting plan.
To start or stop services, use the Docker feature (if available) or contact your hosting provider.
:::

To manage a service, select it's name from the select options.

![Choose Service page with the service dropdown](/img/openpanel-screenshots/advanced/services_list-page.png#gh-light-mode-only)
![Choose Service page with the service dropdown](/img/openpanel-screenshots/advanced/services_list-page_dark.png#gh-dark-mode-only)

On the single service page you can view:

- current service status
- resource usage (CPU Usage, Memory Usage, Memory %, Network I/O, Block I/O, PIDs..)
- container name (to be used to connect to service from other containers)
- manage options: enable/disable
- container logs

![Service page for redis with its status, container resource usage and logs](/img/openpanel-screenshots/advanced/services-page.png#gh-light-mode-only)
![Service page for redis with its status, container resource usage and logs](/img/openpanel-screenshots/advanced/services-page_dark.png#gh-dark-mode-only)
