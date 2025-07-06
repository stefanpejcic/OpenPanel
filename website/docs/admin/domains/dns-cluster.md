---
sidebar_position: 3
---

# DNS Cluster

:::danger
This feature is experimental and not yet ready for production deployment.
:::

DNS clustering allows you to synchronize DNS records from one OpenPanel server to other machines.

To use this feature, **all machines in the cluster must be running BIND9**—either through the OpenPanel installation or as a standalone service/container.

---

## Enable Clustering

Click **Enable DNS Clustering** to activate the feature.

---

### Add Slave Servers

To set up a slave BIND9 server, follow the instructions in [this guide](https://community.openpanel.org/d/128-slave-dns-servers/4).

Once the slave is ready:

1. Enter the slave server's IP address in the form.
2. OpenAdmin will attempt to connect and immediately sync DNS zones.
3. All future domains and DNS changes made by users will be automatically propagated to the slave server.

> ✅ To confirm the setup is working, check that the DNS zones appear on the slave server.

---

### Remove Slave Server

Currently, slave servers can **only be removed via the terminal**.

---

## Disable Clustering

Click **Disable DNS Clustering** to turn off the feature.
