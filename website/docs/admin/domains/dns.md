---
sidebar_position: 2
---

# DNS zone editor

Edit the DNS zone for any domain hosted on the server.

:::info
The DNS zone editor in OpenAdmin is only shown in the **Domains** sidebar group when the **dns** module is enabled on the server.
:::


<Tabs>
  <TabItem value="openadmin-domains-list" label="With OpenAdmin" default>

1. Navigate to **OpenAdmin > Domains > DNS zone editor**.
2. The page lists every domain on the server with its owner, whether it has a zone file, the number of records in it and when the zone was last changed. Type in the search box to filter by domain or owner, then click **Edit zone** next to the domain. Domains marked **No zone file** have no file in the BIND zones directory, so they aren't served by the local DNS server.
   ![DNS zone editor listing the domains with their owner, zone file, record count and last change](/img/openadmin-screenshots/domains/dns-list.png#gh-light-mode-only)
   ![DNS zone editor listing the domains with their owner, zone file, record count and last change](/img/openadmin-screenshots/domains/dns-list_dark.png#gh-dark-mode-only)
3. Make your changes in the editor and click **Save** when done.
   ![DNS zone editor showing the zone file of a domain with the Save button](/img/openadmin-screenshots/domains/dns-edit.png#gh-light-mode-only)
   ![DNS zone editor showing the zone file of a domain with the Save button](/img/openadmin-screenshots/domains/dns-edit_dark.png#gh-dark-mode-only)

Before overwriting the zone file, OpenAdmin automatically creates a temporary backup of the previous version. The new content is validated (`named-checkzone`) before it is applied:

- ✅ If validation passes, the zone file is saved, BIND is reloaded (`rndc reload`), and propagation begins immediately.
- ⚠️ If there are syntax errors, the save is rejected and the zone file is automatically reverted to its previous version — your changes are **not** applied.

> If you reload the page after a failed save with unsaved changes still pending, OpenAdmin will offer to restore your last edited (but not yet valid) version from its temporary backup so you don't lose your work.

  </TabItem>
  <TabItem value="CLI-domains-list" label="With OpenCLI">

DNS zone files are stored in `/etc/bind/zones/`:

Edit the zone file `/etc/bind/zones/<DOMAIN_NAME>.zone` with your text editor, save, and to reload:

```bash
opencli domains-dns check <DOMAIN_NAME>
opencli domains-dns reload <DOMAIN_NAME>
```

  </TabItem>
</Tabs>

