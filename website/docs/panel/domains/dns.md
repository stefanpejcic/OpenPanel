---
sidebar_position: 4
---

# DNS


The DNS Zone Editor enables you to manage and edit Domain Name System (DNS) zone files, which are essential for mapping domain names to IP addresses and handling various DNS records such as A, CNAME, MX, and TXT records.

On the main _DNS Zone Editor_ page, you can choose a domain name to view and modify its records.

![domain_dns_editor_1.png](/img/panel/v1/domains/domain_dns_editor_1.png)

## Edit record

To modify an existing record, click on the 'Edit' button next to it. This will make the row editable, allowing you to save changes by clicking the 'Save' button or discard them by selecting the 'Cancel' button.

![domain_dns_editor_2.png](/img/panel/v1/domains/domain_dns_editor_2.png)

## Create record

To create a new DNS record, click on the 'Add Record' button.

This will add a new row at the top of the table where you can enter the record. On mobile devices, the 'Add' button will trigger a modal instead.

![domain_dns_editor_3.png](/img/panel/v1/domains/domain_dns_editor_3.png)

Supported DNS types: `A`, `AAAA`, `CAA`, `CNAME`, `MX`, `SRV`, `TXT`

## Delete record

To remove an existing DNS record, click the delete button next to it. The button text will change to 'Confirm', and on the second click, it will permanently delete the record.

![domain_dns_editor_4.png](/img/panel/v1/domains/domain_dns_editor_4.png)


## Export Zone
To export a DNS zone, click on the 'Export Zone' button. The button text will change to 'Confirm', and on the second click, the zone will be exported.

![openpanel_export_dns_zone.png](/img/panel/v1/domains/openpanel_export_dns_zone.png)


## Reset Zone

Resetting the zone will delete all existing records and create a default DNS zone, as if the domain had just been added again.

![openpanel_reset_dns_zone.gif](/img/panel/v1/domains/openpanel_reset_dns_zone.gif)


## Advanced Editor

The Advanced Editor allows direct editing of the DNS zone file, including adding custom records or comments. It can also be used to import a DNS zone by pasting content from another server.

![openpanel_advanced_dns_zone.png](/img/panel/v1/domains/openpanel_advanced_dns_zone.png)
