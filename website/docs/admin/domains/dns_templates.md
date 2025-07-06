---
sidebar_position: 3
---

# Edit Zone Templates

This interface allows you to edit the DNS zone file templates that OpenPanel uses when creating zones for new domains.  
It is useful if you require custom DNS configurations.

![edit zone screenshot](/img/admin/dns_templates_admin.png)

- **IPv4 Template** – Used for new domains assigned IPv4 addresses.
- **IPv6 Template** – Used for new domains assigned IPv6 addresses.

---

## Available Template Variables

You can include the following variables in your DNS zone templates:

| Variable        | Description                                                                 |
|----------------|-----------------------------------------------------------------------------|
| `{ns1}`         | Hostname of the primary nameserver (used in the NS record).                |
| `{ns2}`         | Hostname of the secondary nameserver.                                      |
| `{ns3}`         | Hostname of the tertiary nameserver.                                       |
| `{ns4}`         | Hostname of the quaternary nameserver.                                     |
| `{server_ip}`   | IP address of the domain (either IPv4 or IPv6, depending on the template). |

---

## Restore Defaults

To reset a template to its default version:

1. Click **Restore Default** for the corresponding section.
2. Then click **Save Files** to apply the changes.

> ⚠️ Defaults are pulled from the official GitHub repository. Always review before saving.
