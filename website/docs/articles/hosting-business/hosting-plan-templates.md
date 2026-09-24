---
sidebar_label: "Hosting Plan Templates"
description: "Suggested shared hosting plan limits for OpenPanel - Starter, Business, Pro, WordPress and reseller tiers - with CPU, RAM, disk, inodes, email and database limits, ready-to-run opencli commands, and how to size them for your server."
---

# Suggested Hosting Plan Limits (Templates for Shared Hosting Tiers)

Choosing plan limits is one of the first decisions when you start selling hosting. The templates below are a practical starting point for typical shared hosting tiers on OpenPanel - copy them, then adjust to your server and market.

In OpenPanel, every limit is **enforced per account**: CPU and RAM are real container limits, disk and inodes are quotas, and the number of domains, websites, databases, mailboxes and FTP accounts is checked by the panel. See [User Packages](/docs/admin/plans/hosting_plans/).

---

## Plan Templates

| Limit | Starter | Business | Pro | WordPress Pro |
|---|---|---|---|---|
| **CPU cores** | 1 | 2 | 4 | 2 |
| **Memory (RAM)** | 1 GB | 2 GB | 4 GB | 3 GB |
| **Disk** | 10 GB | 30 GB | 75 GB | 40 GB |
| **Inodes (files)** | 150,000 | 400,000 | 1,000,000 | 500,000 |
| **Port speed** | 100 Mbit/s | 200 Mbit/s | 500 Mbit/s | 200 Mbit/s |
| **Domains** | 1 | 5 | 20 | 3 |
| **Websites** | 1 | 5 | 20 | 3 |
| **Databases** | 2 | 10 | 50 | 5 |
| **Email accounts** | 5 | 25 | 100 | 10 |
| **Mailbox quota** | 2G | 5G | 10G | 5G |
| **Max hourly emails** | 100 | 300 | 500 | 200 |
| **FTP accounts** | 1 | 5 | 20 | 3 |
| **Feature set** | basic | default | default | default |

Notes:

- **Subdomains** don't count toward the domain limit, as long as the main domain is in the account - see [Subdomains, addon and parked domains](/docs/articles/domains/subdomain-addon-parked-domain/).
- **Websites** counts installed apps (WordPress, Node.js, Python, Website Builder, ...), not plain folders.
- `0` means **unlimited** for any limit. Avoid unlimited CPU and RAM on shared servers - one busy site could slow down everyone else.
- **Max hourly emails** protects your IP reputation if a customer's site is hacked and starts sending spam.

---

## Create the Plans from the Terminal

Paste these commands to create all four plans at once (plan names without spaces):

```bash
opencli plan-create name="Starter" description="1 website, 10 GB NVMe" emails=5 ftp=1 domains=1 websites=1 disk=10 inodes=150000 databases=2 cpu=1 ram=1 bandwidth=100 feature_set=basic max_email_quota=2G max_hourly_email=100

opencli plan-create name="Business" description="5 websites, 30 GB NVMe" emails=25 ftp=5 domains=5 websites=5 disk=30 inodes=400000 databases=10 cpu=2 ram=2 bandwidth=200 feature_set=default max_email_quota=5G max_hourly_email=300

opencli plan-create name="Pro" description="20 websites, 75 GB NVMe" emails=100 ftp=20 domains=20 websites=20 disk=75 inodes=1000000 databases=50 cpu=4 ram=4 bandwidth=500 feature_set=default max_email_quota=10G max_hourly_email=500

opencli plan-create name="WordPressPro" description="Managed WordPress, 3 sites" emails=10 ftp=3 domains=3 websites=3 disk=40 inodes=500000 databases=5 cpu=2 ram=3 bandwidth=200 feature_set=default max_email_quota=5G max_hourly_email=200
```

Or create them in **OpenAdmin → Hosting Plans → User Packages → Create New**. See [opencli plan](/docs/articles/opencli/plan/).

Use a [Feature Set](/docs/admin/plans/feature-manager/) that matches the tier - for example hide advanced features like the web terminal, containers or PHP.INI editor on the Starter plan, and enable them on Pro.

---

## How Many Accounts Fit on a Server?

CPU and RAM limits are **maximums**, not reservations. Most sites are idle most of the time, so you can safely sell more total limits than the server physically has (overselling). A common rule of thumb:

- Sell up to about **3-4×** the server's RAM in total plan RAM for mixed small sites.
- Stay closer to **1.5-2×** for WordPress/WooCommerce-heavy customers.
- Keep **disk** closer to 1:1 - files actually use the space, and you need room for backups.

Example: a server with **8 cores, 32 GB RAM and 500 GB NVMe** could comfortably host around **60-80 Starter** accounts, or **30-40 Business** accounts, depending on how busy the sites are.

Use the [resource calculator](/calculator) to model your mix, and watch real usage in **OpenAdmin → Server → Resource Usage** as you grow. See [System requirements and sizing](/docs/articles/install-update/system-requirements/).

---

## Pricing Tips

- Price by **value (sites, storage, email)** - customers compare those, not CPU cores.
- Offer **annual billing** with a discount; it improves cash flow and retention.
- Keep **3-4 tiers**; more options slow down buying decisions.
- Include **daily backups** and **free SSL** in every tier - both come built into OpenPanel.
- Since OpenPanel Enterprise is priced **per server, not per account**, your panel cost per customer drops as you add accounts.

---

## Change Plans Later

- Edit a plan to change limits for **all** its users at once.
- Move a single customer to another plan (upgrade/downgrade) from **OpenAdmin → Users → Edit**, or with `opencli user-change_plan`.
- With a [billing integration](/docs/articles/extensions/openpanel-and-whmcs/), upgrades and downgrades happen automatically when the customer changes their package.

---

## Related

- [Start a web hosting business with OpenPanel](/docs/articles/hosting-business/start-web-hosting-business/)
- [Set up reseller accounts](/docs/articles/hosting-business/reseller-hosting/)
- [Are plan limits hard or soft limits?](/docs/articles/containers/are-plan-limits-hard-or-soft-limits/)
