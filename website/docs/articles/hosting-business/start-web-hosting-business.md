---
sidebar_label: "Start a Web Hosting Business"
description: "How to start a web hosting business with OpenPanel: choosing servers, installing the control panel, hosting plans and pricing, billing automation with WHMCS or FOSSBilling, branding, email, backups, security and support - a complete launch checklist."
---

# How to Start a Web Hosting Business with OpenPanel

Starting a hosting company has never been cheaper: a rented server, a control panel and a billing system are all you need to sell hosting to your first customers. This guide walks through every step of launching a shared hosting business with **OpenPanel** - from choosing a server to taking the first payment.

---

## Why OpenPanel for a Hosting Business

- **Flat price per server** - [OpenPanel Enterprise](/enterprise) costs the same for 10 or 1,000 accounts. Your panel cost per customer drops as you grow - unlike per-account licensing.
- **Real isolation** - every customer's web server, PHP and database run in their own containers with their own CPU and RAM limits. One hacked or busy site doesn't take down the others.
- **Everything included** - one-click apps (WordPress and 20+ others), free SSL, email, DNS, backups, reseller accounts, API and billing integrations.
- **Easy migration** - import customers from cPanel backups.

---

## Step 1: Decide What You'll Sell

Pick a focus - it decides your server size, plans and marketing:

- **Shared hosting** for small businesses and blogs.
- **WordPress / WooCommerce hosting** for a specific niche (agencies, restaurants, local businesses).
- **Reseller hosting** for web designers - see [Reseller hosting](/docs/articles/hosting-business/reseller-hosting/).
- **Hosting bundled with your services** - if you build websites, hosting becomes recurring revenue.

A niche is easier to market than "cheap hosting for everyone".

---

## Step 2: Get a Server

Rent a VPS or dedicated server from a provider with a good network and data centers near your customers. For a start:

- **4-8 CPU cores, 16-32 GB RAM, NVMe storage** comfortably hosts dozens to 100+ small sites.
- **Clean IPv4 address**, reverse DNS support, and **port 25 open** (or plan for an SMTP relay).
- **Snapshots or backups** at the provider level as an extra safety net.

Sizing help: [System requirements and sizing](/docs/articles/install-update/system-requirements/) and the [resource calculator](/calculator).

Provider guides: [Hetzner](/docs/articles/install-update/install-on-hetzner/) · [OVHcloud](/docs/articles/install-update/install-on-ovh/) · [Contabo](/docs/articles/install-update/install-on-contabo/) · [DigitalOcean](/docs/articles/install-update/install-on-digitalocean/) · [Vultr](/docs/articles/install-update/install-on-vultr/) · [Linode](/docs/articles/install-update/install-on-linode/) · [AWS](/docs/articles/install-update/install-on-aws/)

---

## Step 3: Install OpenPanel and Activate Enterprise

```bash
bash <(curl -sSL https://openpanel.org)
```

Then [upgrade to Enterprise and activate the license](/docs/articles/license/upgrade_to_openpanel_enterprise_and-activate_license/) - the free Community edition is for personal use and limited to 3 accounts. You can [try Enterprise free for 30 days](/trial).

After installing, follow the [post-install steps](/docs/admin/intro/#post-install-steps).

---

## Step 4: Set Up Your Brand and Infrastructure

1. **Panel domain and SSL** - e.g. `panel.yourhost.com`, without ports in the URL.
2. **Custom nameservers** - `ns1.yourhost.com` and `ns2.yourhost.com`, so customers point domains to you. See [Configure nameservers](/docs/articles/domains/how-to-configure-nameservers-in-openpanel/).
3. **Branding** - logo, favicon, colors, login page, email templates.

The full checklist: [White-label OpenPanel](/docs/articles/hosting-business/white-label-openpanel/).

---

## Step 5: Create Your Hosting Plans

Create 3-4 plans with clear limits (websites, storage, email accounts) and matching feature sets. Ready-made templates with commands: [Suggested hosting plan limits](/docs/articles/hosting-business/hosting-plan-templates/).

---

## Step 6: Set Up Email Properly

Hosting customers expect email to "just work". Before your first customer:

- [Enable email hosting](/docs/articles/user-experience/how-to-setup-email-in-openpanel/) on the server.
- Set [reverse DNS (PTR)](/docs/articles/email/reverse-dns-ptr-record/) for your server IP.
- Check [SPF, DKIM and DMARC](/docs/articles/email/why-emails-go-to-spam/) and score 10/10 on [mail-tester](/docs/articles/email/test-email-with-mail-tester/).
- If your provider blocks port 25, configure an [SMTP relay](/docs/articles/email/port-25-blocked-smtp-relay/).

---

## Step 7: Backups and Security

Customers trust you with their business. Set up:

- **Automatic backups** to remote storage - [Configure backups](/docs/articles/backups/configuring-backups/). Test a restore before launch.
- **Server hardening** - [Securing OpenPanel](/docs/articles/security/securing-openpanel/): firewall, restricted OpenAdmin access, 2FA for admins.
- **Malware scanning** - [ImunifyAV](/docs/articles/security/setup-imunifyav/).
- **Monitoring and alerts** - CPU, RAM, disk and swap thresholds in [Notifications](/docs/admin/settings/notifications/).
- **Email sending limits** per plan (max hourly emails), so a hacked site can't blacklist your IP.

---

## Step 8: Billing and Automation

Automate sign-up, payment, account creation, suspension for unpaid invoices and upgrades with a billing system:

| Billing system | Type |
|---|---|
| [WHMCS](/docs/articles/extensions/openpanel-and-whmcs/) | Paid, industry standard |
| [FOSSBilling](/docs/articles/extensions/openpanel-and-fossbilling/) | Free, open source |
| [Blesta](/docs/articles/extensions/openpanel-and-blesta/) | Paid, developer-friendly |
| [ClientExec](/docs/articles/extensions/openpanel-and-clientexec/) | Paid |
| [WISECP](/docs/articles/extensions/openpanel-and-wisecp/) | Paid |

Customers order on your website, pay, and receive their login automatically - with single sign-on into OpenPanel from the client area.

---

## Step 9: Pricing and Costs

A simple monthly cost model for a single server:

| Cost | Example |
|---|---|
| Server (8 cores, 32 GB RAM, NVMe) | €40-80 |
| OpenPanel Enterprise license | €14.95 |
| Backup storage | €5-15 |
| Billing system | €0-20 |
| Domain, email relay, misc. | €5-10 |
| **Total** | **≈ €65-140 / month** |

At €5-10 per account per month, you break even with roughly 10-20 customers - and every customer after that is margin. See [OpenPanel pricing](/docs/articles/license/pricing/).

---

## Step 10: Launch and Grow

- **Migrate existing sites** you manage - [import cPanel backups](/docs/articles/transfers/import-cpanel-backup-to-openpanel/) or [migrate WordPress sites](/docs/articles/websites/how-to-upload-wordpress-website-to-openpanel/).
- **Write a simple terms of service and privacy policy**, and a clear acceptable-use policy (no spam, no abuse).
- **Offer good support** - it's the main reason customers choose small hosts. A help center with your own articles, linked from the panel, reduces tickets.
- **Add servers** as you grow, and use [account transfers](/docs/articles/transfers/transfer-openpanel-account-to-another-server/) to balance load.
- **Add reseller hosting** once you have spare capacity.

---

## Launch Checklist

- [ ] Server installed, OpenPanel Enterprise active
- [ ] Panel domain with SSL, custom nameservers
- [ ] Branding: logo, colors, login page, email templates
- [ ] Hosting plans and feature sets created
- [ ] Email: PTR, SPF, DKIM, DMARC, mail-tester 10/10
- [ ] Remote backups configured and a restore tested
- [ ] Firewall, admin 2FA, monitoring alerts
- [ ] Billing system connected and a test order completed
- [ ] Terms of service, privacy policy, support channel

---

## Related

- [OpenPanel as a cPanel alternative](/cpanel-alternative)
- [Reseller hosting](/docs/articles/hosting-business/reseller-hosting/)
- [FAQ](/docs/articles/faq/)
