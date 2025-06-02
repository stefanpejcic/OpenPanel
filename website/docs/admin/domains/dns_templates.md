---
sidebar_position: 3
---

# Edit Zone Templates

This interface allows you to edit the templates that OpenPanel uses to create DNS zone files for new domains. You may wish to use this interface when you use custom DNS configurations.

![edit zone screenshot](/img/admin/dns_templates_admin.png)

- **IPv4 Template**: Template used for new user domains with IPv4 addresses.
- **IPv6 Template**: Template used for new user domains with IPv6 addresses.

You can use the following variables within DNS zone template files:

| Variable    | Description |
| -------- | ------- |
| `{ns1}`  | The primary nameserver’s hostname for the primary NS record.    |
| `{ns2}` | The secondary nameserver’s hostname for the secondary NS record.     |
| `{ns3}` | The tertiary nameserver’s hostname for the tertiary NS record.     |
| `{ns4}` | The quaternary nameserver’s hostname for the quaternary NS record.     |
| `{server_ip}` | The domain’s IPv4 or IPv6 address.     |

