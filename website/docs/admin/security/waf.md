---
sidebar_position: 2
---

# Web Firewall (Coraza)

The **Security > Web Firewall (Coraza)** page allows you to manage CorazaWAF, a powerful Web Application Firewall integrated into OpenPanel.

Use this interface to enhance security by enabling protection against common web threats such as SQL injection, XSS, and other malicious behavior.

## Enable
Toggle the Web Application Firewall (WAF) on or off.

When enabled, CorazaWAF inspects incoming requests in real time and blocks suspicious activity according to the configured rules.

- **Enabled**: Executes the command `opencli waf enable`, activating [the WAF module](/docs/admin/settings/modules/#waf). This makes WAF manageable by users and automatically enables it for any new domains.
- **Disabled**: Executes the command `opencli waf disable -y`, deactivating [the WAF module](/docs/admin/settings/modules/#waf). This disables WAF management for users and turns off WAF for all existing and new domains.

![WAF page with the CorazaWAF Enable dropdown and the number of active rule sets](/img/openadmin-screenshots/security/waf-page.png#gh-light-mode-only)
![WAF page with the CorazaWAF Enable dropdown and the number of active rule sets](/img/openadmin-screenshots/security/waf-page_dark.png#gh-dark-mode-only)

## Rule Sets
Manage the rule sets that CorazaWAF uses to protect your applications.

**Active:** Displays the number of currently active rule sets (e.g., 21 / 23).

Click **Manage Rules** to enable or disable individual WAF rule sets according to your security needs.

The rule set table includes the following columns:

- **Name** – The name or identifier of the rule set.

- **Number of Rules** – Total number of rules contained within the set.

- **Status** – Indicates whether the rule set is currently enabled or disabled.

- **Actions** – **View** to inspect the rule set's contents, and a toggle button whose label switches between **Enable** and **Disable** depending on the rule set's current status.

![WAF rule sets page listing each rule set with its number of rules, status and the View and Disable actions](/img/openadmin-screenshots/security/waf-rules.png#gh-light-mode-only)
![WAF rule sets page listing each rule set with its number of rules, status and the View and Disable actions](/img/openadmin-screenshots/security/waf-rules_dark.png#gh-dark-mode-only)

Properly configuring WAF rules helps maintain a balance between strong protection and minimizing false positives.

### Bulk Actions

Tick the checkbox of one or more rule sets, or the checkbox in the table header to select all rule sets shown by the current search. A bar appears at the bottom of the page with the number selected, a **Clear** link and these actions:

![Two WAF rule sets selected with the bulk actions bar offering Enable and Disable](/img/openadmin-screenshots/security/waf-bulk.png#gh-light-mode-only)
![Two WAF rule sets selected with the bulk actions bar offering Enable and Disable](/img/openadmin-screenshots/security/waf-bulk_dark.png#gh-dark-mode-only)

| Action | What it does |
|---|---|
| **Enable** | Enables the selected rule sets. |
| **Disable** | Disables the selected rule sets. |

Restart Caddy afterwards to apply the changes.

Click an action, confirm it, and it runs on the selected rule sets one after another. When it's done the page reloads with a notice listing the rule sets it worked for, or which ones failed and why.
