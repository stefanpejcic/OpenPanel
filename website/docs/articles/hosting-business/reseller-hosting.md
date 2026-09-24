---
sidebar_label: "Reseller Hosting"
description: "How to offer reseller hosting with OpenPanel: create reseller accounts, set account and disk limits, choose which plans they can sell, reseller branding, billing automation and pricing tips."
---

# How to Offer Reseller Hosting with OpenPanel

**Reseller hosting** lets your customers - web designers, agencies, small hosting companies - create and manage hosting accounts for *their own* clients on your server. You sell them a block of resources; they resell it under their name. It's one of the most profitable hosting products, because one reseller brings many accounts.

OpenPanel has reseller accounts built in (**Enterprise**). Resellers log in to OpenAdmin with their own account and see only the users and plans you allow.

---

## How It Works

```
You (Super Admin)
 └── Reseller "Agency A"   (max 20 accounts, 100 GB, plans: Starter + Business)
      ├── client1.com
      ├── client2.com
      └── ...
```

- **You** create the hosting plans and decide which ones each reseller may use.
- **The reseller** creates, suspends and deletes accounts for their clients - within their limits.
- **Each client** gets a normal OpenPanel account, fully isolated from other clients.

---

## Step 1: Prepare the Plans Resellers Can Sell

Resellers don't create plans - they assign plans you've made. Create a few tiers first, for example Starter and Business - see [Suggested hosting plan limits](/docs/articles/hosting-business/hosting-plan-templates/).

Tip: give reseller plans a clear name prefix (e.g. `R-Starter`, `R-Business`) so you can tell them apart from your direct-customer plans.

---

## Step 2: Create a Reseller

1. Go to **OpenAdmin → Accounts → Resellers**.
2. Click **Create New**, enter a username and password, and click **Create**.
3. Open the reseller's menu → **Edit Plans & Limits** and set:
   - **Max Accounts** - how many client accounts they can create,
   - **Max Disk Usage** - their total disk allowance across all clients,
   - **Hosting Plans** - which plans they may assign.

See [Resellers](/docs/admin/accounts/resellers/).

The reseller logs in at your OpenAdmin address (e.g. `https://panel.yourhost.com:2087`) and manages their users from **Accounts → Users**.

---

## Step 3: Reseller Branding

Resellers can set their own **logo** from their **Reseller Account** page - it's shown instead of your logo in their clients' OpenPanel. Combined with custom nameservers on their own domain, their clients see the reseller's brand.

For your own company branding, see [White-label OpenPanel](/docs/articles/hosting-business/white-label-openpanel/).

---

## Step 4: Automate It with Billing

Sell reseller packages through your billing system, so resellers are created and suspended automatically when they pay (or don't). OpenPanel integrates with [WHMCS](/docs/articles/extensions/openpanel-and-whmcs/), [FOSSBilling](/docs/articles/extensions/openpanel-and-fossbilling/), [Blesta](/docs/articles/extensions/openpanel-and-blesta/), [ClientExec](/docs/articles/extensions/openpanel-and-clientexec/) and [WISECP](/docs/articles/extensions/openpanel-and-wisecp/).

Resellers who run their own billing system can use the same modules to automate *their* client accounts.

---

## Managing Resellers

From **Accounts → Resellers** you can see each reseller's account usage (e.g. `12/20`), disk usage and plans, and:

- **Edit Plans & Limits** - upgrade or downgrade a reseller package,
- **Suspend / Unsuspend** - e.g. for unpaid invoices,
- **Rename**, **Change Password**, remove **2FA / Passkeys** if a reseller is locked out,
- **Delete** the reseller.

Resellers can enable their own **2FA and passkeys** - recommend it, since a reseller login controls many client sites.

---

## Pricing Reseller Hosting

Typical reseller packages are sold by **number of accounts and total disk**:

| Package | Accounts | Disk | Plans available |
|---|---|---|---|
| Reseller S | 10 | 50 GB | Starter |
| Reseller M | 25 | 150 GB | Starter, Business |
| Reseller L | 60 | 400 GB | Starter, Business, Pro |

- Price per account should still be lower than your retail price, so the reseller has margin.
- Because each client account has its own CPU/RAM limits from its plan, one reseller's busy client can't slow down the whole server.
- OpenPanel Enterprise is licensed **per server**, not per account - reseller accounts don't increase your license cost.

---

## Related

- [Start a web hosting business with OpenPanel](/docs/articles/hosting-business/start-web-hosting-business/)
- [Suggested hosting plan limits](/docs/articles/hosting-business/hosting-plan-templates/)
- [Resellers (OpenAdmin reference)](/docs/admin/accounts/resellers/)
