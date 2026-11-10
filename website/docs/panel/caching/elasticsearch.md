---
sidebar_position: 7
---

# Elasticsearch

[Elasticsearch](https://www.elastic.co/elasticsearch/) is a powerful open-source search and analytics engine commonly used for:

- Full-text search  
- Log analytics  
- Real-time data exploration  

![ElasticSearch page with the service status, TCP server and port](/img/openpanel-screenshots/caching/elasticsearch-page.png#gh-light-mode-only)
![ElasticSearch page with the service status, TCP server and port](/img/openpanel-screenshots/caching/elasticsearch-page_dark.png#gh-dark-mode-only)

Use **OpenPanel > Cache & search > Elasticsearch** to manage the Elasticsearch service, enabling fast and scalable search functionality across your applications and systems.

---

## Status

Enable or disable the Elasticsearch service.

- The current status is displayed.  
- Click **Click to enable** or **Click to disable**. Disabling asks you to confirm first, since it stops and removes the service until you enable it again.

---

## TCP

Connection details for accessing the Elasticsearch API:

- **Server:** `elasticsearch`  
- **Port:** `9200`

Use these settings in external tools like `curl`, Kibana, Grafana, or application SDKs to query or manage data.

---

## Container

While the service is running, real-time resource usage (CPU, memory, network and block I/O) is displayed. Click **Edit limits** to adjust the container's resource limits from the Containers page.

---

## Logs

Monitor service activity and debug issues by viewing ElasticSearch logs.

- Click **View container log** to see the latest output of the service.
