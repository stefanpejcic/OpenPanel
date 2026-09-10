---
sidebar_position: 3
---

# Logs

The **Containers > Logs** page allows you to view container logs (`docker logs`) directly from the OpenPanel interface.

## Requirements

To access this feature:

- The **Docker** module must be enabled **server-wide** by an Administrator.
- Your account must have the **Docker** feature enabled.

## Accessing Logs

1. In the OpenPanel menu, go to **Containers > Logs**.
2. Click on **Select Container** to display a list of all available services.
3. Select the service you want to view logs for.
4. The log output for the selected container will appear below.

You can optionally adjust the number of log lines shown using the dropdown menu in the top-right corner of the logs panel.

> 💡 Logs are fetched using `docker logs` and show the real-time output of the container’s stdout and stderr streams.
