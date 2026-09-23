---
sidebar_position: 6
---

# OpenSearch

[OpenSearch](https://opensearch.org) is an open-source search and analytics engine, fully compatible with Elasticsearch. It is ideal for:

- Full-text search  
- Log analytics  
- Real-time data exploration  

![OpenSearch page with the service status, TCP server and port](/img/openpanel-screenshots/caching/opensearch-page.png#gh-light-mode-only)
![OpenSearch page with the service status, TCP server and port](/img/openpanel-screenshots/caching/opensearch-page_dark.png#gh-dark-mode-only)

You can manage the OpenSearch service through **OpenPanel > Cache & Search > OpenSearch**, enabling fast, scalable search capabilities for your applications and systems.

---

## Status

Enable or disable the OpenSearch service.

- The current status is displayed.  
- Use the toggle option to **enable** or **disable** the service as needed.

---

## TCP

Connection details for accessing the OpenSearch API:

- **Server:** `opensearch`  
- **Port:** `9200`

Use these settings in external tools like `curl`, Kibana, Grafana, or application SDKs to query or manage data.

---

## Container

While the service is running, real-time resource usage (CPU, memory, network and block I/O) is displayed. Click **Edit limits** to adjust the container's resource limits from the Containers page.

---

## Logs

Monitor service activity and debug issues by viewing OpenSearch logs.

- Click **View service log** to open the service log file and inspect output in real time.
