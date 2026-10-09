# Free Migration to OpenPanel

Effective Date: October 2026

## Overview

Customers with an **annual OpenPanel Enterprise license** can have their existing hosting accounts migrated to OpenPanel by our team at no extra cost. Send us access to your current server and we handle the transfer of websites, databases, email and DNS.

## Eligibility

The free migration service is available only if ALL of the following conditions are met:

- You have an active **yearly** (annual billing) [Enterprise license](https://my.openpanel.com/index.php?rp=/store/openpanel/enterprise-license). Monthly licenses, trial licenses and the Community Edition are not eligible.
- The license is activated on the destination server, and OpenPanel is installed and reachable.
- The source server runs one of the supported control panels listed below.
- The request is submitted by the license holder through our support system.

One free migration is included per annual license.

## Supported Source Panels

| Source panel | Notes |
|---|---|
| cPanel / WHM | All versions with root or WHM access |
| Plesk | Linux servers only |
| DirectAdmin | Admin or root access required |
| CyberPanel | Root access required |
| OpenPanel | Community or Enterprise edition, root access required |

> **Not supported:**
> - Custom server stacks (plain LAMP/LEMP, Docker setups, bare VPS without a control panel).
> - Windows / IIS servers, including Plesk for Windows.
> - Website builders (Wix, Squarespace, Weebly and similar).
> - Any control panel not listed above.

## Limits & Pricing

| | |
|---|---|
| Included free | Up to **20 accounts** and **500 GB** of total data per migration |
| Additional accounts | **€10 per account** above 20 |
| Data above 500 GB | Quoted individually |

Total data includes website files, databases and mailboxes across all migrated accounts. Additional accounts and data are invoiced before the migration starts. An "account" is one hosting user on the source panel (a cPanel account, a Plesk subscription, a DirectAdmin user, a CyberPanel website owner or an OpenPanel user), including all of its domains, databases and mailboxes.

## What We Migrate

- User accounts and their domains, subdomains, addon and parked/alias domains
- Website files
- MySQL / MariaDB databases and database users
- Email accounts, mailboxes, forwarders and autoresponders
- DNS zones and records
- Cron jobs
- FTP accounts
- SSL certificates (existing certificates are copied, or new Let's Encrypt certificates are issued once DNS points to the new server)

## What Is Not Included

- Server-level custom configuration (custom Apache/Nginx rules, custom PHP extensions or compiled modules, kernel or firewall tuning)
- Software or services not managed by the source panel
- Billing system data (WHMCS, Blesta, etc.)
- Domain name transfers between registrars
- Changing DNS records or nameservers at your registrar - this is done by you once you have verified the migrated sites
- Merging several source accounts into one account, or splitting one account into several
- Fixing issues that already existed on the source server (broken sites, outdated or incompatible application code)

## Migration Procedure

1. **Open a ticket** - submit a [migration request ticket](https://my.openpanel.com/submitticket.php?step=2&deptid=1) from the account that owns the annual license.
2. **Provide details** - include in the ticket:
    - Your license key and the destination server IP
    - Source panel name and version
    - Number of accounts to migrate and approximate total disk usage
    - Login credentials for the source server: root SSH access (preferred), or admin login for the panel
    - Root SSH access to the destination OpenPanel server
3. **Review** - we check access and size, confirm the number of accounts and, if the migration exceeds 20 accounts or 500 GB, send an invoice for the difference.
4. **Scheduling** - we agree on a migration window with you.
5. **Migration** - we transfer all accounts to the OpenPanel server. Your source server keeps running and serving traffic during this time.
6. **Verification** - you test the migrated websites and email, for example by using a hosts file entry or the temporary domain.
7. **DNS switch** - you point your domains or nameservers to the new server.
8. **Final sync** - on request, we run one final sync of files, databases and mailboxes changed since the initial migration.
9. **Closing** - we close the ticket and delete all credentials you provided.

## Your Responsibilities

- Keep a full, current backup of the source server before and during the migration. OpenPanel LLC is not liable for data loss on the source or destination server.
- Make sure the destination server has enough disk space, memory and CPU for all migrated accounts.
- Lower the TTL of your DNS records at least 24 hours before the DNS switch to minimize downtime.
- Keep the source server online and accessible until the migration and final sync are complete.
- **Change all passwords** (server root, panel admin, SSH) after the migration is complete.

## Credentials & Security

Credentials are accepted **only through our support ticket system** - never send them by email, chat or social media. Credentials are used only for the migration, are not shared with third parties, and are removed from the ticket once the migration is complete.

## Timeframe

Migrations are usually started within 3 business days after all required information is received. Duration depends on the number of accounts and the amount of data.

## Right to Refuse

We may decline or stop a migration if the source server is not accessible, the data is corrupted or incomplete, the content violates our terms of service, or the migration is not technically possible. In that case we will explain the reason in the ticket.

A migration does not change the [Refund Policy](/refund-policy) - the license is activated before migration and activated licenses are not refundable.

## Questions

Contact us via a [support ticket](https://my.openpanel.com/submitticket.php?step=2&deptid=1) or at [info@openpanel.com](mailto:info@openpanel.com) before purchasing if you are unsure whether your server qualifies.
