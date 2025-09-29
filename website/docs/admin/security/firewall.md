---
sidebar_position: 2
---

# Firewall

OpenPanel supports [Sentinel Firewall](https://sentinelfirewall.org/) - a fork of the *ConfigServer Security and Firewall* (CSF).

## CSF

Firewall (CSF) UI is displayed on **OpenAdmin > Firewall**.

For instructions on how to use the CSF UI, please refer to [Sentinel Firewall official documentation](https://sentinelfirewall.org/docs/usage/introduction/).

![csf firewall](/img/admin/firewall_csf.png)


## External Firewall

Some cloud providers, like [Hetzner](https://docs.hetzner.com/robot/dedicated-server/firewall/), offer their own external firewalls. If you are using an external firewall, ensure the following ports are open for OpenPanel services to be accessible: `53` `80` `443` `2083` `2087` `32768:60999`

If you are [using a custom port for OpenPanel instead of the default 2083](/docs/admin/settings/general/#change-openpanel-port), ensure that port is open as well.

