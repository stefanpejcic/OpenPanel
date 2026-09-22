---
sidebar_position: 4
---

# DNS

![DNS zone editor listing A, MX, CNAME and TXT records with Edit and Delete buttons](/img/openpanel-screenshots/domains/dns-list.png)


The DNS Zone Editor enables you to manage and edit Domain Name System (DNS) zone files, which are essential for mapping domain names to IP addresses and handling various DNS records such as A, CNAME, MX, and TXT records.

On the main _DNS Zone Editor_ page, you can choose a domain name to view and modify its records.

## Edit record

To modify an existing record, click on the 'Edit' button next to it. This will make the row editable, allowing you to save changes by clicking the 'Save' button or discard them by selecting the 'Cancel' button.

![A DNS record row in edit mode with editable fields and the Save and Cancel buttons](/img/openpanel-screenshots/domains/dns_sections-edit.png)

## Create record

To create a new DNS record, click on the 'New Record' button.

This will add a new row at the top of the table where you can enter the record.

![New Record row added at the top of the DNS table with the record type, name, TTL and value fields](/img/openpanel-screenshots/domains/dns_sections-create.png)

Supported DNS types: `A`, `AAAA`, `CAA`, `CNAME`, `MX`, `SRV`, `TLSA`, `TXT`

## Delete record

To remove an existing DNS record, click the delete button next to it. The button text will change to 'Confirm', and on the second click, it will permanently delete the record.

![Delete button of a DNS record turned into a Confirm button after the first click](/img/openpanel-screenshots/domains/dns_sections-delete.png)

## Export Zone
To export a DNS zone, click on the 'Export Zone' button. The zone file will be downloaded immediately.

![DNS Zone Editor menu with the Edit DNS File, Export Zone and Reset options](/img/openpanel-screenshots/domains/dns_sections-toolbar.png)

## Reset Zone

Resetting the zone will delete all existing records and create a default DNS zone, as if the domain had just been added again.

![Reset Zone panel asking for confirmation before the DNS zone is restored to its default records](/img/openpanel-screenshots/domains/dns_sections-reset.png)

## Advanced Editor

The Advanced Editor allows direct editing of the DNS zone file, including adding custom records or comments. It can also be used to import a DNS zone by pasting content from another server.

![Advanced Editor showing the DNS zone file of the domain for direct editing](/img/openpanel-screenshots/domains/dns_editor-editor.png)
