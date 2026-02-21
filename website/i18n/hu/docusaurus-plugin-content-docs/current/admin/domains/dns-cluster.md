---
sidebar_position: 3
---

# DNS Cluster

DNS clustering allows you to synchronize DNS records from one OpenPanel server to other machines.

To use this feature, **all machines in the cluster must be running BIND9** — either through the OpenPanel installation or as a standalone service/container.

---

## Enable Clustering

Click **Enable DNS Clustering** to activate the feature.

---

### Add Slave Servers

To set up a slave BIND9 server, follow the instructions in [How-to Guides > DNS Clustering](/docs/articles/domains/how-to-setup-dns-cluster-in-openpanel/).

Once the slave is ready:

1. Enter the slave server's IP address in the form on master server.
2. OpenAdmin will attempt to connect and immediately sync DNS zones.
3. All future domains and DNS changes made by users will be automatically propagated to the slave server.

> To confirm the setup is working, check that the DNS zones appear on the slave server.

---

### Remove Slave Server

Currently, slave servers can **only be removed via the terminal**.

---

## Disable Clustering

Click **Disable DNS Clustering** to turn off the feature.
