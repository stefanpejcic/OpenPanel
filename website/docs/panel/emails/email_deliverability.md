---
sidebar_position: 10
---

# Email Deliverability

Checks the live DNS records for your domains against the SPF, DKIM, and DMARC records OpenPanel expects, and shows whether email sent from each domain is likely to be trusted by recipient mail servers.

## Overview

Go to **OpenPanel > Emails > Email Deliverability** to see a table of all your domains with a status badge for each record:

| Column | Description |
|---|---|
| **Domain** | The domain being checked |
| **SPF** | Status of the domain's SPF TXT record |
| **DKIM** | Status of the `mail._domainkey` DKIM TXT record |
| **DMARC** | Status of the `_dmarc` DMARC TXT record |
| **Options** | **Details** button, opens the full record comparison for that domain |

Records are checked live via DNS lookup, so each row briefly shows "Checking..." before the status badge loads.

### Status badges

| Badge | Meaning |
|---|---|
| **OK** | The record is published and matches what OpenPanel expects |
| **Missing** | No matching TXT record was found |
| **Mismatch** | A DKIM record was found, but its value does not match the key generated for this server |
| **Wrong IP** | An SPF record was found, but it does not include this server's IP address |

## Viewing record details

Click **Details** next to a domain to see, for each of SPF, DKIM, and DMARC:

- **Current**: the value currently published in DNS for that record (or "Not found" if none exists)
- **Expected**: the value OpenPanel expects, based on this server's configuration

Use this comparison to update your DNS zone so that the current value matches the expected one.
